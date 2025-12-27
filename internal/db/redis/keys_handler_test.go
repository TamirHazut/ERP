package redis

import (
	"encoding/json"
	"errors"
	"testing"

	db "erp.localhost/internal/db"
	"erp.localhost/internal/db/mock"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel is a simple test model for key handler tests
type TestModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// newKeyHandlerWithMock creates a KeyHandler with a mock DBHandler for testing (internal use)
func newKeyHandlerWithMock[T any](mockHandler db.DBHandler, logger *logging.Logger) *KeyHandler[T] {
	return NewKeyHandlerWithMockForTest[T](mockHandler, logger)
}

func TestNewKeyHandler(t *testing.T) {
	// Note: This test requires a running Redis instance
	// If Redis is not available, it will fail
	// For unit testing, use newKeyHandlerWithMock instead
	// This test is commented out as it requires Redis
	// logger := logging.NewLogger(logging.ModuleDB)
	// handler := NewKeyHandler[TestModel]("test_prefix", logger)
	// if handler == nil {
	// 	t.Skip("Redis not available, skipping test")
	// }
	// assert.NotNil(t, handler)
	t.Skip("Requires Redis instance - skipping integration test")
}

func TestKeyHandler_Set(t *testing.T) {
	testCases := []struct {
		tenantID string
		name     string
		key      string
		value    TestModel
		mockFunc func(key string, data any) (string, error)
		wantErr  bool
	}{
		{
			tenantID: "1",
			name:     "successful set",
			key:      "test-key",
			value:    TestModel{ID: "1", Name: "test"},
			mockFunc: func(key string, data any) (string, error) {
				return "ok", nil
			},
			wantErr: false,
		},
		{
			name:     "set with database error",
			tenantID: "1",
			key:      "test-key",
			value:    TestModel{ID: "1", Name: "test"},
			mockFunc: func(key string, data any) (string, error) {
				return "", errors.New("database connection failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				CreateFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleDB)
			handler := newKeyHandlerWithMock[TestModel](mockHandler, logger)

			err := handler.Set(tc.tenantID, tc.key, tc.value)
			if tc.wantErr {
				require.Error(t, err)
				appErr, ok := erp_errors.AsAppError(err)
				require.True(t, ok, "Expected AppError")
				assert.Equal(t, erp_errors.CategoryInternal, appErr.Category)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeyHandler_GetOne(t *testing.T) {
	testModel := TestModel{ID: "1", Name: "test"}
	jsonData, _ := json.Marshal(testModel)

	testCases := []struct {
		name      string
		tenantID  string
		key       string
		mockFunc  func(key string, filter map[string]any) (any, error)
		wantModel TestModel
		wantErr   bool
	}{
		{
			name:     "successful get one",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return string(jsonData), nil
			},
			wantModel: testModel,
			wantErr:   false,
		},
		{
			name:     "get one with error",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return nil, errors.New("get one failed")
			},
			wantErr: true,
		},
		{
			name:     "get one with error incompatible type",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return "invalid json", nil
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindOneFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleDB)
			handler := newKeyHandlerWithMock[TestModel](mockHandler, logger)
			result, err := handler.GetOne(tc.tenantID, tc.key)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.wantModel.ID, result.ID)
				assert.Equal(t, tc.wantModel.Name, result.Name)
			}
		})
	}
}

func TestKeyHandler_GetAll(t *testing.T) {
	testModel := TestModel{ID: "1", Name: "test"}
	jsonData, _ := json.Marshal(testModel)

	testCases := []struct {
		name      string
		tenantID  string
		key       string
		mockFunc  func(key string, filter map[string]any) ([]any, error)
		wantModel *TestModel
		wantErr   bool
	}{
		{
			name:     "successful get",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return []any{string(jsonData)}, nil
			},
			wantModel: &testModel,
			wantErr:   false,
		},
		{
			name:     "key not found",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantModel: nil,
			wantErr:   false,
		},
		{
			name:     "database error",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return nil, errors.New("database query failed")
			},
			wantModel: nil,
			wantErr:   true,
		},
		{
			name:     "invalid JSON",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return []any{"invalid json"}, nil
			},
			wantModel: nil,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindAllFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleDB)
			handler := newKeyHandlerWithMock[TestModel](mockHandler, logger)

			result, err := handler.GetAll(tc.tenantID, tc.key)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				if tc.name == "key not found" {
					appErr, ok := erp_errors.AsAppError(err)
					require.True(t, ok)
					assert.Equal(t, erp_errors.CategoryNotFound, appErr.Category)
				}
			} else {
				require.NoError(t, err)
				if tc.wantModel == nil {
					require.Empty(t, result)
				} else {
					require.NotEmpty(t, result)
					model := result[0]
					assert.Equal(t, tc.wantModel.ID, model.ID)
					assert.Equal(t, tc.wantModel.Name, model.Name)
				}
			}
		})
	}
}

func TestKeyHandler_Update(t *testing.T) {
	testCases := []struct {
		name     string
		tenantID string
		key      string
		value    TestModel
		mockFunc func(key string, filter map[string]any, data any) error
		wantErr  bool
	}{
		{
			name:     "successful update",
			tenantID: "1",
			key:      "test-key",
			value:    TestModel{ID: "1", Name: "updated"},
			mockFunc: func(key string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "update with database error",
			tenantID: "1",
			key:      "test-key",
			value:    TestModel{ID: "1", Name: "updated"},
			mockFunc: func(key string, filter map[string]any, data any) error {
				return errors.New("update failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				UpdateFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleDB)
			handler := newKeyHandlerWithMock[TestModel](mockHandler, logger)

			err := handler.Update(tc.tenantID, tc.key, tc.value)
			if tc.wantErr {
				require.Error(t, err)
				appErr, ok := erp_errors.AsAppError(err)
				require.True(t, ok, "Expected AppError")
				assert.Equal(t, erp_errors.CategoryInternal, appErr.Category)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeyHandler_Delete(t *testing.T) {
	testCases := []struct {
		name     string
		tenantID string
		key      string
		mockFunc func(key string, filter map[string]any) error
		wantErr  bool
	}{
		{
			name:     "successful delete",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "delete with database error",
			tenantID: "1",
			key:      "test-key",
			mockFunc: func(key string, filter map[string]any) error {
				return errors.New("delete failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				DeleteFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleDB)
			handler := newKeyHandlerWithMock[TestModel](mockHandler, logger)

			err := handler.Delete(tc.tenantID, tc.key)
			if tc.wantErr {
				require.Error(t, err)
				appErr, ok := erp_errors.AsAppError(err)
				require.True(t, ok, "Expected AppError")
				assert.Equal(t, erp_errors.CategoryInternal, appErr.Category)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
