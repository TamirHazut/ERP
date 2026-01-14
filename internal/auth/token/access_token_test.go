package token

import (
	"errors"
	"testing"
	"time"

	mock_redis "erp.localhost/internal/infra/db/redis/mock"
	"erp.localhost/internal/infra/logging/logger"
	model_auth_cache "erp.localhost/internal/infra/model/auth/cache"
	model_shared "erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// tokenMetadataMatcher is a custom gomock matcher for TokenMetadata objects
// It skips the RevokedAt field which is set dynamically in Revoke operations
type tokenMetadataMatcher struct {
	expected model_auth_cache.TokenMetadata
}

func (m tokenMetadataMatcher) Matches(x interface{}) bool {
	metadata, ok := x.(model_auth_cache.TokenMetadata)
	if !ok {
		return false
	}
	// Match all fields except RevokedAt which is set dynamically
	return metadata.TokenID == m.expected.TokenID &&
		metadata.UserID == m.expected.UserID &&
		metadata.TenantID == m.expected.TenantID &&
		metadata.TokenType == m.expected.TokenType &&
		metadata.Revoked == m.expected.Revoked &&
		metadata.RevokedBy == m.expected.RevokedBy
}

func (m tokenMetadataMatcher) String() string {
	return "matches token metadata fields except RevokedAt"
}

func TestNewAccessTokenKeyHandler(t *testing.T) {
	// Note: This test requires a running Redis instance
	t.Skip("Skipping test that requires Redis connection")
}

func TestAccessTokenKeyHandler_Store(t *testing.T) {
	validMetadata := model_auth_cache.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		metadata                model_auth_cache.TokenMetadata
		returnGetOneMetadata    *model_auth_cache.TokenMetadata
		returnGetOneError       error
		returnSetError          error
		wantErr                 bool
		expectedTenantID        string
		expectedUserID          string
		expectedSetCallTimes    int
		expectedGetOneCallTimes int
	}{
		{
			name:                    "successful store",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			metadata:                validMetadata,
			returnGetOneMetadata:    nil,
			returnGetOneError:       nil,
			returnSetError:          nil,
			wantErr:                 false,
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			expectedSetCallTimes:    1,
			expectedGetOneCallTimes: 1,
		},
		{
			name:     "store with missing tokenID",
			tenantID: "tenant-123",
			userID:   "user-123",
			metadata: model_auth_cache.TokenMetadata{
				UserID:    "user-123",
				TenantID:  "tenant-123",
				TokenType: "access",
				ExpiresAt: time.Now().Add(time.Hour),
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
			metadata: model_auth_cache.TokenMetadata{
				TokenID:   "token-123",
				UserID:    "user-123",
				TenantID:  "wrong-tenant",
				TokenType: "access",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectedTenantID:     "",
			expectedUserID:       "",
			returnSetError:       nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:                    "store with database error",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			metadata:                validMetadata,
			returnGetOneMetadata:    nil,
			returnGetOneError:       nil,
			returnSetError:          errors.New("database connection failed"),
			wantErr:                 true,
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			expectedSetCallTimes:    1,
			expectedGetOneCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[model_auth_cache.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.userID).
					Return(tc.returnGetOneMetadata, tc.returnGetOneError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedSetCallTimes > 0 {
				mockHandler.EXPECT().
					Set(tc.expectedTenantID, tc.expectedUserID, tc.metadata).
					Return(tc.returnSetError).
					Times(tc.expectedSetCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

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
	validMetadata := model_auth_cache.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		expectedTenantID        string
		expectedUserID          string
		returnMetadata          *model_auth_cache.TokenMetadata
		returnError             error
		wantToken               *model_auth_cache.TokenMetadata
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

			mockHandler := mock_redis.NewMockKeyHandler[model_auth_cache.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.expectedUserID).
					Return(tc.returnMetadata, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

			result, err := handler.GetOne(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantToken.TokenID, result.TokenID)
			}
		})
	}
}

func TestAccessTokenKeyHandler_Validate(t *testing.T) {
	validMetadata := model_auth_cache.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}
	expiredMetadata := model_auth_cache.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour), // Expired
		Revoked:   false,
	}
	revokedMetadata := model_auth_cache.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   true,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		expectedTenantID        string
		expectedUserID          string
		returnMetadata          *model_auth_cache.TokenMetadata
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

			mockHandler := mock_redis.NewMockKeyHandler[model_auth_cache.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.expectedUserID).
					Return(tc.returnMetadata, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

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
	validMetadata := model_auth_cache.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		revokedBy               string
		expectedGetTenantID     string
		expectedGetUserID       string
		expectedUpdateTenantID  string
		expectedUpdateUserID    string
		returnGetMetadata       *model_auth_cache.TokenMetadata
		returnGetError          error
		returnUpdateError       error
		wantErr                 bool
		expectedGetOneCallTimes int
		expectedUpdateCallTimes int
	}{
		{
			name:                    "successful revoke",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			revokedBy:               "admin",
			expectedGetTenantID:     "tenant-123",
			expectedGetUserID:       "user-123",
			expectedUpdateTenantID:  "tenant-123",
			expectedUpdateUserID:    "user-123",
			returnGetMetadata:       &validMetadata,
			returnGetError:          nil,
			returnUpdateError:       nil,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
			expectedUpdateCallTimes: 1,
		},
		{
			name:                    "token not found",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			revokedBy:               "admin",
			expectedGetTenantID:     "tenant-123",
			expectedGetUserID:       "user-123",
			expectedUpdateTenantID:  "",
			expectedUpdateUserID:    "",
			returnGetMetadata:       nil,
			returnGetError:          errors.New("token not found"),
			returnUpdateError:       nil,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
			expectedUpdateCallTimes: 0,
		},
		{
			name:                    "update fails",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			revokedBy:               "admin",
			expectedGetTenantID:     "tenant-123",
			expectedGetUserID:       "user-123",
			expectedUpdateTenantID:  "tenant-123",
			expectedUpdateUserID:    "user-123",
			returnGetMetadata:       &validMetadata,
			returnGetError:          nil,
			returnUpdateError:       errors.New("update failed"),
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
			expectedUpdateCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[model_auth_cache.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedGetTenantID, tc.expectedGetUserID).
					Return(tc.returnGetMetadata, tc.returnGetError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedUpdateCallTimes > 0 {
				// Create expected metadata with Revoked=true and RevokedBy set
				expectedMetadata := *tc.returnGetMetadata
				expectedMetadata.Revoked = true
				expectedMetadata.RevokedBy = tc.revokedBy
				mockHandler.EXPECT().
					Update(tc.expectedUpdateTenantID, tc.expectedUpdateUserID, tokenMetadataMatcher{expected: expectedMetadata}).
					Return(tc.returnUpdateError).
					Times(tc.expectedUpdateCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

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

			mockHandler := mock_redis.NewMockKeyHandler[model_auth_cache.TokenMetadata](ctrl)
			if tc.expectedDeleteCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.expectedDeleteTenantID, tc.expectedDeleteUserID).
					Return(tc.returnDeleteError).
					Times(tc.expectedDeleteCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

			err := handler.Delete(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
