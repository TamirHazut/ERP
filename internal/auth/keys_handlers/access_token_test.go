package keyshandlers

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	db "erp.localhost/internal/db"
	"erp.localhost/internal/db/mock"
	"erp.localhost/internal/db/redis"
	redis_models "erp.localhost/internal/db/redis/models"
	"erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAccessTokenKeyHandlerWithMock creates an AccessTokenKeyHandler with a mock handler for testing
func newAccessTokenKeyHandlerWithMock(mockHandler db.DBHandler, logger *logging.Logger) *AccessTokenKeyHandler {
	if logger == nil {
		logger = logging.NewLogger(logging.ModuleAuth)
	}
	keyHandler := redis.NewKeyHandlerWithMockForTest[redis_models.TokenMetadata](mockHandler, logger)
	return &AccessTokenKeyHandler{
		keyHandler: keyHandler,
		tokenIndex: nil, // Don't use real token index in tests
		logger:     logger,
	}
}

func TestNewAccessTokenKeyHandler(t *testing.T) {
	// Note: This test requires a running Redis instance
	t.Skip("Skipping test that requires Redis connection")
}

func TestAccessTokenKeyHandler_Store(t *testing.T) {
	validMetadata := redis_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name     string
		tenantID string
		userID   string
		tokenID  string
		metadata redis_models.TokenMetadata
		mockFunc func(key string, data any) (string, error)
		wantErr  bool
	}{
		{
			name:     "successful store",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			metadata: validMetadata,
			mockFunc: func(key string, data any) (string, error) {
				return "ok", nil
			},
			wantErr: false,
		},
		{
			name:     "store with missing tokenID",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			metadata: redis_models.TokenMetadata{
				UserID:    "user-123",
				TenantID:  "tenant-123",
				TokenType: "access",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			mockFunc: func(key string, data any) (string, error) {
				return "ok", nil
			},
			wantErr: true,
		},
		{
			name:     "store with tenant_id mismatch",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			metadata: redis_models.TokenMetadata{
				TokenID:   "token-123",
				UserID:    "user-123",
				TenantID:  "wrong-tenant",
				TokenType: "access",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			mockFunc: func(key string, data any) (string, error) {
				return "ok", nil
			},
			wantErr: true,
		},
		{
			name:     "store with database error",
			tenantID: "tenant-123",
			tokenID:  "token-123",
			metadata: validMetadata,
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
			handler := newAccessTokenKeyHandlerWithMock(mockHandler, logger)

			err := handler.Store(tc.tenantID, tc.userID, tc.tokenID, tc.metadata)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccessTokenKeyHandler_GetOne(t *testing.T) {
	validMetadata := redis_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name      string
		tenantID  string
		userID    string
		tokenID   string
		mockFunc  func(db string, filter map[string]any) (any, error)
		wantToken *redis_models.TokenMetadata
		wantErr   bool
	}{
		{
			name:     "successful get",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(db string, filter map[string]any) (any, error) {
				return validMetadata, nil
			},
			wantToken: &validMetadata,
			wantErr:   false,
		},
		{
			name:     "token not found",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(db string, filter map[string]any) (any, error) {
				return redis_models.TokenMetadata{}, errors.New("token not found")
			},
			wantToken: nil,
			wantErr:   true,
		},
		{
			name:     "database error",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(db string, filter map[string]any) (any, error) {
				return redis_models.TokenMetadata{}, errors.New("database query failed")
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
			handler := newAccessTokenKeyHandlerWithMock(mockHandler, logger)

			result, err := handler.GetOne(tc.tenantID, tc.userID, tc.tokenID)
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

func TestAccessTokenKeyHandler_GetAll(t *testing.T) {
	validMetadata := redis_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}
	jsonData, _ := json.Marshal(validMetadata)
	testCases := []struct {
		name      string
		tenantID  string
		userID    string
		mockFunc  func(key string, filter map[string]any) ([]any, error)
		wantToken *redis_models.TokenMetadata
		wantErr   bool
	}{
		{
			name:     "successful get",
			tenantID: "tenant-123",
			userID:   "user-123",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return []any{string(jsonData)}, nil
			},
			wantToken: &validMetadata,
			wantErr:   false,
		},
		{
			name:     "token not found",
			tenantID: "tenant-123",
			userID:   "user-123",
			mockFunc: func(key string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantToken: nil,
			wantErr:   false,
		},
		{
			name:     "database error",
			tenantID: "tenant-123",
			userID:   "user-123",
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
			handler := newAccessTokenKeyHandlerWithMock(mockHandler, logger)

			tokens, err := handler.GetAll(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, tokens)
			} else {
				require.NoError(t, err)
				if tc.wantToken == nil {
					assert.Empty(t, tokens)
				} else {
					require.NotEmpty(t, tokens)
					token := tokens[0]
					assert.Equal(t, tc.wantToken.TokenID, token.TokenID)
					assert.Equal(t, tc.wantToken.UserID, token.UserID)
				}
			}
		})
	}
}

func TestAccessTokenKeyHandler_Validate(t *testing.T) {
	validMetadata := redis_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}
	expiredMetadata := redis_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour), // Expired
		Revoked:   false,
	}
	revokedMetadata := redis_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   true,
	}

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
				return validMetadata, nil
			},
			wantErr: false,
		},
		{
			name:     "expired token",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return expiredMetadata, nil
			},
			wantErr: true,
		},
		{
			name:     "revoked token",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return revokedMetadata, nil
			},
			wantErr: true,
		},
		{
			name:     "token not found",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			mockFunc: func(key string, filter map[string]any) (any, error) {
				return nil, errors.New("token not found")
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
			handler := newAccessTokenKeyHandlerWithMock(mockHandler, logger)

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

func TestAccessTokenKeyHandler_Revoke(t *testing.T) {
	validMetadata := redis_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name       string
		tenantID   string
		userID     string
		tokenID    string
		revokedBy  string
		getFunc    func(key string, filter map[string]any) (any, error)
		updateFunc func(key string, filter map[string]any, data any) error
		wantErr    bool
	}{
		{
			name:      "successful revoke",
			tenantID:  "tenant-123",
			userID:    "user-123",
			tokenID:   "token-123",
			revokedBy: "admin",
			getFunc: func(key string, filter map[string]any) (any, error) {
				return validMetadata, nil
			},
			updateFunc: func(key string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:      "token not found",
			tenantID:  "tenant-123",
			userID:    "user-123",
			tokenID:   "token-123",
			revokedBy: "admin",
			getFunc: func(key string, filter map[string]any) (any, error) {
				return nil, errors.New("token not found")
			},
			updateFunc: func(key string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:      "update fails",
			tenantID:  "tenant-123",
			userID:    "user-123",
			tokenID:   "token-123",
			revokedBy: "admin",
			getFunc: func(key string, filter map[string]any) (any, error) {
				return validMetadata, nil
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
			handler := newAccessTokenKeyHandlerWithMock(mockHandler, logger)

			err := handler.Revoke(tc.tenantID, tc.userID, tc.tokenID, tc.revokedBy)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccessTokenKeyHandler_Delete(t *testing.T) {
	validMetadata := redis_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name       string
		tenantID   string
		userID     string
		tokenID    string
		getFunc    func(key string, filter map[string]any) (any, error)
		deleteFunc func(key string, filter map[string]any) error
		wantErr    bool
	}{
		{
			name:     "successful delete",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			getFunc: func(key string, filter map[string]any) (any, error) {
				return validMetadata, nil
			},
			deleteFunc: func(key string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "delete with database error",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			getFunc: func(key string, filter map[string]any) (any, error) {
				return validMetadata, nil
			},
			deleteFunc: func(key string, filter map[string]any) error {
				return errors.New("delete failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindOneFunc: tc.getFunc,
				DeleteFunc:  tc.deleteFunc,
			}
			logger := logging.NewLogger(logging.ModuleAuth)
			handler := newAccessTokenKeyHandlerWithMock(mockHandler, logger)

			err := handler.Delete(tc.tenantID, tc.userID, tc.tokenID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
