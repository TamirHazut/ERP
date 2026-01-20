package token

import (
	"errors"
	"testing"
	"time"

	mock_token "erp.localhost/internal/auth/token/mock"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1_cache "erp.localhost/internal/infra/model/auth/v1/cache"
	"erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewTokenManager(t *testing.T) {
	tm := NewTokenManager("test-secret-key-12345", time.Hour, 7*24*time.Hour, &logger.BaseLogger{})
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
		accessTokenMetadata       *authv1_cache.TokenMetadata
		refreshToken              *authv1_cache.RefreshToken
		accessStoreError          error
		refreshStoreError         error
		deleteError               error
		wantErr                   bool
		expectedAccessStoreCalls  int
		expectedRefreshStoreCalls int
		expectedDeleteCalls       int
	}{
		{
			name:     "successful store",
			tenantID: "tenant-1",
			userID:   "user-1",
			accessTokenMetadata: &authv1_cache.TokenMetadata{
				UserId:   "user-1",
				TenantId: "tenant-1",
			},
			refreshToken: &authv1_cache.RefreshToken{
				UserId:    "user-1",
				TenantId:  "tenant-1",
				ExpiresAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
				CreatedAt: timestamppb.Now(),
			},
			accessStoreError:          nil,
			refreshStoreError:         nil,
			wantErr:                   false,
			expectedAccessStoreCalls:  1,
			expectedRefreshStoreCalls: 1,
		},
		{
			name:     "access token store fails",
			tenantID: "tenant-1",
			userID:   "user-1",
			accessTokenMetadata: &authv1_cache.TokenMetadata{
				UserId:   "user-1",
				TenantId: "tenant-1",
			},
			refreshToken: &authv1_cache.RefreshToken{
				UserId:    "user-1",
				TenantId:  "tenant-1",
				ExpiresAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
				CreatedAt: timestamppb.Now(),
			},
			accessStoreError:          errors.New("store failed"),
			refreshStoreError:         nil,
			wantErr:                   true,
			expectedAccessStoreCalls:  1,
			expectedRefreshStoreCalls: 0,
		},
		{
			name:     "refresh token store fails - access token cleaned up",
			tenantID: "tenant-1",
			userID:   "user-1",
			accessTokenMetadata: &authv1_cache.TokenMetadata{
				UserId:   "user-1",
				TenantId: "tenant-1",
			},
			refreshToken: &authv1_cache.RefreshToken{
				UserId:    "user-1",
				TenantId:  "tenant-1",
				ExpiresAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
				CreatedAt: timestamppb.Now(),
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

			accessMock := mock_token.NewMockTokenHandler[authv1_cache.TokenMetadata](ctrl)
			refreshMock := mock_token.NewMockTokenHandler[authv1_cache.RefreshToken](ctrl)

			if tc.expectedAccessStoreCalls > 0 {
				accessMock.EXPECT().
					Store(tc.tenantID, tc.userID, tc.accessTokenMetadata).
					Return(tc.accessStoreError).
					Times(tc.expectedAccessStoreCalls)
			}

			if tc.expectedRefreshStoreCalls > 0 {
				refreshMock.EXPECT().
					Store(tc.tenantID, tc.userID, tc.refreshToken).
					Return(tc.refreshStoreError).
					Times(tc.expectedRefreshStoreCalls)
			}
			if tc.expectedDeleteCalls > 0 {
				accessMock.EXPECT().
					Delete(tc.tenantID, tc.userID).
					Return(tc.deleteError).
					Times(tc.expectedDeleteCalls)
			}

			tm := &TokenManager{
				accessTokenHandler:  accessMock,
				refreshTokenHandler: refreshMock,
				logger:              logger.NewBaseLogger(shared.ModuleAuth),
			}

			err := tm.StoreTokens(
				tc.tenantID, tc.userID,
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
		returnMetadata            *authv1_cache.TokenMetadata
		returnError               error
		wantErr                   bool
		expectedValidateCallTimes int
	}{
		{
			name:     "valid token",
			tenantID: "tenant-1",
			userID:   "user-1",
			returnMetadata: &authv1_cache.TokenMetadata{
				TenantId:  "tenant-1",
				UserId:    "user-1",
				Revoked:   false,
				ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
			},
			returnError:               nil,
			wantErr:                   false,
			expectedValidateCallTimes: 1,
		},
		{
			name:                      "invalid token",
			tenantID:                  "tenant-1",
			userID:                    "user-1",
			returnMetadata:            nil,
			returnError:               infra_error.Auth(infra_error.AuthTokenInvalid),
			wantErr:                   true,
			expectedValidateCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mock_token.NewMockTokenHandler[authv1_cache.TokenMetadata](ctrl)
			if tc.expectedValidateCallTimes > 0 {
				mock.EXPECT().
					Validate(tc.tenantID, tc.userID).
					Return(tc.returnMetadata, tc.returnError).
					Times(tc.expectedValidateCallTimes)
			}

			tm := &TokenManager{
				accessTokenHandler: mock,
				logger:             logger.NewBaseLogger(shared.ModuleAuth),
			}

			metadata, err := tm.ValidateAccessTokenFromRedis(tc.tenantID, tc.userID)

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
		returnToken               *authv1_cache.RefreshToken
		returnError               error
		wantErr                   bool
		expectedValidateCallTimes int
	}{
		{
			name:     "valid refresh token",
			tenantID: "tenant-1",
			userID:   "user-1",
			returnToken: &authv1_cache.RefreshToken{
				UserId:    "user-1",
				TenantId:  "tenant-1",
				ExpiresAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
				Revoked:   false,
			},
			returnError:               nil,
			wantErr:                   false,
			expectedValidateCallTimes: 1,
		},
		{
			name:                      "invalid refresh token",
			tenantID:                  "tenant-1",
			userID:                    "user-1",
			returnToken:               nil,
			returnError:               infra_error.Auth(infra_error.AuthTokenInvalid),
			wantErr:                   true,
			expectedValidateCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mock_token.NewMockTokenHandler[authv1_cache.RefreshToken](ctrl)
			if tc.expectedValidateCallTimes > 0 {
				mock.EXPECT().
					Validate(tc.tenantID, tc.userID).
					Return(tc.returnToken, tc.returnError).
					Times(tc.expectedValidateCallTimes)
			}

			tm := &TokenManager{
				refreshTokenHandler: mock,
				logger:              logger.NewBaseLogger(shared.ModuleAuth),
			}

			token, err := tm.ValidateRefreshTokenFromRedis(tc.tenantID, tc.userID)

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

/*func TestTokenManager_RevokeAllTokens(t *testing.T) {
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

			accessMock := mock_token.NewMockTokenHandler[authv1_cache.TokenMetadata](ctrl)
			refreshMock := mock_token.NewMockTokenHandler[authv1_cache.RefreshToken](ctrl)

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
				logger:              logger.NewBaseLogger(shared.ModuleAuth),
			}

			err := tm.RevokeAllTokens(tc.tenantID, tc.userID, tc.revokedBy)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}*/
