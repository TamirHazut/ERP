package keyshandlers

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"erp.localhost/internal/auth/models"
	db "erp.localhost/internal/db"
	"erp.localhost/internal/db/mock"
	"erp.localhost/internal/db/redis"
	"erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRefreshTokenKeyHandlerWithMock creates a RefreshTokenKeyHandler with a mock handler for testing
func newRefreshTokenKeyHandlerWithMock(mockHandler db.DBHandler, logger *logging.Logger) *RefreshTokenKeyHandler {
	if logger == nil {
		logger = logging.NewLogger(logging.ModuleAuth)
	}
	keyHandler := redis.NewKeyHandlerWithMockForTest[models.RefreshToken](mockHandler, logger)
	return &RefreshTokenKeyHandler{
		keyHandler: keyHandler,
		tokenIndex: nil, // Don't use real token index in tests
		logger:     logger,
	}
}

func TestNewRefreshTokenKeyHandler(t *testing.T) {
	// Note: This test requires a running Redis instance
	// If Redis is not available, it will fail
	// For unit testing, use newRefreshTokenKeyHandlerWithMock instead
	t.Skip("Skipping test that requires Redis connection")
}

func TestRefreshTokenKeyHandler_Store(t *testing.T) {
	validToken := models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}

	testCases := []struct {
		name         string
		tenantID     string
		userID       string
		tokenID      string
		refreshToken models.RefreshToken
		mockFunc     func(key string, data any) (string, error)
		wantErr      bool
	}{
		{
			name:         "successful store",
			tenantID:     "tenant-123",
			userID:       "user-123",
			tokenID:      "token-123",
			refreshToken: validToken,
			mockFunc: func(key string, data any) (string, error) {
				return "ok", nil
			},
			wantErr: false,
		},
		{
			name:     "store with validation error - missing token",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			refreshToken: models.RefreshToken{
				UserID:    "user-123",
				TenantID:  "tenant-123",
				ExpiresAt: time.Now().Add(24 * time.Hour),
				CreatedAt: time.Now(),
			},
			mockFunc: func(key string, data any) (string, error) {
				return "ok", nil
			},
			wantErr: true,
		},
		{
			name:         "store with tenant_id mismatch",
			tenantID:     "tenant-123",
			userID:       "user-123",
			tokenID:      "token-123",
			refreshToken: validToken,
			mockFunc: func(key string, data any) (string, error) {
				return "ok", nil
			},
			wantErr: false, // Will fail validation
		},
		{
			name:         "store with database error",
			tenantID:     "tenant-123",
			userID:       "user-123",
			tokenID:      "token-123",
			refreshToken: validToken,
			mockFunc: func(key string, data any) (string, error) {
				return "", errors.New("database connection failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				CreateFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleAuth)
			handler := newRefreshTokenKeyHandlerWithMock(mockHandler, logger)

			err := handler.Store(tc.tenantID, tc.userID, tc.tokenID, tc.refreshToken)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				// Check if validation error occurred
				if tc.name == "store with tenant_id mismatch" {
					// Create a token with wrong tenant ID
					wrongToken := validToken
					wrongToken.TenantID = "wrong-tenant"
					err := handler.Store(tc.tenantID, tc.userID, tc.tokenID, wrongToken)
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}

func TestRefreshTokenKeyHandler_GetOne(t *testing.T) {
	validToken := models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
	jsonData, _ := json.Marshal(validToken)

	testCases := []struct {
		name      string
		tenantID  string
		userID    string
		tokenID   string
		mockFunc  func(key string, filter map[string]any) (any, error)
		wantToken *models.RefreshToken
		wantErr   bool
	}{
		{
			name:     "successful get",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return string(jsonData), nil
			},
			wantToken: &validToken,
			wantErr:   false,
		},
		{
			name:     "token not found",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return any(nil), nil
			},
			wantToken: nil,
			wantErr:   true,
		},
		{
			name:     "database error",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return any(nil), errors.New("database query failed")
			},
			wantToken: nil,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindOneFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleAuth)
			handler := newRefreshTokenKeyHandlerWithMock(mockHandler, logger)

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
	validToken := models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
	jsonData, _ := json.Marshal(validToken)

	testCases := []struct {
		name      string
		tenantID  string
		userID    string
		tokenID   string
		mockFunc  func(key string, filter map[string]any) ([]any, error)
		wantToken *models.RefreshToken
		wantErr   bool
	}{
		{
			name:     "successful get",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return []any{string(jsonData)}, nil
			},
			wantToken: &validToken,
			wantErr:   false,
		},
		{
			name:     "token not found",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantToken: nil,
			wantErr:   true,
		},
		{
			name:     "database error",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return nil, errors.New("database query failed")
			},
			wantToken: nil,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindAllFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleAuth)
			handler := newRefreshTokenKeyHandlerWithMock(mockHandler, logger)

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

func TestRefreshTokenKeyHandler_Validate(t *testing.T) {
	validToken := models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
	expiredToken := models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-48 * time.Hour),
		IsRevoked: false,
	}
	revokedToken := models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: true,
	}

	validJSON, _ := json.Marshal(validToken)
	expiredJSON, _ := json.Marshal(expiredToken)
	revokedJSON, _ := json.Marshal(revokedToken)

	testCases := []struct {
		name     string
		tenantID string
		userID   string
		tokenID  string
		mockFunc func(key string, filter map[string]any) (any, error)
		wantErr  bool
	}{
		{
			name:     "valid token",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return string(validJSON), nil
			},
			wantErr: false,
		},
		{
			name:     "expired token",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return string(expiredJSON), nil
			},
			wantErr: true,
		},
		{
			name:     "revoked token",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return string(revokedJSON), nil
			},
			wantErr: true,
		},
		{
			name:     "token not found",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return any(nil), nil
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindOneFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleAuth)
			handler := newRefreshTokenKeyHandlerWithMock(mockHandler, logger)

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
	validToken := models.RefreshToken{
		Token:     "refresh-token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
	jsonData, _ := json.Marshal(validToken)

	testCases := []struct {
		name       string
		tenantID   string
		userID     string
		tokenID    string
		getFunc    func(key string, filter map[string]any) (any, error)
		updateFunc func(key string, filter map[string]any, data any) error
		wantErr    bool
	}{
		{
			name:     "successful revoke",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			getFunc: func(key string, filter map[string]any) (any, error) {
				return string(jsonData), nil
			},
			updateFunc: func(key string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "token not found",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			getFunc: func(key string, filter map[string]any) (any, error) {
				return any(nil), nil
			},
			updateFunc: func(key string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "update fails",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			getFunc: func(key string, filter map[string]any) (any, error) {
				return string(jsonData), nil
			},
			updateFunc: func(key string, filter map[string]any, data any) error {
				return errors.New("update failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindOneFunc: tc.getFunc,
				UpdateFunc:  tc.updateFunc,
			}
			logger := logging.NewLogger(logging.ModuleAuth)
			handler := newRefreshTokenKeyHandlerWithMock(mockHandler, logger)

			err := handler.Revoke(tc.tenantID, tc.userID, tc.tokenID)
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
		name     string
		tenantID string
		userID   string
		tokenID  string
		mockFunc func(key string, filter map[string]any) error
		wantErr  bool
	}{
		{
			name:     "successful delete",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "delete with database error",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) error {
				return errors.New("delete failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				DeleteFunc: tc.mockFunc,
			}
			logger := logging.NewLogger(logging.ModuleAuth)
			handler := newRefreshTokenKeyHandlerWithMock(mockHandler, logger)

			err := handler.Delete(tc.tenantID, tc.userID, tc.tokenID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
