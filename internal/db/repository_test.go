package db

import (
	"errors"
	"testing"

	"erp.localhost/internal/db/mock"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel is a simple test model for repository tests
type TestModel struct {
	ID   string `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
}

func TestNewRepository(t *testing.T) {
	mockHandler := &mock.MockDBHandler{}
	logger := logging.NewLogger(logging.ModuleDB)

	testCases := []struct {
		name    string
		handler DBHandler
		dbName  string
		logger  *logging.Logger
		wantErr bool
	}{
		{
			name:    "valid repository",
			handler: mockHandler,
			dbName:  "test_db",
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "nil handler",
			handler: nil,
			dbName:  "test_db",
			logger:  logger,
			wantErr: false, // Repository creation doesn't validate handler
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewRepository[TestModel](tc.handler, tc.dbName, tc.logger)
			if tc.wantErr {
				assert.Nil(t, repo)
			} else {
				assert.NotNil(t, repo)
				if repo != nil {
					assert.Equal(t, tc.dbName, repo.dbName)
				}
			}
		})
	}
}

func TestRepository_Create(t *testing.T) {
	testCases := []struct {
		name      string
		item      TestModel
		mockFunc  func(db string, data any) (string, error)
		wantID    string
		wantErr   bool
		wantError error
	}{
		{
			name: "successful create",
			item: TestModel{Name: "test"},
			mockFunc: func(db string, data any) (string, error) {
				return "created-id", nil
			},
			wantID:  "created-id",
			wantErr: false,
		},
		{
			name: "create with database error",
			item: TestModel{Name: "test"},
			mockFunc: func(db string, data any) (string, error) {
				return "", errors.New("database connection failed")
			},
			wantID:  "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				CreateFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleDB)
			repo := NewRepository[TestModel](mockHandler, "test_db", logger)

			id, err := repo.Create(tc.item)
			if tc.wantErr {
				require.Error(t, err)
				appErr, ok := erp_errors.AsAppError(err)
				require.True(t, ok, "Expected AppError")
				assert.Equal(t, erp_errors.CategoryInternal, appErr.Category)
				assert.Empty(t, id)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantID, id)
			}
		})
	}
}

func TestRepository_Find(t *testing.T) {
	testCases := []struct {
		name     string
		filter   map[string]any
		mockFunc func(db string, filter map[string]any) ([]any, error)
		wantLen  int
		wantErr  bool
	}{
		{
			name:   "successful find with results",
			filter: map[string]any{"name": "test"},
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					TestModel{ID: "1", Name: "test1"},
					TestModel{ID: "2", Name: "test2"},
				}, nil
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:   "successful find with no results",
			filter: map[string]any{"name": "nonexistent"},
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:   "find with database error",
			filter: map[string]any{"name": "test"},
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return nil, errors.New("database query failed")
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleDB)
			repo := NewRepository[TestModel](mockHandler, "test_db", logger)

			results, err := repo.Find(tc.filter)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, results)
			} else {
				require.NoError(t, err)
				assert.Len(t, results, tc.wantLen)
			}
		})
	}
}

func TestRepository_Update(t *testing.T) {
	testCases := []struct {
		name     string
		filter   map[string]any
		item     TestModel
		mockFunc func(db string, filter map[string]any, data any) error
		wantErr  bool
	}{
		{
			name:   "successful update",
			filter: map[string]any{"_id": "1"},
			item:   TestModel{ID: "1", Name: "updated"},
			mockFunc: func(db string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "update with nil filter",
			filter: nil,
			item:   TestModel{ID: "1", Name: "updated"},
			mockFunc: func(db string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:   "update with database error",
			filter: map[string]any{"_id": "1"},
			item:   TestModel{ID: "1", Name: "updated"},
			mockFunc: func(db string, filter map[string]any, data any) error {
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
			repo := NewRepository[TestModel](mockHandler, "test_db", logger)

			err := repo.Update(tc.filter, tc.item)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRepository_Delete(t *testing.T) {
	testCases := []struct {
		name     string
		filter   map[string]any
		mockFunc func(db string, filter map[string]any) error
		wantErr  bool
	}{
		{
			name:   "successful delete",
			filter: map[string]any{"_id": "1"},
			mockFunc: func(db string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "delete with nil filter",
			filter: nil,
			mockFunc: func(db string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:   "delete with database error",
			filter: map[string]any{"_id": "1"},
			mockFunc: func(db string, filter map[string]any) error {
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
			repo := NewRepository[TestModel](mockHandler, "test_db", logger)

			err := repo.Delete(tc.filter)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
