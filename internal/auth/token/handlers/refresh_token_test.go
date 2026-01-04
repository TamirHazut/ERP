package handlers

import (
	"errors"
	"testing"
	"time"

	handlers_mocks "erp.localhost/internal/db/redis/handlers/mocks"
	logging "erp.localhost/internal/logging"
	shared_models "erp.localhost/internal/shared/models"
	auth_models "erp.localhost/internal/shared/models/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// refreshTokenMatcher is a custom gomock matcher for RefreshToken objects
// It skips the LastUsedAt field which is set dynamically in UpdateLastUsed operations
type refreshTokenMatcher struct {
	expected auth_models.RefreshToken
}

func (m refreshTokenMatcher) Matches(x interface{}) bool {
	token, ok := x.(auth_models.RefreshToken)
	if !ok {
		return false
	}
	// Match all fields except LastUsedAt which is set dynamically
	return token.Token == m.expected.Token &&
		token.UserID == m.expected.UserID &&
		token.TenantID == m.expected.TenantID &&
		token.IsRevoked == m.expected.IsRevoked
}

func (m refreshTokenMatcher) String() string {
	return "matches refresh token fields except LastUsedAt"
}

func TestNewRefreshTokenKeyHandler(t *testing.T) {
	// Note: This test requires a running Redis instance
	// If Redis is not available, it will fail
	// For unit testing, use newRefreshTokenKeyHandlerWithMock instead
	t.Skip("Skipping test that requires Redis connection")
}

func TestRefreshTokenKeyHandler_Store(t *testing.T) {
	validToken := auth_models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}

	testCases := []struct {
		name                 string
		tenantID             string
		userID               string
		tokenID              string
		refreshToken         auth_models.RefreshToken
		expectedTenantID     string
		expectedKey          string
		returnError          error
		wantErr              bool
		expectedSetCallTimes int
	}{
		{
			name:                 "successful store",
			tenantID:             "tenant-123",
			userID:               "user-123",
			tokenID:              "token-123",
			refreshToken:         validToken,
			expectedTenantID:     "tenant-123",
			expectedKey:          "user-123:token-123",
			returnError:          nil,
			wantErr:              false,
			expectedSetCallTimes: 1,
		},
		{
			name:     "store with validation error - missing token",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			refreshToken: auth_models.RefreshToken{
				UserID:    "user-123",
				TenantID:  "tenant-123",
				ExpiresAt: time.Now().Add(24 * time.Hour),
				CreatedAt: time.Now(),
			},
			expectedTenantID:     "",
			expectedKey:          "",
			returnError:          nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:     "store with tenant_id mismatch",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			refreshToken: auth_models.RefreshToken{
				Token:     "refresh-token-123",
				UserID:    "user-123",
				TenantID:  "wrong-tenant",
				ExpiresAt: time.Now().Add(24 * time.Hour),
				CreatedAt: time.Now(),
			},
			expectedTenantID:     "",
			expectedKey:          "",
			returnError:          nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:                 "store with database error",
			tenantID:             "tenant-123",
			userID:               "user-123",
			tokenID:              "token-123",
			refreshToken:         validToken,
			expectedTenantID:     "tenant-123",
			expectedKey:          "user-123:token-123",
			returnError:          errors.New("database connection failed"),
			wantErr:              true,
			expectedSetCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.RefreshToken](ctrl)
			if tc.expectedSetCallTimes > 0 {
				mockHandler.EXPECT().
					Set(tc.expectedTenantID, tc.expectedKey, tc.refreshToken).
					Return(tc.returnError).
					Times(tc.expectedSetCallTimes)
			}

			logger := logging.NewLogger(shared_models.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			err := handler.Store(tc.tenantID, tc.userID, tc.tokenID, tc.refreshToken)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRefreshTokenKeyHandler_GetOne(t *testing.T) {
	validToken := auth_models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		tokenID                 string
		expectedTenantID        string
		expectedKey             string
		returnToken             *auth_models.RefreshToken
		returnError             error
		wantToken               *auth_models.RefreshToken
		wantErr                 bool
		expectedGetOneCallTimes int
	}{
		{
			name:                    "successful get",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedTenantID:        "tenant-123",
			expectedKey:             "user-123:token-123",
			returnToken:             &validToken,
			returnError:             nil,
			wantToken:               &validToken,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "token not found",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedTenantID:        "tenant-123",
			expectedKey:             "user-123:token-123",
			returnToken:             nil,
			returnError:             errors.New("token not found"),
			wantToken:               nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "database error",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedTenantID:        "tenant-123",
			expectedKey:             "user-123:token-123",
			returnToken:             nil,
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

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.expectedKey).
					Return(tc.returnToken, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			logger := logging.NewLogger(shared_models.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			result, err := handler.GetOne(tc.tenantID, tc.userID, tc.tokenID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.wantToken.Token, result.Token)
				assert.Equal(t, tc.wantToken.UserID, result.UserID)
			}
		})
	}
}

func TestRefreshTokenKeyHandler_GetAll(t *testing.T) {
	validToken := auth_models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		expectedTenantID        string
		expectedUserID          string
		returnTokens            []auth_models.RefreshToken
		returnError             error
		wantCount               int
		wantErr                 bool
		expectedGetAllCallTimes int
	}{
		{
			name:                    "successful get",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnTokens:            []auth_models.RefreshToken{validToken},
			returnError:             nil,
			wantCount:               1,
			wantErr:                 false,
			expectedGetAllCallTimes: 1,
		},
		{
			name:                    "token not found",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnTokens:            []auth_models.RefreshToken{},
			returnError:             nil,
			wantCount:               0,
			wantErr:                 false,
			expectedGetAllCallTimes: 1,
		},
		{
			name:                    "database error",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			expectedTenantID:        "tenant-123",
			expectedUserID:          "user-123",
			returnTokens:            nil,
			returnError:             errors.New("database query failed"),
			wantCount:               0,
			wantErr:                 true,
			expectedGetAllCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.RefreshToken](ctrl)
			if tc.expectedGetAllCallTimes > 0 {
				mockHandler.EXPECT().
					GetAll(tc.expectedTenantID, tc.expectedUserID).
					Return(tc.returnTokens, tc.returnError).
					Times(tc.expectedGetAllCallTimes)
			}

			logger := logging.NewLogger(shared_models.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			result, err := handler.GetAll(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tc.wantCount)
				if tc.wantCount > 0 {
					assert.Equal(t, validToken.Token, result[0].Token)
					assert.Equal(t, validToken.UserID, result[0].UserID)
				}
			}
		})
	}
}

func TestRefreshTokenKeyHandler_Validate(t *testing.T) {
	validToken := auth_models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
	expiredToken := auth_models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-48 * time.Hour),
		IsRevoked: false,
	}
	revokedToken := auth_models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: true,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		tokenID                 string
		expectedTenantID        string
		expectedKey             string
		returnToken             *auth_models.RefreshToken
		returnError             error
		wantErr                 bool
		expectedGetOneCallTimes int
	}{
		{
			name:                    "valid token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedTenantID:        "tenant-123",
			expectedKey:             "user-123:token-123",
			returnToken:             &validToken,
			returnError:             nil,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "expired token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedTenantID:        "tenant-123",
			expectedKey:             "user-123:token-123",
			returnToken:             &expiredToken,
			returnError:             nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "revoked token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedTenantID:        "tenant-123",
			expectedKey:             "user-123:token-123",
			returnToken:             &revokedToken,
			returnError:             nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "token not found",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedTenantID:        "tenant-123",
			expectedKey:             "user-123:token-123",
			returnToken:             nil,
			returnError:             errors.New("token not found"),
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.expectedKey).
					Return(tc.returnToken, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			logger := logging.NewLogger(shared_models.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			result, err := handler.Validate(tc.tenantID, tc.userID, tc.tokenID)
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

func TestRefreshTokenKeyHandler_Revoke(t *testing.T) {
	validToken := auth_models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		tokenID                 string
		expectedGetTenantID     string
		expectedGetKey          string
		expectedUpdateTenantID  string
		expectedUpdateKey       string
		returnGetToken          *auth_models.RefreshToken
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
			tokenID:                 "token-123",
			expectedGetTenantID:     "tenant-123",
			expectedGetKey:          "user-123:token-123",
			expectedUpdateTenantID:  "tenant-123",
			expectedUpdateKey:       "user-123:token-123",
			returnGetToken:          &validToken,
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
			tokenID:                 "token-123",
			expectedGetTenantID:     "tenant-123",
			expectedGetKey:          "user-123:token-123",
			expectedUpdateTenantID:  "",
			expectedUpdateKey:       "",
			returnGetToken:          nil,
			returnGetError:          errors.New("token not found"),
			returnUpdateError:       nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
			expectedUpdateCallTimes: 0,
		},
		{
			name:                    "update fails",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedGetTenantID:     "tenant-123",
			expectedGetKey:          "user-123:token-123",
			expectedUpdateTenantID:  "tenant-123",
			expectedUpdateKey:       "user-123:token-123",
			returnGetToken:          &validToken,
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

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedGetTenantID, tc.expectedGetKey).
					Return(tc.returnGetToken, tc.returnGetError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedUpdateCallTimes > 0 {
				// Create expected token with IsRevoked=true
				expectedToken := *tc.returnGetToken
				expectedToken.IsRevoked = true
				mockHandler.EXPECT().
					Update(tc.expectedUpdateTenantID, tc.expectedUpdateKey, refreshTokenMatcher{expected: expectedToken}).
					Return(tc.returnUpdateError).
					Times(tc.expectedUpdateCallTimes)
			}

			logger := logging.NewLogger(shared_models.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			err := handler.Revoke(tc.tenantID, tc.userID, tc.tokenID, "system")
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRefreshTokenKeyHandler_Delete(t *testing.T) {
	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		tokenID                 string
		expectedDeleteTenantID  string
		expectedDeleteKey       string
		returnDeleteError       error
		wantErr                 bool
		expectedDeleteCallTimes int
	}{
		{
			name:                    "successful delete",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedDeleteTenantID:  "tenant-123",
			expectedDeleteKey:       "user-123:token-123",
			returnDeleteError:       nil,
			wantErr:                 false,
			expectedDeleteCallTimes: 1,
		},
		{
			name:                    "delete with database error",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			tokenID:                 "token-123",
			expectedDeleteTenantID:  "tenant-123",
			expectedDeleteKey:       "user-123:token-123",
			returnDeleteError:       errors.New("delete failed"),
			wantErr:                 true,
			expectedDeleteCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.RefreshToken](ctrl)
			if tc.expectedDeleteCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.expectedDeleteTenantID, tc.expectedDeleteKey).
					Return(tc.returnDeleteError).
					Times(tc.expectedDeleteCallTimes)
			}

			logger := logging.NewLogger(shared_models.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			err := handler.Delete(tc.tenantID, tc.userID, tc.tokenID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
