package redis

import (
	"errors"
	"testing"

	mock_redis "erp.localhost/internal/infra/db/redis/mock"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	"erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewBaseSetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_redis.NewMockRedisHandler(ctrl)
	logger := logger.NewBaseLogger(shared.ModuleDB)

	handler := NewBaseSetHandler(mockHandler, logger)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.redisHandler)
	assert.NotNil(t, handler.logger)
}

func TestBaseSetHandler_Add(t *testing.T) {
	testCases := []struct {
		name                 string
		tenantID             string
		key                  string
		member               string
		opts                 []map[string]any
		expectedFormattedKey string
		returnError          error
		wantErr              bool
		errCategory          infra_error.ErrorCategory
		expectedSAddCalls    int
	}{
		{
			name:                 "successful add",
			tenantID:             "tenant-1",
			key:                  "my-set",
			member:               "member-1",
			opts:                 nil,
			expectedFormattedKey: "tenant-1:my-set",
			returnError:          nil,
			wantErr:              false,
			expectedSAddCalls:    1,
		},
		{
			name:                 "add with database error",
			tenantID:             "tenant-1",
			key:                  "my-set",
			member:               "member-1",
			opts:                 nil,
			expectedFormattedKey: "tenant-1:my-set",
			returnError:          infra_error.Internal(infra_error.InternalDatabaseError, errors.New("redis connection failed")),
			wantErr:              true,
			errCategory:          infra_error.CategoryInternal,
			expectedSAddCalls:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockRedisHandler(ctrl)
			if tc.expectedSAddCalls > 0 {
				mockHandler.EXPECT().
					SAdd(tc.expectedFormattedKey, tc.member).
					Return(tc.returnError).
					Times(tc.expectedSAddCalls)
			}

			logger := logger.NewBaseLogger(shared.ModuleDB)
			handler := NewBaseSetHandler(mockHandler, logger)

			err := handler.Add(tc.tenantID, tc.key, tc.member, tc.opts...)
			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, err.Category, tc.errCategory)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestBaseSetHandler_Remove(t *testing.T) {
	testCases := []struct {
		name                 string
		tenantID             string
		key                  string
		member               string
		expectedFormattedKey string
		returnError          error
		wantErr              bool
		errCategory          infra_error.ErrorCategory
		expectedSRemCalls    int
	}{
		{
			name:                 "successful remove",
			tenantID:             "tenant-1",
			key:                  "my-set",
			member:               "member-1",
			expectedFormattedKey: "tenant-1:my-set",
			returnError:          nil,
			wantErr:              false,
			expectedSRemCalls:    1,
		},
		{
			name:                 "remove with database error",
			tenantID:             "tenant-1",
			key:                  "my-set",
			member:               "member-1",
			expectedFormattedKey: "tenant-1:my-set",
			returnError:          infra_error.Internal(infra_error.InternalDatabaseError, errors.New("redis connection failed")),
			wantErr:              true,
			errCategory:          infra_error.CategoryInternal,
			expectedSRemCalls:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockRedisHandler(ctrl)
			if tc.expectedSRemCalls > 0 {
				mockHandler.EXPECT().
					SRem(tc.expectedFormattedKey, tc.member).
					Return(tc.returnError).
					Times(tc.expectedSRemCalls)
			}

			logger := logger.NewBaseLogger(shared.ModuleDB)
			handler := NewBaseSetHandler(mockHandler, logger)

			err := handler.Remove(tc.tenantID, tc.key, tc.member)
			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, err.Category, tc.errCategory)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestBaseSetHandler_Members(t *testing.T) {
	testCases := []struct {
		name                  string
		tenantID              string
		key                   string
		expectedFormattedKey  string
		returnMembers         []string
		returnError           error
		wantErr               bool
		errCategory           infra_error.ErrorCategory
		expectedSMembersCalls int
	}{
		{
			name:                  "successful get members",
			tenantID:              "tenant-1",
			key:                   "my-set",
			expectedFormattedKey:  "tenant-1:my-set",
			returnMembers:         []string{"member-1", "member-2", "member-3"},
			returnError:           nil,
			wantErr:               false,
			expectedSMembersCalls: 1,
		},
		{
			name:                  "get members from empty set",
			tenantID:              "tenant-1",
			key:                   "my-set",
			expectedFormattedKey:  "tenant-1:my-set",
			returnMembers:         []string{},
			returnError:           nil,
			wantErr:               false,
			expectedSMembersCalls: 1,
		},
		{
			name:                  "get members with database error",
			tenantID:              "tenant-1",
			key:                   "my-set",
			expectedFormattedKey:  "tenant-1:my-set",
			returnMembers:         nil,
			returnError:           infra_error.Internal(infra_error.InternalDatabaseError, errors.New("redis connection failed")),
			wantErr:               true,
			errCategory:           infra_error.CategoryInternal,
			expectedSMembersCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockRedisHandler(ctrl)
			if tc.expectedSMembersCalls > 0 {
				mockHandler.EXPECT().
					SMembers(tc.expectedFormattedKey).
					Return(tc.returnMembers, tc.returnError).
					Times(tc.expectedSMembersCalls)
			}

			logger := logger.NewBaseLogger(shared.ModuleDB)
			handler := NewBaseSetHandler(mockHandler, logger)

			members, err := handler.Members(tc.tenantID, tc.key)
			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, err.Category, tc.errCategory)
				assert.Nil(t, members)
			} else {
				require.Nil(t, err)
				assert.Equal(t, tc.returnMembers, members)
			}
		})
	}
}

func TestBaseSetHandler_Clear(t *testing.T) {
	testCases := []struct {
		name                 string
		tenantID             string
		key                  string
		expectedFormattedKey string
		returnError          error
		wantErr              bool
		errCategory          infra_error.ErrorCategory
		expectedClearCalls   int
	}{
		{
			name:                 "successful clear",
			tenantID:             "tenant-1",
			key:                  "my-set",
			expectedFormattedKey: "tenant-1:my-set",
			returnError:          nil,
			wantErr:              false,
			expectedClearCalls:   1,
		},
		{
			name:                 "clear with database error",
			tenantID:             "tenant-1",
			key:                  "my-set",
			expectedFormattedKey: "tenant-1:my-set",
			returnError:          infra_error.Internal(infra_error.InternalDatabaseError, errors.New("redis connection failed")),
			wantErr:              true,
			errCategory:          infra_error.CategoryInternal,
			expectedClearCalls:   1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockRedisHandler(ctrl)
			if tc.expectedClearCalls > 0 {
				mockHandler.EXPECT().
					Clear(tc.expectedFormattedKey).
					Return(tc.returnError).
					Times(tc.expectedClearCalls)
			}

			logger := logger.NewBaseLogger(shared.ModuleDB)
			handler := NewBaseSetHandler(mockHandler, logger)

			err := handler.Clear(tc.tenantID, tc.key)
			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, err.Category, tc.errCategory)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
