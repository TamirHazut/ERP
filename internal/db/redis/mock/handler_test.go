package mock

import (
	"errors"
	"testing"

	db "erp.localhost/internal/db"
	"erp.localhost/internal/db/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that MockRedisHandler implements DBHandler interface
func TestMockRedisHandler_ImplementsDBHandler(t *testing.T) {
	var _ db.DBHandler = (*MockRedisHandler)(nil)
}

func TestNewMockRedisHandler(t *testing.T) {
	keyPrefix := redis.KeyPrefix("test_prefix")
	handler := NewMockRedisHandler(keyPrefix)

	require.NotNil(t, handler)
	assert.Equal(t, keyPrefix, handler.GetKeyPrefix())
}

func TestMockRedisHandler_Create(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		value     any
		mockFunc  func(key string, value any) (string, error)
		wantID    string
		wantErr   bool
	}{
		{
			name:  "successful create",
			key:   "test-key",
			value: "test-value",
			mockFunc: func(key string, value any) (string, error) {
				return "created-id", nil
			},
			wantID:  "created-id",
			wantErr: false,
		},
		{
			name:  "create with error",
			key:   "test-key",
			value: "test-value",
			mockFunc: func(key string, value any) (string, error) {
				return "", errors.New("create failed")
			},
			wantID:  "",
			wantErr: true,
		},
		{
			name:     "create with default behavior",
			key:      "test-key",
			value:    "test-value",
			mockFunc: nil,
			wantID:   "mock-key",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewMockRedisHandler("test_prefix")
			handler.CreateFunc = tc.mockFunc

			id, err := handler.Create(tc.key, tc.value)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, id)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantID, id)
			}
		})
	}
}

func TestMockRedisHandler_Find(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		filter    map[string]any
		mockFunc  func(key string, filter map[string]any) ([]any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "successful find",
			key:    "test-key",
			filter: nil,
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return []any{"value1", "value2"}, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "find with error",
			key:    "test-key",
			filter: nil,
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return nil, errors.New("find failed")
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "find with default behavior",
			key:       "test-key",
			filter:    nil,
			mockFunc:  nil,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewMockRedisHandler("test_prefix")
			handler.FindFunc = tc.mockFunc

			results, err := handler.Find(tc.key, tc.filter)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, results)
			} else {
				require.NoError(t, err)
				assert.Len(t, results, tc.wantCount)
			}
		})
	}
}

func TestMockRedisHandler_Update(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		filter    map[string]any
		value     any
		mockFunc  func(key string, filter map[string]any, value any) error
		wantErr   bool
	}{
		{
			name:   "successful update",
			key:    "test-key",
			filter: nil,
			value:  "updated-value",
			mockFunc: func(key string, filter map[string]any, value any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "update with error",
			key:    "test-key",
			filter: nil,
			value:  "updated-value",
			mockFunc: func(key string, filter map[string]any, value any) error {
				return errors.New("update failed")
			},
			wantErr: true,
		},
		{
			name:     "update with default behavior",
			key:      "test-key",
			filter:   nil,
			value:    "updated-value",
			mockFunc: nil,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewMockRedisHandler("test_prefix")
			handler.UpdateFunc = tc.mockFunc

			err := handler.Update(tc.key, tc.filter, tc.value)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMockRedisHandler_Delete(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		filter    map[string]any
		mockFunc  func(key string, filter map[string]any) error
		wantErr   bool
	}{
		{
			name:   "successful delete",
			key:    "test-key",
			filter: nil,
			mockFunc: func(key string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "delete with error",
			key:    "test-key",
			filter: nil,
			mockFunc: func(key string, filter map[string]any) error {
				return errors.New("delete failed")
			},
			wantErr: true,
		},
		{
			name:     "delete with default behavior",
			key:      "test-key",
			filter:   nil,
			mockFunc: nil,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewMockRedisHandler("test_prefix")
			handler.DeleteFunc = tc.mockFunc

			err := handler.Delete(tc.key, tc.filter)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMockRedisHandler_Close(t *testing.T) {
	testCases := []struct {
		name      string
		mockFunc  func() error
		wantErr   bool
	}{
		{
			name: "successful close",
			mockFunc: func() error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "close with error",
			mockFunc: func() error {
				return errors.New("close failed")
			},
			wantErr: true,
		},
		{
			name:     "close with default behavior",
			mockFunc: nil,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewMockRedisHandler("test_prefix")
			handler.CloseFunc = tc.mockFunc

			err := handler.Close()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

