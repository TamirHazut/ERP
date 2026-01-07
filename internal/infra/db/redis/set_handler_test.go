package redis

import (
	"errors"
	"testing"
	"time"

	mock_redis "erp.localhost/internal/infra/db/redis/mock"
	"erp.localhost/internal/infra/logging/logger"
	model_shared "erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewBaseSetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_redis.NewMockRedisHandler(ctrl)
	logger := logger.NewBaseLogger(model_shared.ModuleDB)

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
			returnError:          errors.New("redis connection failed"),
			wantErr:              true,
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

			logger := logger.NewBaseLogger(model_shared.ModuleDB)
			handler := NewBaseSetHandler(mockHandler, logger)

			err := handler.Add(tc.tenantID, tc.key, tc.member, tc.opts...)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBaseSetHandler_Add_WithTTL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_redis.NewMockRedisHandler(ctrl)

	tenantID := "tenant-1"
	key := "my-set"
	member := "member-1"
	formattedKey := "tenant-1:my-set"
	ttl := 3600
	ttlUnit := "1s"
	opts := []map[string]any{
		{
			"ttl":      ttl,
			"ttl_unit": ttlUnit,
		},
	}

	mockHandler.EXPECT().
		SAdd(formattedKey, member).
		Return(nil).
		Times(1)

	mockHandler.EXPECT().
		Expire(formattedKey, ttl, time.Second).
		Return(nil).
		Times(1)

	logger := logger.NewBaseLogger(model_shared.ModuleDB)
	handler := NewBaseSetHandler(mockHandler, logger)

	err := handler.Add(tenantID, key, member, opts...)
	require.NoError(t, err)
}

func TestBaseSetHandler_Add_WithTTL_ExpireFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_redis.NewMockRedisHandler(ctrl)

	tenantID := "tenant-1"
	key := "my-set"
	member := "member-1"
	formattedKey := "tenant-1:my-set"
	ttl := 3600
	ttlUnit := "1s"
	opts := []map[string]any{
		{
			"ttl":      ttl,
			"ttl_unit": ttlUnit,
		},
	}

	mockHandler.EXPECT().
		SAdd(formattedKey, member).
		Return(nil).
		Times(1)

	mockHandler.EXPECT().
		Expire(formattedKey, ttl, time.Second).
		Return(errors.New("expire failed")).
		Times(1)

	logger := logger.NewBaseLogger(model_shared.ModuleDB)
	handler := NewBaseSetHandler(mockHandler, logger)

	err := handler.Add(tenantID, key, member, opts...)
	require.Error(t, err)
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
			returnError:          errors.New("redis connection failed"),
			wantErr:              true,
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

			logger := logger.NewBaseLogger(model_shared.ModuleDB)
			handler := NewBaseSetHandler(mockHandler, logger)

			err := handler.Remove(tc.tenantID, tc.key, tc.member)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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
			returnError:           errors.New("redis connection failed"),
			wantErr:               true,
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

			logger := logger.NewBaseLogger(model_shared.ModuleDB)
			handler := NewBaseSetHandler(mockHandler, logger)

			members, err := handler.Members(tc.tenantID, tc.key)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, members)
			} else {
				require.NoError(t, err)
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
			returnError:          errors.New("redis connection failed"),
			wantErr:              true,
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

			logger := logger.NewBaseLogger(model_shared.ModuleDB)
			handler := NewBaseSetHandler(mockHandler, logger)

			err := handler.Clear(tc.tenantID, tc.key)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
