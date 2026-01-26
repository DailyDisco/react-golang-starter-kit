package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Concurrency tests for SessionService
// These tests verify thread-safety and race condition handling

// --- Concurrent Session Creation ---

func TestSessionService_ConcurrentSessionCreation(t *testing.T) {
	t.Run("10 goroutines create sessions simultaneously", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		numGoroutines := 10
		userID := uint(1)

		var wg sync.WaitGroup
		var successCount int64
		var errorCount int64

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				req := httptest.NewRequest(http.MethodPost, "/login", nil)
				req.Header.Set("User-Agent", fmt.Sprintf("TestAgent-%d/1.0", i))
				req.RemoteAddr = fmt.Sprintf("192.168.1.%d:8080", i)

				token := fmt.Sprintf("token-%d-%d", userID, i)
				_, err := svc.CreateSessionWithContext(ctx, userID, token, req)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}(i)
		}

		wg.Wait()

		assert.Equal(t, int64(numGoroutines), successCount, "all sessions should be created successfully")
		assert.Equal(t, int64(0), errorCount, "no errors should occur")

		// Verify all sessions were created
		sessions, err := svc.GetUserSessionsWithContext(ctx, userID, "")
		require.NoError(t, err)
		assert.Equal(t, numGoroutines, len(sessions), "all sessions should exist")
	})

	t.Run("different users create sessions concurrently", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		numUsers := 5
		sessionsPerUser := 3

		var wg sync.WaitGroup
		sessionIDs := make(map[uint][]uint)
		var mu sync.Mutex

		for userID := 1; userID <= numUsers; userID++ {
			for j := 0; j < sessionsPerUser; j++ {
				wg.Add(1)
				go func(uid uint, idx int) {
					defer wg.Done()

					req := httptest.NewRequest(http.MethodPost, "/login", nil)
					req.Header.Set("User-Agent", fmt.Sprintf("User%d-Session%d/1.0", uid, idx))
					req.RemoteAddr = "192.168.1.1:8080"

					token := fmt.Sprintf("token-%d-%d", uid, idx)
					session, err := svc.CreateSessionWithContext(ctx, uid, token, req)
					if err == nil && session != nil {
						mu.Lock()
						sessionIDs[uid] = append(sessionIDs[uid], session.ID)
						mu.Unlock()
					}
				}(uint(userID), j)
			}
		}

		wg.Wait()

		// Verify each user has the correct number of sessions
		for userID := 1; userID <= numUsers; userID++ {
			sessions, err := svc.GetUserSessionsWithContext(ctx, uint(userID), "")
			require.NoError(t, err)
			assert.Equal(t, sessionsPerUser, len(sessions), "user %d should have %d sessions", userID, sessionsPerUser)
		}
	})
}

// --- Concurrent Session Retrieval ---

func TestSessionService_ConcurrentSessionRetrieval(t *testing.T) {
	t.Run("concurrent reads don't block each other", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		userID := uint(1)

		// Create some sessions first
		for i := 0; i < 5; i++ {
			sessionRepo.AddSession(models.UserSession{
				ID:               uint(i + 1),
				UserID:           userID,
				SessionTokenHash: fmt.Sprintf("hash-%d", i),
				ExpiresAt:        time.Now().Add(time.Hour),
			})
		}

		// Concurrent reads
		numReaders := 20
		var wg sync.WaitGroup
		results := make([]int, numReaders)

		start := make(chan struct{})

		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				<-start // Wait for all goroutines to be ready

				sessions, err := svc.GetUserSessionsWithContext(ctx, userID, "")
				if err == nil {
					results[idx] = len(sessions)
				}
			}(i)
		}

		// Start all goroutines at once
		close(start)
		wg.Wait()

		// All should read the same number
		for i, count := range results {
			assert.Equal(t, 5, count, "reader %d should see 5 sessions", i)
		}
	})
}

// --- Concurrent Revocation ---

func TestSessionService_ConcurrentRevokeSingleSession(t *testing.T) {
	t.Run("multiple goroutines try to revoke same session", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		userID := uint(1)
		sessionID := uint(1)

		// Create a session
		sessionRepo.AddSession(models.UserSession{
			ID:               sessionID,
			UserID:           userID,
			SessionTokenHash: "hash-to-revoke",
			ExpiresAt:        time.Now().Add(time.Hour),
		})

		numGoroutines := 10
		var wg sync.WaitGroup
		var successCount int64
		var notFoundCount int64

		start := make(chan struct{})

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start

				err := svc.RevokeSessionWithContext(ctx, userID, sessionID)
				if err == nil {
					atomic.AddInt64(&successCount, 1)
				} else if err == ErrSessionNotFound {
					atomic.AddInt64(&notFoundCount, 1)
				}
			}()
		}

		close(start)
		wg.Wait()

		// Only one should succeed, rest should get not found
		assert.Equal(t, int64(1), successCount, "exactly one revocation should succeed")
		assert.Equal(t, int64(numGoroutines-1), notFoundCount, "others should get not found")
	})
}

func TestSessionService_ConcurrentRevokeAllSessions(t *testing.T) {
	t.Run("revoke all while creating new sessions", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		userID := uint(1)

		// Create initial sessions
		for i := 0; i < 5; i++ {
			sessionRepo.AddSession(models.UserSession{
				ID:               uint(i + 1),
				UserID:           userID,
				SessionTokenHash: fmt.Sprintf("old-hash-%d", i),
				ExpiresAt:        time.Now().Add(time.Hour),
			})
		}

		var wg sync.WaitGroup
		keepHash := "new-current-hash"

		// Goroutine 1: Revoke all sessions except current
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := svc.RevokeAllSessionsWithContext(ctx, userID, keepHash)
			assert.NoError(t, err)
		}()

		// Goroutine 2-4: Create new sessions
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				req := httptest.NewRequest(http.MethodPost, "/login", nil)
				req.Header.Set("User-Agent", "TestAgent/1.0")
				req.RemoteAddr = "192.168.1.1:8080"

				_, err := svc.CreateSessionWithContext(ctx, userID, fmt.Sprintf("new-token-%d", idx), req)
				// May or may not succeed depending on timing
				_ = err
			}(i)
		}

		wg.Wait()

		// No panics or deadlocks occurred - that's the main assertion
	})
}

func TestSessionService_RevokeDuringActiveRequest(t *testing.T) {
	t.Run("session revoked while being used", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		userID := uint(1)
		tokenHash := "active-session-hash"

		// Create session
		sessionRepo.AddSession(models.UserSession{
			ID:               1,
			UserID:           userID,
			SessionTokenHash: tokenHash,
			ExpiresAt:        time.Now().Add(time.Hour),
		})

		var wg sync.WaitGroup

		// Goroutine 1: Continuously check session (simulating active use)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				sessions, _ := svc.GetUserSessionsWithContext(ctx, userID, tokenHash)
				// Sessions may be 0 or 1 depending on timing - both are valid
				_ = sessions
			}
		}()

		// Goroutine 2: Revoke the session
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := svc.RevokeSessionWithContext(ctx, userID, 1)
			// May succeed or fail with not found if already revoked
			_ = err
		}()

		wg.Wait()

		// Main assertion: no panics or races occurred
	})
}

// --- Cleanup Race Conditions ---

func TestSessionService_ConcurrentCleanup(t *testing.T) {
	t.Run("multiple cleanup calls don't interfere", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()

		// Create a mix of expired and active sessions
		for i := 0; i < 20; i++ {
			expiry := time.Now().Add(time.Hour)
			if i%2 == 0 {
				expiry = time.Now().Add(-time.Hour) // expired
			}
			sessionRepo.AddSession(models.UserSession{
				ID:               uint(i + 1),
				UserID:           uint(i%5 + 1),
				SessionTokenHash: fmt.Sprintf("hash-%d", i),
				ExpiresAt:        expiry,
			})
		}

		numCleaners := 5
		var wg sync.WaitGroup
		var totalDeleted int64

		start := make(chan struct{})

		for i := 0; i < numCleaners; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start

				deleted, err := svc.CleanupExpiredSessionsWithContext(ctx)
				if err == nil {
					atomic.AddInt64(&totalDeleted, deleted)
				}
			}()
		}

		close(start)
		wg.Wait()

		// Total deleted should be 10 (half are expired)
		// But due to concurrent cleanup, might be counted multiple times
		// or less if cleanup happens sequentially
		assert.GreaterOrEqual(t, totalDeleted, int64(10), "at least 10 expired sessions should be cleaned up")
	})
}

func TestSessionService_CleanupDuringSessionUse(t *testing.T) {
	t.Run("cleanup runs while sessions are being accessed", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()

		// Create sessions with varying expiry times
		for i := 0; i < 10; i++ {
			expiry := time.Now().Add(time.Hour)
			if i < 3 {
				expiry = time.Now().Add(-time.Minute) // expired
			}
			sessionRepo.AddSession(models.UserSession{
				ID:               uint(i + 1),
				UserID:           uint(1),
				SessionTokenHash: fmt.Sprintf("hash-%d", i),
				ExpiresAt:        expiry,
			})
		}

		var wg sync.WaitGroup
		stopReading := make(chan struct{})

		// Continuous readers
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stopReading:
						return
					default:
						sessions, _ := svc.GetUserSessionsWithContext(ctx, 1, "")
						// Active sessions should still be readable
						_ = sessions
					}
				}
			}()
		}

		// Run cleanup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.CleanupExpiredSessionsWithContext(ctx)
			assert.NoError(t, err)
		}()

		// Let it run for a bit
		time.Sleep(10 * time.Millisecond)
		close(stopReading)

		wg.Wait()

		// Verify active sessions are still accessible
		sessions, err := svc.GetUserSessionsWithContext(ctx, 1, "")
		require.NoError(t, err)
		assert.Equal(t, 7, len(sessions), "7 active sessions should remain")
	})
}

// --- Update Last Active Concurrency ---

func TestSessionService_ConcurrentUpdateLastActive(t *testing.T) {
	t.Run("concurrent updates to same session", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		tokenHash := "test-hash"

		sessionRepo.AddSession(models.UserSession{
			ID:               1,
			UserID:           1,
			SessionTokenHash: tokenHash,
			LastActiveAt:     time.Now().Add(-time.Hour),
			ExpiresAt:        time.Now().Add(time.Hour),
		})

		numUpdates := 50
		var wg sync.WaitGroup
		var successCount int64

		for i := 0; i < numUpdates; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := svc.UpdateLastActiveWithContext(ctx, tokenHash)
				if err == nil {
					atomic.AddInt64(&successCount, 1)
				}
			}()
		}

		wg.Wait()

		assert.Equal(t, int64(numUpdates), successCount, "all updates should succeed")
	})
}

// --- Login History Concurrency ---

func TestSessionService_ConcurrentLoginHistoryRecording(t *testing.T) {
	t.Run("concurrent login attempts recorded correctly", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		numAttempts := 20

		var wg sync.WaitGroup

		for i := 0; i < numAttempts; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				req := httptest.NewRequest(http.MethodPost, "/login", nil)
				req.Header.Set("User-Agent", fmt.Sprintf("TestAgent-%d/1.0", idx))
				req.RemoteAddr = fmt.Sprintf("192.168.1.%d:8080", idx%256)

				success := idx%2 == 0
				failureReason := ""
				if !success {
					failureReason = "invalid password"
				}

				err := svc.RecordLoginAttemptWithContext(
					ctx,
					uint(idx%5+1), // 5 different users
					success,
					failureReason,
					models.AuthMethodPassword,
					req,
					nil,
				)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()

		assert.Equal(t, numAttempts, historyRepo.CreateCalls, "all login attempts should be recorded")
	})
}

// --- Mixed Operations Stress Test ---

func TestSessionService_MixedOperationsStress(t *testing.T) {
	t.Run("all operations running concurrently", func(t *testing.T) {
		sessionRepo := mocks.NewMockSessionRepository()
		historyRepo := mocks.NewMockLoginHistoryRepository()
		svc := NewSessionServiceWithRepo(sessionRepo, historyRepo)

		ctx := context.Background()
		duration := 100 * time.Millisecond
		stop := make(chan struct{})
		var wg sync.WaitGroup

		// Seed some initial sessions
		for i := 0; i < 5; i++ {
			sessionRepo.AddSession(models.UserSession{
				ID:               uint(i + 1),
				UserID:           uint(i%3 + 1),
				SessionTokenHash: fmt.Sprintf("initial-hash-%d", i),
				ExpiresAt:        time.Now().Add(time.Hour),
			})
		}

		// Creator goroutines
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				counter := 0
				for {
					select {
					case <-stop:
						return
					default:
						req := httptest.NewRequest(http.MethodPost, "/login", nil)
						req.Header.Set("User-Agent", "StressTest/1.0")
						req.RemoteAddr = "10.0.0.1:8080"

						svc.CreateSessionWithContext(ctx, uint(idx%5+1), fmt.Sprintf("stress-%d-%d", idx, counter), req)
						counter++
					}
				}
			}(i)
		}

		// Reader goroutines
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return
					default:
						svc.GetUserSessionsWithContext(ctx, uint(idx%5+1), "")
					}
				}
			}(i)
		}

		// Revoker goroutines
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				sessionID := uint(1)
				for {
					select {
					case <-stop:
						return
					default:
						svc.RevokeSessionWithContext(ctx, uint(idx%3+1), sessionID)
						sessionID++
					}
				}
			}(i)
		}

		// Cleanup goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					svc.CleanupExpiredSessionsWithContext(ctx)
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()

		// Run for duration
		time.Sleep(duration)
		close(stop)
		wg.Wait()

		// Main assertion: no panics, deadlocks, or data races occurred
		t.Log("Mixed operations stress test completed without panics or deadlocks")
	})
}
