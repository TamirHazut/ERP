package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	common_models "erp.localhost/internal/common/models"
	db_mocks "erp.localhost/internal/db/mocks"
	logging "erp.localhost/internal/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestModel is a simple test model for key handler tests
type TestModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestKeyHandler_Set(t *testing.T) {
	testCases := []struct {
		tenantID          string
		name              string
		key               string
		value             TestModel
		returnID          string
		returnError       error
		expectedCallTimes int
	}{
		{
			tenantID:          "1",
			name:              "successful set",
			key:               "test-key",
			value:             TestModel{ID: "1", Name: "test"},
			returnID:          "ok",
			returnError:       nil,
			expectedCallTimes: 1,
		},
		{
			name:              "set with database error",
			tenantID:          "1",
			key:               "test-key",
			value:             TestModel{ID: "1", Name: "test"},
			returnID:          "",
			returnError:       errors.New("database connection failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			mockHandler.EXPECT().Create(formattedKey, tc.value).Return(tc.returnID, tc.returnError).Times(tc.expectedCallTimes)
			handler := NewBaseKeyHandler[TestModel](mockHandler, logging.NewLogger(common_models.ModuleDB))
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
	jsonData, _ := json.Marshal(testModel)

	testCases := []struct {
		name              string
		tenantID          string
		key               string
		returnData        any
		returnError       error
		expectedResult    TestModel
		expectedCallTimes int
	}{
		{
			name:              "successful get one",
			tenantID:          "1",
			key:               "test-key",
			returnData:        string(jsonData),
			returnError:       nil,
			expectedResult:    testModel,
			expectedCallTimes: 1,
		},
		{
			name:              "get one with error",
			tenantID:          "1",
			key:               "test-key",
			returnData:        nil,
			returnError:       errors.New("get one failed"),
			expectedCallTimes: 1,
		},
		{
			name:              "get one with error incompatible type",
			tenantID:          "1",
			key:               "test-key",
			returnData:        "invalid json",
			returnError:       errors.New("invalid json"),
			expectedCallTimes: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			mockHandler.EXPECT().FindOne(formattedKey, nil).Return(tc.returnData, tc.returnError).Times(tc.expectedCallTimes)

			handler := NewBaseKeyHandler[TestModel](mockHandler, logging.NewLogger(common_models.ModuleDB))

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
	testModel := TestModel{ID: "1", Name: "test"}
	jsonData, _ := json.Marshal(testModel)

	testCases := []struct {
		name              string
		tenantID          string
		key               string
		returnData        []any
		returnError       error
		expectedResult    []TestModel
		expectedCallTimes int
	}{
		{
			name:              "successful get",
			tenantID:          "1",
			key:               "test-key",
			returnData:        []any{string(jsonData)},
			returnError:       nil,
			expectedResult:    []TestModel{testModel},
			expectedCallTimes: 1,
		},
		{
			name:              "key not found",
			tenantID:          "1",
			key:               "test-key",
			returnData:        []any{},
			returnError:       nil,
			expectedResult:    []TestModel{},
			expectedCallTimes: 1,
		},
		{
			name:              "database error",
			tenantID:          "1",
			key:               "test-key",
			returnData:        []any{},
			returnError:       errors.New("database query failed"),
			expectedResult:    []TestModel{},
			expectedCallTimes: 1,
		},
		{
			name:              "invalid JSON",
			tenantID:          "1",
			key:               "test-key",
			returnData:        []any{"invalid json"},
			returnError:       errors.New("invalid json"),
			expectedResult:    []TestModel{},
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			mockHandler.EXPECT().FindAll(formattedKey, nil).Return(tc.returnData, tc.returnError).Times(tc.expectedCallTimes)
			handler := NewBaseKeyHandler[TestModel](mockHandler, logging.NewLogger(common_models.ModuleDB))

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
		value             TestModel
		returnError       error
		expectedCallTimes int
	}{
		{
			name:              "successful update",
			tenantID:          "1",
			key:               "test-key",
			value:             TestModel{ID: "1", Name: "updated"},
			returnError:       nil,
			expectedCallTimes: 1,
		},
		{
			name:              "update with database error",
			tenantID:          "1",
			key:               "test-key",
			value:             TestModel{ID: "1", Name: "updated"},
			returnError:       errors.New("update failed"),
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			mockHandler.EXPECT().Update(formattedKey, nil, tc.value).Return(tc.returnError).Times(tc.expectedCallTimes)
			handler := NewBaseKeyHandler[TestModel](mockHandler, logging.NewLogger(common_models.ModuleDB))

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
			mockHandler := db_mocks.NewMockDBHandler(ctrl)
			formattedKey := fmt.Sprintf("%s:%s", tc.tenantID, tc.key)
			mockHandler.EXPECT().Delete(formattedKey, nil).Return(tc.returnError).Times(tc.expectedCallTimes)
			handler := NewBaseKeyHandler[TestModel](mockHandler, logging.NewLogger(common_models.ModuleDB))

			err := handler.Delete(tc.tenantID, tc.key)
			if tc.returnError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
