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

// refreshTokenMatcher is a custom gomock matcher for RefreshToken objects
// It skips the LastUsedAt field which is set dynamically in UpdateLastUsed operations
type refreshTokenMatcher struct {
	expected *authv1_cache.RefreshToken
}

func (m refreshTokenMatcher) Matches(x interface{}) bool {
	token, ok := x.(*authv1_cache.RefreshToken)
	if !ok {
		return false
	}
	// Match all fields except LastUsedAt which is set dynamically
	return token.UserId == m.expected.UserId &&
		token.TenantId == m.expected.TenantId &&
		token.Revoked == m.expected.Revoked
}

func (m refreshTokenMatcher) String() string {
	return "matches refresh token fields except LastUsedAt"
}

func createNewRefreshTokenHandler(mockHandler *mock_redis.MockKeyHandler[authv1_cache.RefreshToken]) *RefreshTokenHandler {
	handler := &RefreshTokenHandler{
		handler: mockHandler,
		logger:  logger.NewBaseLogger(shared.ModuleAuth),
	}
	return handler
}

func TestRefreshTokenKeyHandler_Store(t *testing.T) {
	validToken := &authv1_cache.RefreshToken{
		TokenHash: "refresh-token-123",
		UserId:    "user-123",
		TenantId:  "tenant-123",
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		CreatedAt: timestamppb.Now(),
		Revoked:   false,
	}

	testCases := []struct {
		name                 string
		tenantID             string
		userID               string
		refreshToken         *authv1_cache.RefreshToken
		returnSetError       error
		wantErr              bool
		expectedSetCallTimes int
	}{
		{
			name:                 "successful store",
			tenantID:             "tenant-123",
			userID:               "user-123",
			refreshToken:         validToken,
			returnSetError:       nil,
			wantErr:              false,
			expectedSetCallTimes: 1,
		},
		{
			name:     "store with validation error - missing token",
			tenantID: "tenant-123",
			userID:   "user-123",
			refreshToken: &authv1_cache.RefreshToken{
				UserId:    "user-123",
				TenantId:  "tenant-123",
				ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
				CreatedAt: timestamppb.Now(),
			},
			returnSetError:       nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:     "store with tenant_id mismatch",
			tenantID: "tenant-123",
			userID:   "user-123",
			refreshToken: &authv1_cache.RefreshToken{
				TokenHash: "refresh-token-123",
				UserId:    "user-123",
				TenantId:  "wrong-tenant",
				ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
				CreatedAt: timestamppb.Now(),
			},
			returnSetError:       nil,
			wantErr:              true,
			expectedSetCallTimes: 0,
		},
		{
			name:                 "store with database error",
			tenantID:             "tenant-123",
			userID:               "user-123",
			refreshToken:         validToken,
			returnSetError:       errors.New("database connection failed"),
			wantErr:              true,
			expectedSetCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.RefreshToken](ctrl)
			if tc.expectedSetCallTimes > 0 {
				mockHandler.EXPECT().
					Set(tc.tenantID, tc.userID, tc.refreshToken, gomock.Any()).
					Return(tc.returnSetError).
					Times(tc.expectedSetCallTimes)
			}

			handler := createNewRefreshTokenHandler(mockHandler)

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
	validToken := authv1_cache.RefreshToken{
		TokenHash: "refresh-token-123",
		UserId:    "user-123",
		TenantId:  "tenant-123",
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		CreatedAt: timestamppb.Now(),
		Revoked:   false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		returnToken             *authv1_cache.RefreshToken
		returnError             error
		wantToken               *authv1_cache.RefreshToken
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

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.tenantID, tc.userID).
					Return(tc.returnToken, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			handler := createNewRefreshTokenHandler(mockHandler)

			result, err := handler.GetOne(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.wantToken.TokenHash, result.TokenHash)
				assert.Equal(t, tc.wantToken.UserId, result.UserId)
			}
		})
	}
}

func TestRefreshTokenKeyHandler_Validate(t *testing.T) {
	validToken := authv1_cache.RefreshToken{
		TokenHash: "refresh-token-123",
		UserId:    "user-123",
		TenantId:  "tenant-123",
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		CreatedAt: timestamppb.Now(),
		Revoked:   false,
	}
	expiredToken := authv1_cache.RefreshToken{
		TokenHash: "refresh-token-123",
		UserId:    "user-123",
		TenantId:  "tenant-123",
		ExpiresAt: timestamppb.New(time.Now().Add(-24 * time.Hour)), // Expired
		CreatedAt: timestamppb.New(time.Now().Add(-48 * time.Hour)),
		Revoked:   false,
	}
	revokedToken := authv1_cache.RefreshToken{
		TokenHash: "refresh-token-123",
		UserId:    "user-123",
		TenantId:  "tenant-123",
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		CreatedAt: timestamppb.Now(),
		Revoked:   true,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		returnToken             *authv1_cache.RefreshToken
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

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.tenantID, tc.userID).
					Return(tc.returnToken, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			handler := createNewRefreshTokenHandler(mockHandler)

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
	validToken := authv1_cache.RefreshToken{
		TokenHash: "refresh-token-123",
		UserId:    "user-123",
		TenantId:  "tenant-123",
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		CreatedAt: timestamppb.Now(),
		Revoked:   false,
	}

	testCases := []struct {
		name                    string
		tenantID                string
		userID                  string
		returnGetToken          *authv1_cache.RefreshToken
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
			returnGetToken:          &validToken,
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
			returnGetToken:          nil,
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
			returnGetToken:          &validToken,
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

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.RefreshToken](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.tenantID, tc.userID).
					Return(tc.returnGetToken, tc.returnGetError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedDeleteCallTimes > 0 {
				// Create expected token with Revoked=true
				expectedToken := tc.returnGetToken
				expectedToken.Revoked = true
				mockHandler.EXPECT().
					Delete(tc.tenantID, tc.userID).
					Return(tc.returnDeleteError).
					Times(tc.expectedDeleteCallTimes)
			}

			handler := createNewRefreshTokenHandler(mockHandler)

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

			mockHandler := mock_redis.NewMockKeyHandler[authv1_cache.RefreshToken](ctrl)
			if tc.expectedDeleteCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.tenantID, tc.userID).
					Return(tc.returnDeleteError).
					Times(tc.expectedDeleteCallTimes)
			}

			handler := createNewRefreshTokenHandler(mockHandler)

			err := handler.Delete(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
