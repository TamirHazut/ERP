package handlers

import (
	"errors"
	"testing"
	"time"

	auth_models "erp.localhost/internal/auth/models/cache"
	handlers_mocks "erp.localhost/internal/db/redis/handlers/mocks"
	"erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// tokenMetadataMatcher is a custom gomock matcher for TokenMetadata objects
// It skips the RevokedAt field which is set dynamically in Revoke operations
type tokenMetadataMatcher struct {
	expected auth_models.TokenMetadata
}

func (m tokenMetadataMatcher) Matches(x interface{}) bool {
	metadata, ok := x.(auth_models.TokenMetadata)
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
	validMetadata := auth_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name                string
		tenantID            string
		userID              string
		tokenID             string
		metadata            auth_models.TokenMetadata
		expectedTenantID    string
		expectedTokenID     string
		returnError         error
		wantErr             bool
		expectedSetCallTimes int
	}{
		{
			name:                "successful store",
			tenantID:            "tenant-123",
			userID:              "user-123",
			tokenID:             "token-123",
			metadata:            validMetadata,
			expectedTenantID:    "tenant-123",
			expectedTokenID:     "token-123",
			returnError:         nil,
			wantErr:             false,
			expectedSetCallTimes: 1,
		},
		{
			name:     "store with missing tokenID",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			metadata: auth_models.TokenMetadata{
				UserID:    "user-123",
				TenantID:  "tenant-123",
				TokenType: "access",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectedTenantID:    "",
			expectedTokenID:     "",
			returnError:         nil,
			wantErr:             true,
			expectedSetCallTimes: 0,
		},
		{
			name:     "store with tenant_id mismatch",
			tenantID: "tenant-123",
			userID:   "user-123",
			tokenID:  "token-123",
			metadata: auth_models.TokenMetadata{
				TokenID:   "token-123",
				UserID:    "user-123",
				TenantID:  "wrong-tenant",
				TokenType: "access",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expectedTenantID:    "",
			expectedTokenID:     "",
			returnError:         nil,
			wantErr:             true,
			expectedSetCallTimes: 0,
		},
		{
			name:                "store with database error",
			tenantID:            "tenant-123",
			userID:              "user-123",
			tokenID:             "token-123",
			metadata:            validMetadata,
			expectedTenantID:    "tenant-123",
			expectedTokenID:     "token-123",
			returnError:         errors.New("database connection failed"),
			wantErr:             true,
			expectedSetCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.TokenMetadata](ctrl)
			if tc.expectedSetCallTimes > 0 {
				mockHandler.EXPECT().
					Set(tc.expectedTenantID, tc.expectedTokenID, tc.metadata).
					Return(tc.returnError).
					Times(tc.expectedSetCallTimes)
			}

			logger := logging.NewLogger(logging.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

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
	validMetadata := auth_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name                   string
		tenantID               string
		userID                 string
		tokenID                string
		expectedTenantID       string
		expectedKey            string
		returnMetadata         *auth_models.TokenMetadata
		returnError            error
		wantToken              *auth_models.TokenMetadata
		wantErr                bool
		expectedGetOneCallTimes int
	}{
		{
			name:                   "successful get",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			tokenID:                "token-123",
			expectedTenantID:       "tenant-123",
			expectedKey:            "user-123:token-123",
			returnMetadata:         &validMetadata,
			returnError:            nil,
			wantToken:              &validMetadata,
			wantErr:                false,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                   "token not found",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			tokenID:                "token-123",
			expectedTenantID:       "tenant-123",
			expectedKey:            "user-123:token-123",
			returnMetadata:         nil,
			returnError:            errors.New("token not found"),
			wantToken:              nil,
			wantErr:                true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                   "database error",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			tokenID:                "token-123",
			expectedTenantID:       "tenant-123",
			expectedKey:            "user-123:token-123",
			returnMetadata:         nil,
			returnError:            errors.New("database query failed"),
			wantToken:              nil,
			wantErr:                true,
			expectedGetOneCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.expectedKey).
					Return(tc.returnMetadata, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			logger := logging.NewLogger(logging.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

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
	validMetadata := auth_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name                   string
		tenantID               string
		userID                 string
		expectedTenantID       string
		expectedUserID         string
		returnTokens           []auth_models.TokenMetadata
		returnError            error
		wantToken              *auth_models.TokenMetadata
		wantErr                bool
		expectedGetAllCallTimes int
	}{
		{
			name:                   "successful get",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			expectedTenantID:       "tenant-123",
			expectedUserID:         "user-123",
			returnTokens:           []auth_models.TokenMetadata{validMetadata},
			returnError:            nil,
			wantToken:              &validMetadata,
			wantErr:                false,
			expectedGetAllCallTimes: 1,
		},
		{
			name:                   "token not found",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			expectedTenantID:       "tenant-123",
			expectedUserID:         "user-123",
			returnTokens:           []auth_models.TokenMetadata{},
			returnError:            nil,
			wantToken:              nil,
			wantErr:                false,
			expectedGetAllCallTimes: 1,
		},
		{
			name:                   "database error",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			expectedTenantID:       "tenant-123",
			expectedUserID:         "user-123",
			returnTokens:           nil,
			returnError:            errors.New("database query failed"),
			wantToken:              nil,
			wantErr:                true,
			expectedGetAllCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.TokenMetadata](ctrl)
			if tc.expectedGetAllCallTimes > 0 {
				mockHandler.EXPECT().
					GetAll(tc.expectedTenantID, tc.expectedUserID).
					Return(tc.returnTokens, tc.returnError).
					Times(tc.expectedGetAllCallTimes)
			}

			logger := logging.NewLogger(logging.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

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
	validMetadata := auth_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}
	expiredMetadata := auth_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour), // Expired
		Revoked:   false,
	}
	revokedMetadata := auth_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   true,
	}

	testCases := []struct {
		name                   string
		tenantID               string
		userID                 string
		tokenID                string
		expectedTenantID       string
		expectedKey            string
		returnMetadata         *auth_models.TokenMetadata
		returnError            error
		wantErr                bool
		expectedGetOneCallTimes int
	}{
		{
			name:                   "valid token",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			tokenID:                "token-123",
			expectedTenantID:       "tenant-123",
			expectedKey:            "user-123:token-123",
			returnMetadata:         &validMetadata,
			returnError:            nil,
			wantErr:                false,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                   "expired token",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			tokenID:                "token-123",
			expectedTenantID:       "tenant-123",
			expectedKey:            "user-123:token-123",
			returnMetadata:         &expiredMetadata,
			returnError:            nil,
			wantErr:                true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                   "revoked token",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			tokenID:                "token-123",
			expectedTenantID:       "tenant-123",
			expectedKey:            "user-123:token-123",
			returnMetadata:         &revokedMetadata,
			returnError:            nil,
			wantErr:                true,
			expectedGetOneCallTimes: 1,
		},
		{
			name:                   "token not found",
			tenantID:               "tenant-123",
			userID:                 "user-123",
			tokenID:                "token-123",
			expectedTenantID:       "tenant-123",
			expectedKey:            "user-123:token-123",
			returnMetadata:         nil,
			returnError:            errors.New("token not found"),
			wantErr:                true,
			expectedGetOneCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedTenantID, tc.expectedKey).
					Return(tc.returnMetadata, tc.returnError).
					Times(tc.expectedGetOneCallTimes)
			}

			logger := logging.NewLogger(logging.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

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
	validMetadata := auth_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name                     string
		tenantID                 string
		userID                   string
		tokenID                  string
		revokedBy                string
		expectedGetTenantID      string
		expectedGetKey           string
		expectedUpdateTenantID   string
		expectedUpdateTokenID    string
		returnGetMetadata        *auth_models.TokenMetadata
		returnGetError           error
		returnUpdateError        error
		wantErr                  bool
		expectedGetOneCallTimes   int
		expectedUpdateCallTimes  int
	}{
		{
			name:                     "successful revoke",
			tenantID:                 "tenant-123",
			userID:                   "user-123",
			tokenID:                  "token-123",
			revokedBy:                "admin",
			expectedGetTenantID:      "tenant-123",
			expectedGetKey:           "user-123:token-123",
			expectedUpdateTenantID:   "tenant-123",
			expectedUpdateTokenID:    "token-123",
			returnGetMetadata:        &validMetadata,
			returnGetError:           nil,
			returnUpdateError:        nil,
			wantErr:                  false,
			expectedGetOneCallTimes:   1,
			expectedUpdateCallTimes:  1,
		},
		{
			name:                     "token not found",
			tenantID:                 "tenant-123",
			userID:                   "user-123",
			tokenID:                  "token-123",
			revokedBy:                "admin",
			expectedGetTenantID:      "tenant-123",
			expectedGetKey:           "user-123:token-123",
			expectedUpdateTenantID:   "",
			expectedUpdateTokenID:    "",
			returnGetMetadata:        nil,
			returnGetError:           errors.New("token not found"),
			returnUpdateError:        nil,
			wantErr:                  true,
			expectedGetOneCallTimes:   1,
			expectedUpdateCallTimes:  0,
		},
		{
			name:                     "update fails",
			tenantID:                 "tenant-123",
			userID:                   "user-123",
			tokenID:                  "token-123",
			revokedBy:                "admin",
			expectedGetTenantID:      "tenant-123",
			expectedGetKey:           "user-123:token-123",
			expectedUpdateTenantID:   "tenant-123",
			expectedUpdateTokenID:    "token-123",
			returnGetMetadata:        &validMetadata,
			returnGetError:           nil,
			returnUpdateError:        errors.New("update failed"),
			wantErr:                  true,
			expectedGetOneCallTimes:   1,
			expectedUpdateCallTimes:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedGetTenantID, tc.expectedGetKey).
					Return(tc.returnGetMetadata, tc.returnGetError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedUpdateCallTimes > 0 {
				// Create expected metadata with Revoked=true and RevokedBy set
				expectedMetadata := *tc.returnGetMetadata
				expectedMetadata.Revoked = true
				expectedMetadata.RevokedBy = tc.revokedBy
				mockHandler.EXPECT().
					Update(tc.expectedUpdateTenantID, tc.expectedUpdateTokenID, tokenMetadataMatcher{expected: expectedMetadata}).
					Return(tc.returnUpdateError).
					Times(tc.expectedUpdateCallTimes)
			}

			logger := logging.NewLogger(logging.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

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
	validMetadata := auth_models.TokenMetadata{
		TokenID:   "token-123",
		UserID:    "user-123",
		TenantID:  "tenant-123",
		TokenType: "access",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}

	testCases := []struct {
		name                     string
		tenantID                 string
		userID                   string
		tokenID                  string
		expectedGetTenantID      string
		expectedGetKey           string
		expectedDeleteTenantID   string
		expectedDeleteTokenID    string
		returnGetMetadata        *auth_models.TokenMetadata
		returnGetError           error
		returnDeleteError        error
		wantErr                  bool
		expectedGetOneCallTimes   int
		expectedDeleteCallTimes  int
	}{
		{
			name:                     "successful delete",
			tenantID:                 "tenant-123",
			userID:                   "user-123",
			tokenID:                  "token-123",
			expectedGetTenantID:      "tenant-123",
			expectedGetKey:           "user-123:token-123",
			expectedDeleteTenantID:   "tenant-123",
			expectedDeleteTokenID:    "token-123",
			returnGetMetadata:        &validMetadata,
			returnGetError:           nil,
			returnDeleteError:        nil,
			wantErr:                  false,
			expectedGetOneCallTimes:   1,
			expectedDeleteCallTimes:  1,
		},
		{
			name:                     "delete with database error",
			tenantID:                 "tenant-123",
			userID:                   "user-123",
			tokenID:                  "token-123",
			expectedGetTenantID:      "tenant-123",
			expectedGetKey:           "user-123:token-123",
			expectedDeleteTenantID:   "tenant-123",
			expectedDeleteTokenID:    "token-123",
			returnGetMetadata:        &validMetadata,
			returnGetError:           nil,
			returnDeleteError:        errors.New("delete failed"),
			wantErr:                  true,
			expectedGetOneCallTimes:   1,
			expectedDeleteCallTimes:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := handlers_mocks.NewMockKeyHandler[auth_models.TokenMetadata](ctrl)
			if tc.expectedGetOneCallTimes > 0 {
				mockHandler.EXPECT().
					GetOne(tc.expectedGetTenantID, tc.expectedGetKey).
					Return(tc.returnGetMetadata, tc.returnGetError).
					Times(tc.expectedGetOneCallTimes)
			}
			if tc.expectedDeleteCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.expectedDeleteTenantID, tc.expectedDeleteTokenID).
					Return(tc.returnDeleteError).
					Times(tc.expectedDeleteCallTimes)
			}

			logger := logging.NewLogger(logging.ModuleAuth)
			handler := NewAccessTokenHandler(mockHandler, nil, logger)

			err := handler.Delete(tc.tenantID, tc.userID, tc.tokenID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
