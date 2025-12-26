package mock

import (
	"erp.localhost/internal/db/redis"
)

// MockRedisHandler is a mock implementation of RedisHandler for testing
type MockRedisHandler struct {
	keyPrefix  redis.KeyPrefix
	CreateFunc func(key string, value any) (string, error)
	FindFunc   func(key string, filter map[string]any) ([]any, error)
	UpdateFunc func(key string, filter map[string]any, value any) error
	DeleteFunc func(key string, filter map[string]any) error
	CloseFunc  func() error
}

// NewMockRedisHandler creates a new mock RedisHandler
func NewMockRedisHandler(keyPrefix redis.KeyPrefix) *MockRedisHandler {
	return &MockRedisHandler{
		keyPrefix: keyPrefix,
	}
}

// Create implements the DBHandler interface
func (m *MockRedisHandler) Create(key string, value any) (string, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(key, value)
	}
	return "mock-key", nil
}

// Find implements the DBHandler interface
func (m *MockRedisHandler) Find(key string, filter map[string]any) ([]any, error) {
	if m.FindFunc != nil {
		return m.FindFunc(key, filter)
	}
	return []any{}, nil
}

// Update implements the DBHandler interface
func (m *MockRedisHandler) Update(key string, filter map[string]any, value any) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(key, filter, value)
	}
	return nil
}

// Delete implements the DBHandler interface
func (m *MockRedisHandler) Delete(key string, filter map[string]any) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key, filter)
	}
	return nil
}

// Close mocks the Close method
func (m *MockRedisHandler) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// GetKeyPrefix returns the key prefix (for testing purposes)
func (m *MockRedisHandler) GetKeyPrefix() redis.KeyPrefix {
	return m.keyPrefix
}

