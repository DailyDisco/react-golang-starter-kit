package cache

import (
	"fmt"
	"react-golang-starter/internal/models"
	"time"
)

// Service represents a caching service for user data
type Service struct {
	cache *Cache
}

// NewService creates a new cache service
func NewService(cache *Cache) *Service {
	return &Service{
		cache: cache,
	}
}

// IsAvailable returns true if caching is available
func (s *Service) IsAvailable() bool {
	return s.cache != nil
}

// Cache keys constants
const (
	UserKeyPrefix = "user:"
	UserListKey   = "users:list"
	UserCountKey  = "users:count"
	UserKeyTTL    = 15 * time.Minute // 15 minutes for individual users
	UserListTTL   = 10 * time.Minute // 10 minutes for user lists
	UserCountTTL  = 30 * time.Minute // 30 minutes for user counts
)

// UserCacheKey generates a cache key for a specific user
func (s *Service) UserCacheKey(userID uint) string {
	return fmt.Sprintf("%s%d", UserKeyPrefix, userID)
}

// UserListCacheKey generates a cache key for user list with pagination
func (s *Service) UserListCacheKey(page, limit int) string {
	return fmt.Sprintf("%s:page:%d:limit:%d", UserListKey, page, limit)
}

// SetUser caches a single user
func (s *Service) SetUser(user *models.UserResponse) error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	key := s.UserCacheKey(user.ID)
	return s.cache.Set(key, user, UserKeyTTL)
}

// GetUser retrieves a user from cache
func (s *Service) GetUser(userID uint) (*models.UserResponse, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("cache not available")
	}
	key := s.UserCacheKey(userID)
	var user models.UserResponse
	err := s.cache.Get(key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser removes a user from cache
func (s *Service) DeleteUser(userID uint) error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	key := s.UserCacheKey(userID)
	return s.cache.Delete(key)
}

// SetUserList caches a paginated user list
func (s *Service) SetUserList(page, limit int, usersResponse *models.UsersResponse) error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	key := s.UserListCacheKey(page, limit)
	return s.cache.Set(key, usersResponse, UserListTTL)
}

// GetUserList retrieves a paginated user list from cache
func (s *Service) GetUserList(page, limit int) (*models.UsersResponse, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("cache not available")
	}
	key := s.UserListCacheKey(page, limit)
	var usersResponse models.UsersResponse
	err := s.cache.Get(key, &usersResponse)
	if err != nil {
		return nil, err
	}
	return &usersResponse, nil
}

// SetUserCount caches the total user count
func (s *Service) SetUserCount(count int) error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	return s.cache.Set(UserCountKey, count, UserCountTTL)
}

// GetUserCount retrieves the user count from cache
func (s *Service) GetUserCount() (int, error) {
	if !s.IsAvailable() {
		return 0, fmt.Errorf("cache not available")
	}
	var count int
	err := s.cache.Get(UserCountKey, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// InvalidateUserList invalidates all cached user lists
func (s *Service) InvalidateUserList() error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	pattern := UserListKey + ":*"
	return s.cache.DeletePattern(pattern)
}

// InvalidateAllUsers invalidates all user-related cache
func (s *Service) InvalidateAllUsers() error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	// Invalidate individual users
	userPattern := UserKeyPrefix + "*"
	if err := s.cache.DeletePattern(userPattern); err != nil {
		return fmt.Errorf("failed to invalidate user cache: %w", err)
	}

	// Invalidate user lists
	if err := s.InvalidateUserList(); err != nil {
		return fmt.Errorf("failed to invalidate user list cache: %w", err)
	}

	// Invalidate user count
	if err := s.cache.Delete(UserCountKey); err != nil {
		return fmt.Errorf("failed to invalidate user count cache: %w", err)
	}

	return nil
}

// InvalidateUser invalidates cache for a specific user and related lists
func (s *Service) InvalidateUser(userID uint) error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	// Delete the specific user cache
	if err := s.DeleteUser(userID); err != nil {
		return fmt.Errorf("failed to delete user cache: %w", err)
	}

	// Invalidate user lists (since they might contain this user)
	if err := s.InvalidateUserList(); err != nil {
		return fmt.Errorf("failed to invalidate user list cache: %w", err)
	}

	// Invalidate user count (since count might have changed)
	if err := s.cache.Delete(UserCountKey); err != nil {
		return fmt.Errorf("failed to invalidate user count cache: %w", err)
	}

	return nil
}

// UserExists checks if a user exists in cache
func (s *Service) UserExists(userID uint) bool {
	if !s.IsAvailable() {
		return false
	}
	key := s.UserCacheKey(userID)
	return s.cache.Exists(key)
}

// UserListExists checks if a user list exists in cache
func (s *Service) UserListExists(page, limit int) bool {
	if !s.IsAvailable() {
		return false
	}
	key := s.UserListCacheKey(page, limit)
	return s.cache.Exists(key)
}

// UserCountExists checks if user count exists in cache
func (s *Service) UserCountExists() bool {
	if !s.IsAvailable() {
		return false
	}
	return s.cache.Exists(UserCountKey)
}

// RefreshUserTTL refreshes the TTL for a specific user
func (s *Service) RefreshUserTTL(userID uint) error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	key := s.UserCacheKey(userID)
	return s.cache.SetTTL(key, UserKeyTTL)
}

// RefreshUserListTTL refreshes the TTL for a user list
func (s *Service) RefreshUserListTTL(page, limit int) error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	key := s.UserListCacheKey(page, limit)
	return s.cache.SetTTL(key, UserListTTL)
}

// RefreshUserCountTTL refreshes the TTL for user count
func (s *Service) RefreshUserCountTTL() error {
	if !s.IsAvailable() {
		return nil // Silently skip if cache is not available
	}
	return s.cache.SetTTL(UserCountKey, UserCountTTL)
}
