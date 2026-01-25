package redis

import (
	"errors"
	"fmt"
	"testing"

	mock_db "erp.localhost/internal/infra/db/mock"
	"erp.localhost/internal/infra/logging/logger"
	"erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestModel is a simple test model for key handler tests
type TestModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func createNewHandler(mockDBHandler *mock_db.MockDBHandler) *BaseKeyHandler[TestModel] {
	handler := &BaseKeyHandler[TestModel]{
		dbHandler: mockDBHandler,
		logger:    logger.NewBaseLogger(shared.ModuleDB),
	}
	return handler
}

func TestKeyHandler_Set(t *testing.T) {
	testCases := []struct {
		tenantID          string
		name              string
		key               string
		value             *TestModel
		returnID          string
		returnError       error
		expectedCallTimes int
	}{
		{
			tenantID:          "1",
			name:              "successful set",
			key:               "test-key",
			value:             &TestModel{ID: "1", Name: "test"},
			returnID:          "ok",
			returnError:       nil,
			expectedCallTimes: 1,
		},
		{
			name:              "set with database error",
			tenantID:          "1",
			key:               "test-key",
			value:             &TestModel{ID: "1", Name: "test"},
			returnID:          "",
			returnError:       errors.New("database connection failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := mock_db.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			mockHandler.EXPECT().Create(formattedKey, tc.value).Return(tc.returnID, tc.returnError).Times(tc.expectedCallTimes)
			handler := createNewHandler(mockHandler)
			err := handler.Set(tc.tenantID, tc.key, tc.value)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeyHandler_GetOne(t *testing.T) {
	testModel := TestModel{ID: "1", Name: "test"}

	testCases := []struct {
		name              string
		tenantID          string
		key               string
		returnData        TestModel
		returnError       error
		expectedResult    TestModel
		expectedCallTimes int
	}{
		{
			name:              "successful get one",
			tenantID:          "1",
			key:               "test-key",
			returnData:        testModel,
			returnError:       nil,
			expectedResult:    testModel,
			expectedCallTimes: 1,
		},
		{
			name:              "get one with error",
			tenantID:          "1",
			key:               "test-key",
			returnData:        TestModel{},
			returnError:       errors.New("get one failed"),
			expectedCallTimes: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := mock_db.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			model := &TestModel{}
			mockHandler.EXPECT().
				FindOne(formattedKey, nil, model).
				DoAndReturn(func(formattedKey string, filter map[string]any, result any) error {
					// Cast result to the correct type and set its value
					if m, ok := result.(*TestModel); ok {
						*m = tc.returnData
					}
					return tc.returnError
				}).Times(tc.expectedCallTimes)

			handler := createNewHandler(mockHandler)

			result, err := handler.GetOne(tc.tenantID, tc.key)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, *result)
			}
		})
	}
}

func TestKeyHandler_GetAll(t *testing.T) {

	testCases := []struct {
		name              string
		tenantID          string
		key               string
		returnData        []any
		returnError       error
		expectedResult    []*TestModel
		expectedCallTimes int
	}{
		{
			name:     "successful get",
			tenantID: "1",
			key:      "test-key",
			returnData: []any{
				&TestModel{ID: "1", Name: "test1"},
				&TestModel{ID: "2", Name: "test2"},
			},
			returnError: nil,
			expectedResult: []*TestModel{
				{ID: "1", Name: "test1"},
				{ID: "2", Name: "test2"},
			},
			expectedCallTimes: 1,
		},
		{
			name:              "key not found",
			tenantID:          "1",
			key:               "test-key",
			returnData:        []any{},
			returnError:       nil,
			expectedResult:    []*TestModel{},
			expectedCallTimes: 1,
		},
		{
			name:              "database error",
			tenantID:          "1",
			key:               "test-key",
			returnData:        []any{},
			returnError:       errors.New("database query failed"),
			expectedResult:    []*TestModel{},
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := mock_db.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)

			models := make([]*TestModel, 0)
			mockHandler.EXPECT().
				FindAll(formattedKey, nil, &models).
				DoAndReturn(func(formattedKey string, filter map[string]any, result any) error {
					if m, ok := result.(*[]*TestModel); ok {
						*m = make([]*TestModel, len(tc.returnData))
						for i, item := range tc.returnData {
							(*m)[i] = item.(*TestModel)
						}
					}
					return tc.returnError
				}).Times(tc.expectedCallTimes)

			handler := createNewHandler(mockHandler)

			result, err := handler.GetAll(tc.tenantID, tc.key)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestKeyHandler_Update(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		key               string
		value             *TestModel
		returnError       error
		expectedCallTimes int
	}{
		{
			name:              "successful update",
			tenantID:          "1",
			key:               "test-key",
			value:             &TestModel{ID: "1", Name: "updated"},
			returnError:       nil,
			expectedCallTimes: 1,
		},
		{
			name:              "update with database error",
			tenantID:          "1",
			key:               "test-key",
			value:             &TestModel{ID: "1", Name: "updated"},
			returnError:       errors.New("update failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := mock_db.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			mockHandler.EXPECT().Update(formattedKey, nil, tc.value).Return(tc.returnError).Times(tc.expectedCallTimes)
			handler := createNewHandler(mockHandler)

			err := handler.Update(tc.tenantID, tc.key, tc.value)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeyHandler_Delete(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		key               string
		returnError       error
		expectedCallTimes int
	}{
		{
			name:              "successful delete",
			tenantID:          "1",
			key:               "test-key",
			returnError:       nil,
			expectedCallTimes: 1,
		},
		{
			name:              "delete with database error",
			tenantID:          "1",
			key:               "test-key",
			returnError:       errors.New("delete failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := mock_db.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			mockHandler.EXPECT().Delete(formattedKey, nil).Return(tc.returnError).Times(tc.expectedCallTimes)
			handler := createNewHandler(mockHandler)

			err := handler.Delete(tc.tenantID, tc.key)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
