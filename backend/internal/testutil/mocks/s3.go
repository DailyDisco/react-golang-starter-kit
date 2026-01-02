package mocks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
)

// MockS3Client provides an in-memory S3 mock.
type MockS3Client struct {
	mu      sync.RWMutex
	buckets map[string]map[string][]byte

	// Configurable errors
	PutObjectError    error
	GetObjectError    error
	DeleteObjectError error
	HeadObjectError   error

	// Track calls for assertions
	PutObjectCalls    int
	GetObjectCalls    int
	DeleteObjectCalls int
	HeadObjectCalls   int
}

// NewMockS3Client creates a new mock S3 client.
func NewMockS3Client() *MockS3Client {
	return &MockS3Client{
		buckets: make(map[string]map[string][]byte),
	}
}

// CreateBucket creates a mock bucket.
func (m *MockS3Client) CreateBucket(ctx context.Context, bucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucket]; !exists {
		m.buckets[bucket] = make(map[string][]byte)
	}
	return nil
}

// PutObject stores an object in the mock S3.
func (m *MockS3Client) PutObject(ctx context.Context, bucket, key string, body io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PutObjectCalls++

	if m.PutObjectError != nil {
		return m.PutObjectError
	}

	if _, exists := m.buckets[bucket]; !exists {
		m.buckets[bucket] = make(map[string][]byte)
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	m.buckets[bucket][key] = data
	return nil
}

// GetObject retrieves an object from the mock S3.
func (m *MockS3Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.GetObjectCalls++

	if m.GetObjectError != nil {
		return nil, m.GetObjectError
	}

	if b, exists := m.buckets[bucket]; exists {
		if data, ok := b[key]; ok {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
	}

	return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
}

// DeleteObject removes an object from the mock S3.
func (m *MockS3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteObjectCalls++

	if m.DeleteObjectError != nil {
		return m.DeleteObjectError
	}

	if b, exists := m.buckets[bucket]; exists {
		delete(b, key)
	}

	return nil
}

// HeadObject checks if an object exists in the mock S3.
func (m *MockS3Client) HeadObject(ctx context.Context, bucket, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.HeadObjectCalls++

	if m.HeadObjectError != nil {
		return 0, m.HeadObjectError
	}

	if b, exists := m.buckets[bucket]; exists {
		if data, ok := b[key]; ok {
			return int64(len(data)), nil
		}
	}

	return 0, fmt.Errorf("object not found: %s/%s", bucket, key)
}

// ObjectExists checks if an object exists in the mock S3.
func (m *MockS3Client) ObjectExists(ctx context.Context, bucket, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if b, exists := m.buckets[bucket]; exists {
		_, ok := b[key]
		return ok
	}
	return false
}

// ListObjects lists objects in a bucket with optional prefix.
func (m *MockS3Client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if b, exists := m.buckets[bucket]; exists {
		var keys []string
		for key := range b {
			if prefix == "" || strings.HasPrefix(key, prefix) {
				keys = append(keys, key)
			}
		}
		return keys, nil
	}

	return nil, fmt.Errorf("bucket not found: %s", bucket)
}

// GetObjectSize returns the size of an object.
func (m *MockS3Client) GetObjectSize(bucket, key string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if b, exists := m.buckets[bucket]; exists {
		if data, ok := b[key]; ok {
			return int64(len(data))
		}
	}
	return 0
}

// GetObjectData returns the raw data of an object.
func (m *MockS3Client) GetObjectData(bucket, key string) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if b, exists := m.buckets[bucket]; exists {
		if data, ok := b[key]; ok {
			result := make([]byte, len(data))
			copy(result, data)
			return result
		}
	}
	return nil
}

// BucketExists checks if a bucket exists.
func (m *MockS3Client) BucketExists(bucket string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.buckets[bucket]
	return exists
}

// ObjectCount returns the number of objects in a bucket.
func (m *MockS3Client) ObjectCount(bucket string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if b, exists := m.buckets[bucket]; exists {
		return len(b)
	}
	return 0
}

// Reset clears all mock data and resets counters.
func (m *MockS3Client) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buckets = make(map[string]map[string][]byte)
	m.PutObjectError = nil
	m.GetObjectError = nil
	m.DeleteObjectError = nil
	m.HeadObjectError = nil
	m.PutObjectCalls = 0
	m.GetObjectCalls = 0
	m.DeleteObjectCalls = 0
	m.HeadObjectCalls = 0
}
