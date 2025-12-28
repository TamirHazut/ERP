package mongo

import (
	"errors"
	"testing"

	db "erp.localhost/internal/db"
	"erp.localhost/internal/db/mock"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel is a simple test model for collection handler tests
type TestModel struct {
	ID   string `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
}

func TestNewCollectionHandler(t *testing.T) {
	mockHandler := &mock.MockDBHandler{}
	logger := logging.NewLogger(logging.ModuleDB)

	testCases := []struct {
		name       string
		handler    db.DBHandler
		collection string
		logger     *logging.Logger
		wantErr    bool
	}{
		{
			name:       "valid collection handler",
			handler:    mockHandler,
			collection: "test_collection",
			logger:     logger,
			wantErr:    false,
		},
		{
			name:       "nil handler",
			handler:    nil,
			collection: "test_collection",
			logger:     logger,
			wantErr:    false, // Collection creation doesn't validate handler
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewCollectionHandler[TestModel](tc.handler, tc.collection, tc.logger)
			if tc.wantErr {
				assert.Nil(t, repo)
			} else {
				assert.NotNil(t, repo)
				if repo != nil {
					assert.Equal(t, tc.collection, repo.collection)
				}
			}
		})
	}
}

func TestCollection_Create(t *testing.T) {
	testCases := []struct {
		name      string
		item      TestModel
		mockFunc  func(collection string, data any, opts ...map[string]any) (string, error)
		wantID    string
		wantErr   bool
		wantError error
	}{
		{
			name: "successful create",
			item: TestModel{Name: "test"},
			mockFunc: func(collection string, data any, opts ...map[string]any) (string, error) {
				return "created-id", nil
			},
			wantID:  "created-id",
			wantErr: false,
		},
		{
			name: "create with database error",
			item: TestModel{Name: "test"},
			mockFunc: func(collection string, data any, opts ...map[string]any) (string, error) {
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
			repo := NewCollectionHandler[TestModel](mockHandler, "test_collection", logger)

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

func TestCollection_FindOne(t *testing.T) {
	testModel := TestModel{ID: "1", Name: "test"}

	testCases := []struct {
		name      string
		filter    map[string]any
		mockFunc  func(collection string, filter map[string]any) (any, error)
		wantModel TestModel
		wantErr   bool
	}{
		{
			name:   "successful find one",
			filter: map[string]any{"name": "test"},
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				return testModel, nil
			},
			wantModel: testModel,
		},
		{
			name:   "find one with error",
			filter: map[string]any{"name": "test"},
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				return nil, errors.New("find one failed")
			},
			wantErr: true,
		},
		{
			name:     "find one with default behavior",
			filter:   map[string]any{"name": "test"},
			mockFunc: nil,
			wantErr:  true,
		},
		{
			name:   "find one with item not found",
			filter: map[string]any{"name": "test"},
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				return nil, nil
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
			repo := NewCollectionHandler[TestModel](mockHandler, "test_collection", logger)
			result, err := repo.FindOne(tc.filter)
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

func TestCollection_FindAll(t *testing.T) {
	testCases := []struct {
		name     string
		filter   map[string]any
		mockFunc func(collection string, filter map[string]any) ([]any, error)
		wantLen  int
		wantErr  bool
	}{
		{
			name:   "successful find with results",
			filter: map[string]any{"name": "test"},
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
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
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:   "find with database error",
			filter: map[string]any{"name": "test"},
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return nil, errors.New("database query failed")
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindAllFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleDB)
			repo := NewCollectionHandler[TestModel](mockHandler, "test_collection", logger)

			results, err := repo.FindAll(tc.filter)
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

func TestCollection_Update(t *testing.T) {
	testCases := []struct {
		name     string
		filter   map[string]any
		item     TestModel
		mockFunc func(collection string, filter map[string]any, data any, opts ...map[string]any) error
		wantErr  bool
	}{
		{
			name:   "successful update",
			filter: map[string]any{"_id": "1"},
			item:   TestModel{ID: "1", Name: "updated"},
			mockFunc: func(collection string, filter map[string]any, data any, opts ...map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "update with nil filter",
			filter: nil,
			item:   TestModel{ID: "1", Name: "updated"},
			mockFunc: func(collection string, filter map[string]any, data any, opts ...map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:   "update with database error",
			filter: map[string]any{"_id": "1"},
			item:   TestModel{ID: "1", Name: "updated"},
			mockFunc: func(collection string, filter map[string]any, data any, opts ...map[string]any) error {
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
			repo := NewCollectionHandler[TestModel](mockHandler, "test_collection", logger)

			err := repo.Update(tc.filter, tc.item)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCollection_Delete(t *testing.T) {
	testCases := []struct {
		name     string
		filter   map[string]any
		mockFunc func(collection string, filter map[string]any) error
		wantErr  bool
	}{
		{
			name:   "successful delete",
			filter: map[string]any{"_id": "1"},
			mockFunc: func(collection string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "delete with nil filter",
			filter: nil,
			mockFunc: func(collection string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:   "delete with database error",
			filter: map[string]any{"_id": "1"},
			mockFunc: func(collection string, filter map[string]any) error {
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
			repo := NewCollectionHandler[TestModel](mockHandler, "test_collection", logger)

			err := repo.Delete(tc.filter)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
