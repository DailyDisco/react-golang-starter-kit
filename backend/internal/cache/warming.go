package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// WarmingConfig holds configuration for cache warming.
type WarmingConfig struct {
	// Enabled determines if cache warming runs on startup
	Enabled bool
	// Concurrency is the number of concurrent warming operations
	Concurrency int
	// Timeout is the maximum time for the entire warming operation
	Timeout time.Duration
}

// DefaultWarmingConfig returns sensible defaults for cache warming.
func DefaultWarmingConfig() *WarmingConfig {
	return &WarmingConfig{
		Enabled:     true,
		Concurrency: 5,
		Timeout:     30 * time.Second,
	}
}

// WarmingTask represents a cache warming task.
type WarmingTask struct {
	Name     string
	CacheKey string
	Loader   func(ctx context.Context) (interface{}, error)
	TTL      time.Duration
}

// WarmingResult represents the result of a warming task.
type WarmingResult struct {
	Name    string
	Success bool
	Error   error
	Latency time.Duration
}

// CacheWarmer handles preloading frequently-accessed data into cache.
type CacheWarmer struct {
	db     *gorm.DB
	config *WarmingConfig
	tasks  []WarmingTask
}

// NewCacheWarmer creates a new cache warmer.
func NewCacheWarmer(db *gorm.DB, config *WarmingConfig) *CacheWarmer {
	if config == nil {
		config = DefaultWarmingConfig()
	}
	return &CacheWarmer{
		db:     db,
		config: config,
		tasks:  make([]WarmingTask, 0),
	}
}

// RegisterTask adds a warming task to the warmer.
func (w *CacheWarmer) RegisterTask(task WarmingTask) {
	w.tasks = append(w.tasks, task)
}

// Warm executes all registered warming tasks.
func (w *CacheWarmer) Warm(ctx context.Context) []WarmingResult {
	if !w.config.Enabled || !IsAvailable() {
		log.Info().Msg("cache warming skipped (disabled or cache unavailable)")
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, w.config.Timeout)
	defer cancel()

	log.Info().
		Int("tasks", len(w.tasks)).
		Int("concurrency", w.config.Concurrency).
		Msg("starting cache warming")

	start := time.Now()
	results := make([]WarmingResult, len(w.tasks))

	// Use semaphore for concurrency control
	sem := make(chan struct{}, w.config.Concurrency)
	var wg sync.WaitGroup

	for i, task := range w.tasks {
		wg.Add(1)
		go func(idx int, t WarmingTask) {
			defer wg.Done()

			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			result := w.executeTask(ctx, t)
			results[idx] = result
		}(i, task)
	}

	wg.Wait()

	// Log summary
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	log.Info().
		Int("total", len(w.tasks)).
		Int("success", successCount).
		Int("failed", len(w.tasks)-successCount).
		Dur("duration", time.Since(start)).
		Msg("cache warming completed")

	return results
}

// executeTask runs a single warming task.
func (w *CacheWarmer) executeTask(ctx context.Context, task WarmingTask) WarmingResult {
	start := time.Now()
	result := WarmingResult{
		Name: task.Name,
	}

	// Load the data
	data, err := task.Loader(ctx)
	if err != nil {
		result.Error = err
		result.Latency = time.Since(start)
		log.Warn().
			Err(err).
			Str("task", task.Name).
			Msg("cache warming task failed to load data")
		return result
	}

	// Store in cache
	if err := SetJSON(ctx, task.CacheKey, data, task.TTL); err != nil {
		result.Error = err
		result.Latency = time.Since(start)
		log.Warn().
			Err(err).
			Str("task", task.Name).
			Str("key", task.CacheKey).
			Msg("cache warming task failed to store data")
		return result
	}

	result.Success = true
	result.Latency = time.Since(start)

	log.Debug().
		Str("task", task.Name).
		Str("key", task.CacheKey).
		Dur("latency", result.Latency).
		Msg("cache warming task completed")

	return result
}

// RegisterDefaultTasks registers common warming tasks for a typical application.
func (w *CacheWarmer) RegisterDefaultTasks(db *gorm.DB) {
	// Feature flags - checked on every request
	w.RegisterTask(WarmingTask{
		Name:     "feature_flags",
		CacheKey: "feature_flags:all",
		TTL:      5 * time.Minute,
		Loader: func(ctx context.Context) (interface{}, error) {
			type FeatureFlag struct {
				ID          uint   `json:"id"`
				Key         string `json:"key"`
				Enabled     bool   `json:"enabled"`
				Description string `json:"description"`
			}
			var flags []FeatureFlag
			if err := db.WithContext(ctx).
				Table("feature_flags").
				Select("id, key, enabled, description").
				Where("enabled = ?", true).
				Find(&flags).Error; err != nil {
				return nil, err
			}
			return flags, nil
		},
	})

	// Site settings - global configuration
	w.RegisterTask(WarmingTask{
		Name:     "site_settings",
		CacheKey: "settings:site",
		TTL:      10 * time.Minute,
		Loader: func(ctx context.Context) (interface{}, error) {
			settings := make(map[string]interface{})
			type Setting struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			var rows []Setting
			if err := db.WithContext(ctx).
				Table("settings").
				Select("key, value").
				Where("category = ?", "site").
				Find(&rows).Error; err != nil {
				return nil, err
			}
			for _, row := range rows {
				var value interface{}
				if err := json.Unmarshal([]byte(row.Value), &value); err != nil {
					settings[row.Key] = row.Value
				} else {
					settings[row.Key] = value
				}
			}
			return settings, nil
		},
	})

	// Active announcements
	w.RegisterTask(WarmingTask{
		Name:     "active_announcements",
		CacheKey: "announcements:active",
		TTL:      5 * time.Minute,
		Loader: func(ctx context.Context) (interface{}, error) {
			type Announcement struct {
				ID        uint      `json:"id"`
				Title     string    `json:"title"`
				Content   string    `json:"content"`
				Type      string    `json:"type"`
				StartDate time.Time `json:"start_date"`
				EndDate   time.Time `json:"end_date"`
			}
			var announcements []Announcement
			now := time.Now()
			if err := db.WithContext(ctx).
				Table("announcements").
				Select("id, title, content, type, start_date, end_date").
				Where("is_active = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)", true, now, now).
				Order("created_at DESC").
				Limit(10).
				Find(&announcements).Error; err != nil {
				return nil, err
			}
			return announcements, nil
		},
	})

	// Recently active users (for quick profile lookups)
	w.RegisterTask(WarmingTask{
		Name:     "recent_active_users",
		CacheKey: "users:recent_active",
		TTL:      2 * time.Minute,
		Loader: func(ctx context.Context) (interface{}, error) {
			type UserSummary struct {
				ID        uint   `json:"id"`
				Name      string `json:"name"`
				Email     string `json:"email"`
				Role      string `json:"role"`
				AvatarURL string `json:"avatar_url"`
			}
			var users []UserSummary
			// Get users who logged in within the last hour
			cutoff := time.Now().Add(-1 * time.Hour)
			if err := db.WithContext(ctx).
				Table("users").
				Select("id, name, email, role, avatar_url").
				Where("is_active = ? AND deleted_at IS NULL AND updated_at > ?", true, cutoff).
				Order("updated_at DESC").
				Limit(100).
				Find(&users).Error; err != nil {
				return nil, err
			}
			return users, nil
		},
	})

	log.Info().Int("tasks", len(w.tasks)).Msg("registered default cache warming tasks")
}

// WarmUserCache warms the cache for a specific user.
func WarmUserCache(ctx context.Context, db *gorm.DB, userID uint) error {
	if !IsAvailable() {
		return nil
	}

	type UserData struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		EmailVerified bool   `json:"email_verified"`
		IsActive      bool   `json:"is_active"`
		AvatarURL     string `json:"avatar_url"`
	}

	var user UserData
	if err := db.WithContext(ctx).
		Table("users").
		Select("id, name, email, role, email_verified, is_active, avatar_url").
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&user).Error; err != nil {
		return fmt.Errorf("failed to load user for cache warming: %w", err)
	}

	cacheKey := fmt.Sprintf("user:%d", userID)
	if err := SetJSON(ctx, cacheKey, user, 5*time.Minute); err != nil {
		return fmt.Errorf("failed to cache user: %w", err)
	}

	log.Debug().Uint("user_id", userID).Msg("warmed user cache")
	return nil
}

// WarmFeatureFlagsCache warms the feature flags cache.
func WarmFeatureFlagsCache(ctx context.Context, db *gorm.DB) error {
	if !IsAvailable() {
		return nil
	}

	type FeatureFlag struct {
		Key     string `json:"key"`
		Enabled bool   `json:"enabled"`
	}

	var flags []FeatureFlag
	if err := db.WithContext(ctx).
		Table("feature_flags").
		Select("key, enabled").
		Find(&flags).Error; err != nil {
		return fmt.Errorf("failed to load feature flags: %w", err)
	}

	// Store as a map for fast lookups
	flagMap := make(map[string]bool)
	for _, f := range flags {
		flagMap[f.Key] = f.Enabled
	}

	if err := SetJSON(ctx, "feature_flags:map", flagMap, 5*time.Minute); err != nil {
		return fmt.Errorf("failed to cache feature flags: %w", err)
	}

	log.Debug().Int("count", len(flags)).Msg("warmed feature flags cache")
	return nil
}
