package mongo

import (
	"errors"
	"testing"

	db_mocks "erp.localhost/internal/db/mocks"
	"erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestModel is a simple test model for collection handler tests
type TestModel struct {
	ID   string `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
}

func TestCollection_Create(t *testing.T) {
	testCases := []struct {
		name        string
		collection  string
		data        TestModel
		returnID    string
		returnError error
	}{
		{
			name:        "successful create",
			collection:  "test_collection",
			data:        TestModel{Name: "test"},
			returnID:    "created-id",
			returnError: nil,
		},
		{
			name:        "create with database error",
			collection:  "test_collection",
			data:        TestModel{Name: "test"},
			returnID:    "",
			returnError: errors.New("database connection failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			mockHandler.EXPECT().Create(tc.collection, tc.data).Return(tc.returnID, tc.returnError)

			collectionHanlder := BaseCollectionHandler[TestModel]{
				dbHandler:  mockHandler,
				collection: tc.collection,
				logger:     logging.NewLogger(logging.ModuleDB),
			}

			id, err := collectionHanlder.Create(tc.data)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.returnID, id)
			}
		})
	}
}

func TestCollection_FindOne(t *testing.T) {
	testModel := TestModel{ID: "1", Name: "test"}

	testCases := []struct {
		name        string
		collection  string
		filter      map[string]any
		returnModel TestModel
		returnError error
	}{
		{
			name:        "successful find one",
			collection:  "test_collection",
			filter:      map[string]any{"name": "test"},
			returnModel: testModel,
			returnError: nil,
		},
		{
			name:        "find one with error - missing collection",
			filter:      map[string]any{"name": "test"},
			returnModel: TestModel{},
			returnError: errors.New("find one failed"),
		},
		{
			name:        "find one with error - item not found",
			collection:  "test_collection",
			filter:      map[string]any{"name": "test"},
			returnModel: TestModel{},
			returnError: errors.New("no result found"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			mockHandler.EXPECT().FindOne(tc.collection, tc.filter).Return(tc.returnModel, tc.returnError)

			collectionHanlder := BaseCollectionHandler[TestModel]{
				dbHandler:  mockHandler,
				collection: tc.collection,
				logger:     logging.NewLogger(logging.ModuleDB),
			}
			result, err := collectionHanlder.FindOne(tc.filter)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.returnModel, *result)
			}
		})
	}
}

func TestCollection_FindAll(t *testing.T) {
	testCases := []struct {
		name           string
		collection     string
		filter         map[string]any
		returnModels   []any
		returnError    error
		expectedResult []TestModel
	}{
		{
			name:       "successful find with results",
			collection: "test_collection",
			filter:     map[string]any{"name": "test"},
			returnModels: []any{
				TestModel{ID: "1", Name: "test1"},
				TestModel{ID: "2", Name: "test2"},
			},
			returnError: nil,
			expectedResult: []TestModel{
				{ID: "1", Name: "test1"},
				{ID: "2", Name: "test2"},
			},
		},
		{
			name:           "successful find with no results",
			collection:     "test_collection",
			filter:         map[string]any{"name": "nonexistent"},
			returnModels:   []any{},
			returnError:    nil,
			expectedResult: []TestModel{},
		},
		{
			name:         "find with database error",
			collection:   "test_collection",
			filter:       map[string]any{"name": "test"},
			returnModels: []any{},
			returnError:  errors.New("database query failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			mockHandler.EXPECT().FindAll(tc.collection, tc.filter).Return(tc.returnModels, tc.returnError)

			collectionHanlder := BaseCollectionHandler[TestModel]{
				dbHandler:  mockHandler,
				collection: tc.collection,
				logger:     logging.NewLogger(logging.ModuleDB),
			}
			results, err := collectionHanlder.FindAll(tc.filter)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, results)
			}
		})
	}
}

func TestCollection_Update(t *testing.T) {
	testCases := []struct {
		name              string
		collection        string
		filter            map[string]any
		item              TestModel
		returnError       error
		expectedCallTimes int
	}{
		{
			name:              "successful update",
			collection:        "test_collection",
			filter:            map[string]any{"_id": "1"},
			item:              TestModel{ID: "1", Name: "updated"},
			returnError:       nil,
			expectedCallTimes: 1,
		},
		{
			name:              "update with nil filter",
			collection:        "test_collection",
			filter:            nil,
			item:              TestModel{ID: "1", Name: "updated"},
			returnError:       errors.New("filter is required and cannot be nil"),
			expectedCallTimes: 0,
		},
		{
			name:              "update with database error",
			collection:        "test_collection",
			filter:            map[string]any{"_id": "1"},
			item:              TestModel{ID: "1", Name: "updated"},
			returnError:       errors.New("update failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			mockHandler.EXPECT().Update(tc.collection, tc.filter, tc.item).Return(tc.returnError).Times(tc.expectedCallTimes)

			collectionHanlder := BaseCollectionHandler[TestModel]{
				dbHandler:  mockHandler,
				collection: tc.collection,
				logger:     logging.NewLogger(logging.ModuleDB),
			}
			err := collectionHanlder.Update(tc.filter, tc.item)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCollection_Delete(t *testing.T) {
	testCases := []struct {
		name              string
		collection        string
		filter            map[string]any
		returnError       error
		expectedCallTimes int
	}{
		{
			name:              "successful delete",
			collection:        "test_collection",
			filter:            map[string]any{"_id": "1"},
			returnError:       nil,
			expectedCallTimes: 1,
		},
		{
			name:              "delete with nil filter",
			collection:        "test_collection",
			filter:            nil,
			returnError:       errors.New("filter is required and cannot be nil"),
			expectedCallTimes: 0,
		},
		{
			name:              "delete with database error",
			collection:        "test_collection",
			filter:            map[string]any{"_id": "1"},
			returnError:       errors.New("delete failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			mockHandler.EXPECT().Delete(tc.collection, tc.filter).Return(tc.returnError).Times(tc.expectedCallTimes)

			collectionHanlder := BaseCollectionHandler[TestModel]{
				dbHandler:  mockHandler,
				collection: tc.collection,
				logger:     logging.NewLogger(logging.ModuleDB),
			}
			err := collectionHanlder.Delete(tc.filter)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
