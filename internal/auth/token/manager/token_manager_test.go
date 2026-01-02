package token

import (
	"errors"
	"testing"
	"time"

	"erp.localhost/internal/auth/models"
	auth_models "erp.localhost/internal/auth/models/cache"
	handlers_mocks "erp.localhost/internal/auth/token/handlers/mocks"
	common_models "erp.localhost/internal/common/models"
	erp_errors "erp.localhost/internal/errors"
	logging "erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewTokenManager(t *testing.T) {
	tm := NewTokenManager("test-secret-key-12345", time.Hour, 7*24*time.Hour)
	require.NotNil(t, tm)
	assert.NotNil(t, tm.accessTokenHandler)
	assert.NotNil(t, tm.refreshTokenHandler)
	assert.NotNil(t, tm.logger)
}

func TestTokenManager_StoreTokens(t *testing.T) {
	testCases := []struct {
		name                      string
		tenantID                  string
		userID                    string
		accessTokenID             string
		refreshTokenID            string
		accessTokenMetadata       auth_models.TokenMetadata
		refreshToken              models.RefreshToken
		accessStoreError          error
		refreshStoreError         error
		deleteError               error
		wantErr                   bool
		expectedAccessStoreCalls  int
		expectedRefreshStoreCalls int
		expectedDeleteCalls       int
	}{
		{
			name:           "successful store",
			tenantID:       "tenant-1",
			userID:         "user-1",
			accessTokenID:  "token-1",
			refreshTokenID: "refresh-1",
			accessTokenMetadata: auth_models.TokenMetadata{
				TokenID:  "token-1",
				UserID:   "user-1",
				TenantID: "tenant-1",
			},
			refreshToken: models.RefreshToken{
				Token:     "refresh-1",
				UserID:    "user-1",
				TenantID:  "tenant-1",
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				CreatedAt: time.Now(),
			},
			accessStoreError:          nil,
			refreshStoreError:         nil,
			wantErr:                   false,
			expectedAccessStoreCalls:  1,
			expectedRefreshStoreCalls: 1,
		},
		{
			name:           "access token store fails",
			tenantID:       "tenant-1",
			userID:         "user-1",
			accessTokenID:  "token-1",
			refreshTokenID: "refresh-1",
			accessTokenMetadata: auth_models.TokenMetadata{
				TokenID:  "token-1",
				UserID:   "user-1",
				TenantID: "tenant-1",
			},
			refreshToken: models.RefreshToken{
				Token:     "refresh-1",
				UserID:    "user-1",
				TenantID:  "tenant-1",
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				CreatedAt: time.Now(),
			},
			accessStoreError:          errors.New("store failed"),
			refreshStoreError:         nil,
			wantErr:                   true,
			expectedAccessStoreCalls:  1,
			expectedRefreshStoreCalls: 0,
		},
		{
			name:           "refresh token store fails - access token cleaned up",
			tenantID:       "tenant-1",
			userID:         "user-1",
			accessTokenID:  "token-1",
			refreshTokenID: "refresh-1",
			accessTokenMetadata: auth_models.TokenMetadata{
				TokenID:  "token-1",
				UserID:   "user-1",
				TenantID: "tenant-1",
			},
			refreshToken: models.RefreshToken{
				Token:     "refresh-1",
				UserID:    "user-1",
				TenantID:  "tenant-1",
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				CreatedAt: time.Now(),
			},
			deleteError:               errors.New("delete failed"),
			accessStoreError:          nil,
			refreshStoreError:         errors.New("store failed"),
			wantErr:                   true,
			expectedAccessStoreCalls:  1,
			expectedRefreshStoreCalls: 1,
			expectedDeleteCalls:       1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			accessMock := handlers_mocks.NewMockTokenHandler[auth_models.TokenMetadata](ctrl)
			refreshMock := handlers_mocks.NewMockTokenHandler[models.RefreshToken](ctrl)

			if tc.expectedAccessStoreCalls > 0 {
				accessMock.EXPECT().
					Store(tc.tenantID, tc.userID, tc.accessTokenID, tc.accessTokenMetadata).
					Return(tc.accessStoreError).
					Times(tc.expectedAccessStoreCalls)
			}

			if tc.expectedRefreshStoreCalls > 0 {
				refreshMock.EXPECT().
					Store(tc.tenantID, tc.userID, tc.refreshTokenID, tc.refreshToken).
					Return(tc.refreshStoreError).
					Times(tc.expectedRefreshStoreCalls)
			}
			if tc.expectedDeleteCalls > 0 {
				accessMock.EXPECT().
					Delete(tc.tenantID, tc.userID, tc.accessTokenID).
					Return(tc.deleteError).
					Times(tc.expectedDeleteCalls)
			}

			tm := &TokenManager{
				accessTokenHandler:  accessMock,
				refreshTokenHandler: refreshMock,
				logger:              logging.NewLogger(common_models.ModuleAuth),
			}

			err := tm.StoreTokens(
				tc.tenantID, tc.userID,
				tc.accessTokenID, tc.refreshTokenID,
				tc.accessTokenMetadata, tc.refreshToken,
			)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTokenManager_ValidateAccessToken(t *testing.T) {
	testCases := []struct {
		name                      string
		tenantID                  string
		userID                    string
		tokenID                   string
		returnMetadata            *auth_models.TokenMetadata
		returnError               error
		wantErr                   bool
		expectedValidateCallTimes int
	}{
		{
			name:     "valid token",
			tenantID: "tenant-1",
			userID:   "user-1",
			tokenID:  "token-1",
			returnMetadata: &auth_models.TokenMetadata{
				TokenID:   "token-1",
				TenantID:  "tenant-1",
				UserID:    "user-1",
				Revoked:   false,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			returnError:               nil,
			wantErr:                   false,
			expectedValidateCallTimes: 1,
		},
		{
			name:                      "invalid token",
			tenantID:                  "tenant-1",
			userID:                    "user-1",
			tokenID:                   "token-1",
			returnMetadata:            nil,
			returnError:               erp_errors.Auth(erp_errors.AuthTokenInvalid),
			wantErr:                   true,
			expectedValidateCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := handlers_mocks.NewMockTokenHandler[auth_models.TokenMetadata](ctrl)
			if tc.expectedValidateCallTimes > 0 {
				mock.EXPECT().
					Validate(tc.tenantID, tc.userID, tc.tokenID).
					Return(tc.returnMetadata, tc.returnError).
					Times(tc.expectedValidateCallTimes)
			}

			tm := &TokenManager{
				accessTokenHandler: mock,
				logger:             logging.NewLogger(common_models.ModuleAuth),
			}

			metadata, err := tm.ValidateAccessTokenFromRedis(tc.tenantID, tc.userID, tc.tokenID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, metadata)
			}
		})
	}
}

func TestTokenManager_ValidateRefreshToken(t *testing.T) {
	testCases := []struct {
		name                      string
		tenantID                  string
		userID                    string
		tokenID                   string
		returnToken               *models.RefreshToken
		returnError               error
		wantErr                   bool
		expectedValidateCallTimes int
	}{
		{
			name:     "valid refresh token",
			tenantID: "tenant-1",
			userID:   "user-1",
			tokenID:  "refresh-1",
			returnToken: &models.RefreshToken{
				Token:     "refresh-1",
				UserID:    "user-1",
				TenantID:  "tenant-1",
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				IsRevoked: false,
			},
			returnError:               nil,
			wantErr:                   false,
			expectedValidateCallTimes: 1,
		},
		{
			name:                      "invalid refresh token",
			tenantID:                  "tenant-1",
			userID:                    "user-1",
			tokenID:                   "refresh-1",
			returnToken:               nil,
			returnError:               erp_errors.Auth(erp_errors.AuthTokenInvalid),
			wantErr:                   true,
			expectedValidateCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := handlers_mocks.NewMockTokenHandler[models.RefreshToken](ctrl)
			if tc.expectedValidateCallTimes > 0 {
				mock.EXPECT().
					Validate(tc.tenantID, tc.userID, tc.tokenID).
					Return(tc.returnToken, tc.returnError).
					Times(tc.expectedValidateCallTimes)
			}

			tm := &TokenManager{
				refreshTokenHandler: mock,
				logger:              logging.NewLogger(common_models.ModuleAuth),
			}

			token, err := tm.ValidateRefreshTokenFromRedis(tc.tenantID, tc.userID, tc.tokenID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, token)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, token)
			}
		})
	}
}

func TestTokenManager_RevokeAllTokens(t *testing.T) {
	testCases := []struct {
		name                       string
		tenantID                   string
		userID                     string
		revokedBy                  string
		accessRevokeError          error
		refreshRevokeError         error
		wantErr                    bool
		expectedAccessRevokeCalls  int
		expectedRefreshRevokeCalls int
	}{
		{
			name:                       "successful revoke all",
			tenantID:                   "tenant-1",
			userID:                     "user-1",
			revokedBy:                  "admin",
			accessRevokeError:          nil,
			refreshRevokeError:         nil,
			wantErr:                    false,
			expectedAccessRevokeCalls:  1,
			expectedRefreshRevokeCalls: 1,
		},
		{
			name:                       "refresh token revoke fails",
			tenantID:                   "tenant-1",
			userID:                     "user-1",
			revokedBy:                  "admin",
			accessRevokeError:          nil,
			refreshRevokeError:         errors.New("revoke failed"),
			wantErr:                    true,
			expectedAccessRevokeCalls:  1,
			expectedRefreshRevokeCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			accessMock := handlers_mocks.NewMockTokenHandler[auth_models.TokenMetadata](ctrl)
			refreshMock := handlers_mocks.NewMockTokenHandler[models.RefreshToken](ctrl)

			if tc.expectedAccessRevokeCalls > 0 {
				accessMock.EXPECT().
					RevokeAll(tc.tenantID, tc.userID, tc.revokedBy).
					Return(tc.accessRevokeError).
					Times(tc.expectedAccessRevokeCalls)
			}

			if tc.expectedRefreshRevokeCalls > 0 {
				refreshMock.EXPECT().
					RevokeAll(tc.tenantID, tc.userID, tc.revokedBy).
					Return(tc.refreshRevokeError).
					Times(tc.expectedRefreshRevokeCalls)
			}

			tm := &TokenManager{
				accessTokenHandler:  accessMock,
				refreshTokenHandler: refreshMock,
				logger:              logging.NewLogger(common_models.ModuleAuth),
			}

			err := tm.RevokeAllTokens(tc.tenantID, tc.userID, tc.revokedBy)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
