package handler

import (
	"errors"
	"testing"
	"time"

	mock_redis "erp.localhost/internal/infra/db/redis/mock"
	"erp.localhost/internal/infra/logging/logger"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	"erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// tokenMetadataMatcher is a custom gomock matcher for TokenMetadata objects
// It skips the RevokedAt field which is set dynamically in Revoke operations
type tokenMetadataMatcher struct {
	expected *authv1_cache.TokenMetadata
}

func (m tokenMetadataMatcher) Matches(x interface{}) bool {
	metadata, ok := x.(*authv1_cache.TokenMetadata)
	if !ok {
		return false
	}
	// Match all fields except RevokedAt which is set dynamically
	return metadata.UserId == m.expected.UserId &&
		metadata.TenantId == m.expected.TenantId &&
		metadata.Revoked == m.expected.Revoked &&
		metadata.RevokedBy == m.expected.RevokedBy
}

func (m tokenMetadataMatcher) String() string {
	return "matches token metadata fields except RevokedAt"
}

func createNewAccessTokenHandler(mockHandler *mock_redis.MockKeyHandler[authv1_cache.TokenMetadata]) *AccessTokenHandler {
	handler := &AccessTokenHandler{
		handler: mockHandler,
		logger:  logger.NewBaseLogger(shared.ModuleAuth),
	}
	return handler
}
func TestAccessTokenKeyHandler_Store(t *testing.T) {
	validMetadata := &authv1_cache.TokenMetadata{
		UserId:    "user-123",
		TenantId:  "tenant-123",
		IssuedAt:  timestamppb.Now(),
		ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
		Revoked:   false,
		Jti:       "test",
	}

	testCases := []struct {
		name                 string
		tenantID             string
		userID               string
		metadata             *authv1_cache.TokenMetadata
		returnSetError       error
		wantErr              bool
		expectedTenantID     string
		expectedUserID       string
		expectedSetCallTimes int
	}{
		{
			name:                 "successful store",
			tenantID:             "tenant-123",
			userID:               "user-123",
			metadata:             validMetadata,
			returnSetError:       nil,
			wantErr:              false,
			expectedTenantID:     "tenant-123",
			expectedUserID:       "user-123",
			expectedSetCallTimes: 1,
		},
		{
			name:     "store with missing userID",
			tenantID: "tenant-123",
			userID:   "user-123",
			metadata: &authv1_cache.TokenMetadata{
				TenantId:  "tenant-123",
				ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
			},
			expectedTenantID:     "",
			expectedUserID:       "",
			returnSetError:       nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:     "store with tenant_id mismatch",
			tenantID: "tenant-123",
			userID:   "user-123",
			metadata: &authv1_cache.TokenMetadata{
				UserId:    "user-123",
				TenantId:  "wrong-tenant",
				ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
			},
			expectedTenantID:     "",
			expectedUserID:       "",
			returnSetError:       nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:                 "store with database error",
			tenantID:             "tenant-123",
			userID:               "user-123",
			metadata:             validMetadata,
			returnSetError:       errors.New("database connection failed"),
			wantErr:              true,
			expectedTenantID:     "tenant-123",
			expectedUserID:       "user-123",
			expectedSetCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.TokenMetadata](ctrl)
			if tc.expectedSetCallTimes > 0 {
				mockHandler.EXPECT().
					Set(tc.expectedTenantID, tc.expectedUserID, tc.metadata, gomock.Any()).
					Return(tc.returnSetError).
					Times(tc.expectedSetCallTimes)
			}

			handler := createNewAccessTokenHandler(mockHandler)

			err := handler.Store(tc.tenantID, tc.userID, tc.metadata)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccessTokenKeyHandler_GetOne(t *testing.T) {
	validMetadata := authv1_cache.TokenMetadata{
		UserId:    "user-123",
		TenantId:  "tenant-123",
		IssuedAt:  timestamppb.Now(),
		ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
		Revoked:   false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		expectedTenantID        string
		expectedUserID          string
		returnMetadata          *authv1_cache.TokenMetadata
		returnError             error
		wantToken               *authv1_cache.TokenMetadata
		wantErr                 bool
		expectedGetOneCallTimes int
	}{
		{
			name:                    "successful get",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnMetadata:          &validMetadata,
			returnError:             nil,
			wantToken:               &validMetadata,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "token not found",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnMetadata:          nil,
			returnError:             errors.New("token not found"),
			wantToken:               nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "database error",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnMetadata:          nil,
			returnError:             errors.New("database query failed"),
			wantToken:               nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.expectedUserID).
					Return(tc.returnMetadata, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			handler := createNewAccessTokenHandler(mockHandler)

			result, err := handler.GetOne(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantToken.UserId, result.UserId)
				assert.Equal(t, tc.wantToken.TenantId, result.TenantId)
			}
		})
	}
}

func TestAccessTokenKeyHandler_Validate(t *testing.T) {
	validMetadata := authv1_cache.TokenMetadata{
		UserId:    "user-123",
		TenantId:  "tenant-123",
		IssuedAt:  timestamppb.Now(),
		ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
		Revoked:   false,
	}
	expiredMetadata := authv1_cache.TokenMetadata{
		UserId:    "user-123",
		TenantId:  "tenant-123",
		IssuedAt:  timestamppb.New(time.Now().Add(-2 * time.Hour)),
		ExpiresAt: timestamppb.New(time.Now().Add(-time.Hour)), // Expired
		Revoked:   false,
	}
	revokedMetadata := authv1_cache.TokenMetadata{
		UserId:    "user-123",
		TenantId:  "tenant-123",
		IssuedAt:  timestamppb.Now(),
		ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
		Revoked:   true,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		expectedTenantID        string
		expectedUserID          string
		returnMetadata          *authv1_cache.TokenMetadata
		returnError             error
		wantErr                 bool
		expectedGetOneCallTimes int
	}{
		{
			name:                    "valid token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnMetadata:          &validMetadata,
			returnError:             nil,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "expired token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnMetadata:          &expiredMetadata,
			returnError:             nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "revoked token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnMetadata:          &revokedMetadata,
			returnError:             nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "token not found",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnMetadata:          nil,
			returnError:             errors.New("token not found"),
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.expectedUserID).
					Return(tc.returnMetadata, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			handler := createNewAccessTokenHandler(mockHandler)

			result, err := handler.Validate(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAccessTokenKeyHandler_Revoke(t *testing.T) {
	validMetadata := &authv1_cache.TokenMetadata{
		UserId:    "user-123",
		TenantId:  "tenant-123",
		IssuedAt:  timestamppb.Now(),
		ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
		Revoked:   false,
		Jti:       "test",
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		revokedBy               string
		expectedGetTenantID     string
		expectedGetUserID       string
		expectedDeleteTenantID  string
		expectedDeleteUserID    string
		returnGetMetadata       *authv1_cache.TokenMetadata
		returnGetError          error
		returnDeleteError       error
		wantErr                 bool
		expectedGetOneCallTimes int
		expectedDeleteCallTimes int
	}{
		{
			name:                    "successful revoke",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			revokedBy:               "admin",
			expectedGetTenantID:     "tenant-123",
			expectedGetUserID:       "user-123",
			expectedDeleteTenantID:  "tenant-123",
			expectedDeleteUserID:    "user-123",
			returnGetMetadata:       validMetadata,
			returnGetError:          nil,
			returnDeleteError:       nil,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
			expectedDeleteCallTimes: 1,
		},
		{
			name:                    "token not found",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			revokedBy:               "admin",
			expectedGetTenantID:     "tenant-123",
			expectedGetUserID:       "user-123",
			expectedDeleteTenantID:  "",
			expectedDeleteUserID:    "",
			returnGetMetadata:       nil,
			returnGetError:          errors.New("token not found"),
			returnDeleteError:       nil,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
			expectedDeleteCallTimes: 0,
		},
		{
			name:                    "update fails",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			revokedBy:               "admin",
			expectedGetTenantID:     "tenant-123",
			expectedGetUserID:       "user-123",
			expectedDeleteTenantID:  "tenant-123",
			expectedDeleteUserID:    "user-123",
			returnGetMetadata:       validMetadata,
			returnGetError:          nil,
			returnDeleteError:       errors.New("update failed"),
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
			expectedDeleteCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedGetTenantID, tc.expectedGetUserID).
					Return(tc.returnGetMetadata, tc.returnGetError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedDeleteCallTimes > 0 {
				// Create expected metadata with Revoked=true and RevokedBy set
				expectedMetadata := tc.returnGetMetadata
				expectedMetadata.Revoked = true
				expectedMetadata.RevokedBy = tc.revokedBy
				mockHandler.EXPECT().
					Delete(tc.expectedDeleteTenantID, tc.expectedDeleteUserID).
					Return(tc.returnDeleteError).
					Times(tc.expectedDeleteCallTimes)
			}

			handler := createNewAccessTokenHandler(mockHandler)

			err := handler.Revoke(tc.tenantID, tc.userID, tc.revokedBy)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccessTokenKeyHandler_Delete(t *testing.T) {
	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		expectedDeleteTenantID  string
		expectedDeleteUserID    string
		returnDeleteError       error
		wantErr                 bool
		expectedDeleteCallTimes int
	}{
		{
			name:                    "successful delete",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedDeleteTenantID:  "tenant-123",
			expectedDeleteUserID:    "user-123",
			returnDeleteError:       nil,
			wantErr:                 false,
			expectedDeleteCallTimes: 1,
		},
		{
			name:                    "delete with database error",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedDeleteTenantID:  "tenant-123",
			expectedDeleteUserID:    "user-123",
			returnDeleteError:       errors.New("delete failed"),
			wantErr:                 true,
			expectedDeleteCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.TokenMetadata](ctrl)
			if tc.expectedDeleteCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.expectedDeleteTenantID, tc.expectedDeleteUserID).
					Return(tc.returnDeleteError).
					Times(tc.expectedDeleteCallTimes)
			}

			handler := createNewAccessTokenHandler(mockHandler)

			err := handler.Delete(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
