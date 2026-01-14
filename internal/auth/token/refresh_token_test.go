package token

import (
	"errors"
	"testing"
	"time"

	mock_redis "erp.localhost/internal/infra/db/redis/mock"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_shared "erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// refreshTokenMatcher is a custom gomock matcher for RefreshToken objects
// It skips the LastUsedAt field which is set dynamically in UpdateLastUsed operations
type refreshTokenMatcher struct {
	expected model_auth.RefreshToken
}

func (m refreshTokenMatcher) Matches(x interface{}) bool {
	token, ok := x.(model_auth.RefreshToken)
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
	validToken := model_auth.RefreshToken{
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
		refreshToken            model_auth.RefreshToken
		returnGetOneToken       *model_auth.RefreshToken
		returnGetOneError       error
		returnSetError          error
		wantErr                 bool
		expectedSetCallTimes    int
		expectedGetOneCallTimes int
	}{
		{
			name:                    "successful store",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			refreshToken:            validToken,
			returnGetOneToken:       nil,
			returnGetOneError:       nil,
			returnSetError:          nil,
			wantErr:                 false,
			expectedSetCallTimes:    1,
			expectedGetOneCallTimes: 1,
		},
		{
			name:     "store with validation error - missing token",
			tenantID: "tenant-123",
			userID:   "user-123",
			refreshToken: model_auth.RefreshToken{
				UserID:    "user-123",
				TenantID:  "tenant-123",
				ExpiresAt: time.Now().Add(24 * time.Hour),
				CreatedAt: time.Now(),
			},
			returnSetError:       nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:     "store with tenant_id mismatch",
			tenantID: "tenant-123",
			userID:   "user-123",
			refreshToken: model_auth.RefreshToken{
				Token:     "refresh-token-123",
				UserID:    "user-123",
				TenantID:  "wrong-tenant",
				ExpiresAt: time.Now().Add(24 * time.Hour),
				CreatedAt: time.Now(),
			},
			returnSetError:       nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:                    "store with database error",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			refreshToken:            validToken,
			returnGetOneToken:       nil,
			returnGetOneError:       nil,
			returnSetError:          errors.New("database connection failed"),
			wantErr:                 true,
			expectedSetCallTimes:    1,
			expectedGetOneCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[model_auth.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.tenantID, tc.userID).
					Return(tc.returnGetOneToken, tc.returnGetOneError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedSetCallTimes > 0 {
				mockHandler.EXPECT().
					Set(tc.tenantID, tc.userID, tc.refreshToken).
					Return(tc.returnSetError).
					Times(tc.expectedSetCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			err := handler.Store(tc.tenantID, tc.userID, tc.refreshToken)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRefreshTokenKeyHandler_GetOne(t *testing.T) {
	validToken := model_auth.RefreshToken{
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
		returnToken             *model_auth.RefreshToken
		returnError             error
		wantToken               *model_auth.RefreshToken
		wantErr                 bool
		expectedGetOneCallTimes int
	}{
		{
			name:                    "successful get",
			tenantID:                "tenant-123",
			userID:                  "user-123",
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

			mockHandler := mock_redis.NewMockKeyHandler[model_auth.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.tenantID, tc.userID).
					Return(tc.returnToken, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			result, err := handler.GetOne(tc.tenantID, tc.userID)
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

func TestRefreshTokenKeyHandler_Validate(t *testing.T) {
	validToken := model_auth.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
	expiredToken := model_auth.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-48 * time.Hour),
		IsRevoked: false,
	}
	revokedToken := model_auth.RefreshToken{
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
		returnToken             *model_auth.RefreshToken
		returnError             error
		wantErr                 bool
		expectedGetOneCallTimes int
	}{
		{
			name:                    "valid token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			returnToken:             &validToken,
			returnError:             nil,
			wantErr:                 false,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "expired token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			returnToken:             &expiredToken,
			returnError:             nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "revoked token",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			returnToken:             &revokedToken,
			returnError:             nil,
			wantErr:                 true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                    "token not found",
			tenantID:                "tenant-123",
			userID:                  "user-123",
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

			mockHandler := mock_redis.NewMockKeyHandler[model_auth.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.tenantID, tc.userID).
					Return(tc.returnToken, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

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

func TestRefreshTokenKeyHandler_Revoke(t *testing.T) {
	validToken := model_auth.RefreshToken{
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
		returnGetToken          *model_auth.RefreshToken
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
			returnGetToken:          nil,
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

			mockHandler := mock_redis.NewMockKeyHandler[model_auth.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.tenantID, tc.userID).
					Return(tc.returnGetToken, tc.returnGetError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedUpdateCallTimes > 0 {
				// Create expected token with IsRevoked=true
				expectedToken := *tc.returnGetToken
				expectedToken.IsRevoked = true
				mockHandler.EXPECT().
					Update(tc.tenantID, tc.userID, refreshTokenMatcher{expected: expectedToken}).
					Return(tc.returnUpdateError).
					Times(tc.expectedUpdateCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			err := handler.Revoke(tc.tenantID, tc.userID, "system")
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
		returnDeleteError       error
		wantErr                 bool
		expectedDeleteCallTimes int
	}{
		{
			name:                    "successful delete",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			returnDeleteError:       nil,
			wantErr:                 false,
			expectedDeleteCallTimes: 1,
		},
		{
			name:                    "delete with database error",
			tenantID:                "tenant-123",
			userID:                  "user-123",
			returnDeleteError:       errors.New("delete failed"),
			wantErr:                 true,
			expectedDeleteCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[model_auth.RefreshToken](ctrl)
			if tc.expectedDeleteCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.tenantID, tc.userID).
					Return(tc.returnDeleteError).
					Times(tc.expectedDeleteCallTimes)
			}

			logger := logger.NewBaseLogger(model_shared.ModuleAuth)
			handler := NewRefreshTokenHandler(mockHandler, nil, logger)

			err := handler.Delete(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
