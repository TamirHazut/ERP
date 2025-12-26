package auth

import (
	"errors"
	"testing"
	"time"

	"erp.localhost/internal/auth/mocks"
	"erp.localhost/internal/auth/models"
	redis_models "erp.localhost/internal/db/redis/models"
	erp_errors "erp.localhost/internal/errors"
	logging "erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		name                string
		setupMocks          func() (AccessTokenHandler, RefreshTokenHandler)
		tenantID            string
		userID              string
		accessTokenID       string
		refreshTokenID      string
		accessTokenMetadata redis_models.TokenMetadata
		refreshToken        models.RefreshToken
		wantErr             bool
	}{
		{
			name: "successful store",
			setupMocks: func() (AccessTokenHandler, RefreshTokenHandler) {
				var accessMock AccessTokenHandler = &mocks.MockAccessTokenKeyHandler{
					StoreFunc: func(tenantID, tokenID string, metadata redis_models.TokenMetadata) error {
						return nil
					},
				}
				var refreshMock RefreshTokenHandler = &mocks.MockRefreshTokenKeyHandler{
					StoreFunc: func(tenantID, userID, tokenID string, refreshToken models.RefreshToken) error {
						return nil
					},
				}
				return accessMock, refreshMock
			},
			tenantID:       "tenant-1",
			userID:         "user-1",
			accessTokenID:  "token-1",
			refreshTokenID: "refresh-1",
			accessTokenMetadata: redis_models.TokenMetadata{
				TokenID:   "token-1",
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
			wantErr: false,
		},
		{
			name: "access token store fails",
			setupMocks: func() (AccessTokenHandler, RefreshTokenHandler) {
				var accessMock AccessTokenHandler = &mocks.MockAccessTokenKeyHandler{
					StoreFunc: func(tenantID, tokenID string, metadata redis_models.TokenMetadata) error {
						return errors.New("store failed")
					},
				}
				var refreshMock RefreshTokenHandler = &mocks.MockRefreshTokenKeyHandler{}
				return accessMock, refreshMock
			},
			tenantID:       "tenant-1",
			userID:         "user-1",
			accessTokenID:  "token-1",
			refreshTokenID: "refresh-1",
			accessTokenMetadata: redis_models.TokenMetadata{
				TokenID:   "token-1",
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
			wantErr: true,
		},
		{
			name: "refresh token store fails - access token cleaned up",
			setupMocks: func() (AccessTokenHandler, RefreshTokenHandler) {
				var accessMock AccessTokenHandler = &mocks.MockAccessTokenKeyHandler{
					StoreFunc: func(tenantID, tokenID string, metadata redis_models.TokenMetadata) error {
						return nil
					},
					DeleteFunc: func(tenantID, tokenID string) error {
						return nil
					},
				}
				var refreshMock RefreshTokenHandler = &mocks.MockRefreshTokenKeyHandler{
					StoreFunc: func(tenantID, userID, tokenID string, refreshToken models.RefreshToken) error {
						return errors.New("refresh store failed")
					},
				}
				return accessMock, refreshMock
			},
			tenantID:       "tenant-1",
			userID:         "user-1",
			accessTokenID:  "token-1",
			refreshTokenID: "refresh-1",
			accessTokenMetadata: redis_models.TokenMetadata{
				TokenID:   "token-1",
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
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accessMock, refreshMock := tc.setupMocks()
			tm := &TokenManager{
				accessTokenHandler:  accessMock,
				refreshTokenHandler: refreshMock,
				logger:              logging.NewLogger(logging.ModuleAuth),
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
		name      string
		setupMock func() *mocks.MockAccessTokenKeyHandler
		tenantID  string
		tokenID   string
		wantErr   bool
	}{
		{
			name: "valid token",
			setupMock: func() *mocks.MockAccessTokenKeyHandler {
				return &mocks.MockAccessTokenKeyHandler{
					ValidateFunc: func(tenantID, tokenID string) (*redis_models.TokenMetadata, error) {
						return &redis_models.TokenMetadata{
							TokenID:   tokenID,
							TenantID:  tenantID,
							UserID:    "user-1",
							Revoked:   false,
							ExpiresAt: time.Now().Add(time.Hour),
						}, nil
					},
				}
			},
			tenantID: "tenant-1",
			tokenID:  "token-1",
			wantErr:  false,
		},
		{
			name: "invalid token",
			setupMock: func() *mocks.MockAccessTokenKeyHandler {
				return &mocks.MockAccessTokenKeyHandler{
					ValidateFunc: func(tenantID, tokenID string) (*redis_models.TokenMetadata, error) {
						return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid)
					},
				}
			},
			tenantID: "tenant-1",
			tokenID:  "token-1",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := tc.setupMock()
			tm := &TokenManager{
				accessTokenHandler: mock,
				logger:            logging.NewLogger(logging.ModuleAuth),
			}

			metadata, err := tm.ValidateAccessTokenFromRedis(tc.tenantID, tc.tokenID)

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
		name      string
		setupMock func() *mocks.MockRefreshTokenKeyHandler
		tenantID  string
		userID    string
		tokenID   string
		wantErr   bool
	}{
		{
			name: "valid refresh token",
			setupMock: func() *mocks.MockRefreshTokenKeyHandler {
				return &mocks.MockRefreshTokenKeyHandler{
					ValidateFunc: func(tenantID, userID, tokenID string) (*models.RefreshToken, error) {
						return &models.RefreshToken{
							Token:     tokenID,
							UserID:    userID,
							TenantID:  tenantID,
							ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
							IsRevoked: false,
						}, nil
					},
				}
			},
			tenantID: "tenant-1",
			userID:   "user-1",
			tokenID:  "refresh-1",
			wantErr:  false,
		},
		{
			name: "invalid refresh token",
			setupMock: func() *mocks.MockRefreshTokenKeyHandler {
				return &mocks.MockRefreshTokenKeyHandler{
					ValidateFunc: func(tenantID, userID, tokenID string) (*models.RefreshToken, error) {
						return nil, erp_errors.Auth(erp_errors.AuthTokenInvalid)
					},
				}
			},
			tenantID: "tenant-1",
			userID:   "user-1",
			tokenID:  "refresh-1",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := tc.setupMock()
			tm := &TokenManager{
				refreshTokenHandler: mock,
				logger:              logging.NewLogger(logging.ModuleAuth),
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
		name       string
		setupMocks func() (*mocks.MockAccessTokenKeyHandler, *mocks.MockRefreshTokenKeyHandler)
		tenantID   string
		userID     string
		revokedBy  string
		wantErr    bool
	}{
		{
			name: "successful revoke all",
			setupMocks: func() (*mocks.MockAccessTokenKeyHandler, *mocks.MockRefreshTokenKeyHandler) {
				accessMock := &mocks.MockAccessTokenKeyHandler{
					RevokeAllFunc: func(tenantID, userID, revokedBy string) error {
						return nil
					},
				}
				refreshMock := &mocks.MockRefreshTokenKeyHandler{
					RevokeAllFunc: func(tenantID, userID string) error {
						return nil
					},
				}
				return accessMock, refreshMock
			},
			tenantID:  "tenant-1",
			userID:    "user-1",
			revokedBy: "admin",
			wantErr:   false,
		},
		{
			name: "refresh token revoke fails",
			setupMocks: func() (*mocks.MockAccessTokenKeyHandler, *mocks.MockRefreshTokenKeyHandler) {
				accessMock := &mocks.MockAccessTokenKeyHandler{
					RevokeAllFunc: func(tenantID, userID, revokedBy string) error {
						return nil
					},
				}
				refreshMock := &mocks.MockRefreshTokenKeyHandler{
					RevokeAllFunc: func(tenantID, userID string) error {
						return errors.New("revoke failed")
					},
				}
				return accessMock, refreshMock
			},
			tenantID:  "tenant-1",
			userID:    "user-1",
			revokedBy: "admin",
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accessMock, refreshMock := tc.setupMocks()
			tm := &TokenManager{
				accessTokenHandler:  accessMock,
				refreshTokenHandler: refreshMock,
				logger:              logging.NewLogger(logging.ModuleAuth),
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

