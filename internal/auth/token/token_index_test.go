package token

import (
	"errors"
	"testing"

	handlers_mocks "erp.localhost/internal/infra/db/redis/handlers/mocks"
	"erp.localhost/internal/infra/logging"
	shared_models "erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Tests for NewTokenIndex

func TestNewTokenIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAccessTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)
	mockRefreshTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

	tokenIndex := NewTokenIndex(mockAccessTokenHandler, mockRefreshTokenHandler)

	require.NotNil(t, tokenIndex)
	assert.NotNil(t, tokenIndex.accessTokenSetHandler)
	assert.NotNil(t, tokenIndex.refreshTokenSetHandler)
	assert.NotNil(t, tokenIndex.logger)
}

func TestNewTokenIndex_WithNilHandlers(t *testing.T) {
	tokenIndex := NewTokenIndex(nil, nil)

	require.NotNil(t, tokenIndex)
	assert.NotNil(t, tokenIndex.accessTokenSetHandler)
	assert.NotNil(t, tokenIndex.refreshTokenSetHandler)
	assert.NotNil(t, tokenIndex.logger)
}

// Tests for AddAccessToken

func TestTokenIndex_AddAccessToken(t *testing.T) {
	tenantID := "tenant-1"
	userID := "user-1"
	tokenID := "token-123"

	testCases := []struct {
		name        string
		tenantID    string
		userID      string
		tokenID     string
		returnError error
		wantErr     bool
	}{
		{
			name:        "successful add access token",
			tenantID:    tenantID,
			userID:      userID,
			tokenID:     tokenID,
			returnError: nil,
			wantErr:     false,
		},
		{
			name:        "add access token with database error",
			tenantID:    tenantID,
			userID:      userID,
			tokenID:     tokenID,
			returnError: errors.New("redis connection failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAccessTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

			opts := map[string]any{
				"ttl":      accessTokenTTL,
				"ttl_unit": accessTokenTTLUnit,
			}

			mockAccessTokenHandler.EXPECT().
				Add(tc.tenantID, tc.userID, tc.tokenID, opts).
				Return(tc.returnError).
				Times(1)

			tokenIndex := &TokenIndex{
				accessTokenSetHandler:  mockAccessTokenHandler,
				refreshTokenSetHandler: nil,
				logger:                 logging.NewLogger(shared_models.ModuleAuth),
			}

			err := tokenIndex.AddAccessToken(tc.tenantID, tc.userID, tc.tokenID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Tests for RemoveAccessToken

func TestTokenIndex_RemoveAccessToken(t *testing.T) {
	tenantID := "tenant-1"
	userID := "user-1"
	tokenID := "token-123"

	testCases := []struct {
		name        string
		tenantID    string
		userID      string
		tokenID     string
		returnError error
		wantErr     bool
	}{
		{
			name:        "successful remove access token",
			tenantID:    tenantID,
			userID:      userID,
			tokenID:     tokenID,
			returnError: nil,
			wantErr:     false,
		},
		{
			name:        "remove access token with database error",
			tenantID:    tenantID,
			userID:      userID,
			tokenID:     tokenID,
			returnError: errors.New("redis connection failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAccessTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

			mockAccessTokenHandler.EXPECT().
				Remove(tc.tenantID, tc.userID, tc.tokenID).
				Return(tc.returnError).
				Times(1)

			tokenIndex := &TokenIndex{
				accessTokenSetHandler:  mockAccessTokenHandler,
				refreshTokenSetHandler: nil,
				logger:                 logging.NewLogger(shared_models.ModuleAuth),
			}

			err := tokenIndex.RemoveAccessToken(tc.tenantID, tc.userID, tc.tokenID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Tests for GetAccessTokens

func TestTokenIndex_GetAccessTokens(t *testing.T) {
	tenantID := "tenant-1"
	userID := "user-1"

	testCases := []struct {
		name          string
		tenantID      string
		userID        string
		returnTokens  []string
		returnError   error
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "successful get access tokens with multiple tokens",
			tenantID:      tenantID,
			userID:        userID,
			returnTokens:  []string{"token-1", "token-2", "token-3"},
			returnError:   nil,
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name:          "get access tokens with empty result",
			tenantID:      tenantID,
			userID:        userID,
			returnTokens:  []string{},
			returnError:   nil,
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:         "get access tokens with database error",
			tenantID:     tenantID,
			userID:       userID,
			returnTokens: nil,
			returnError:  errors.New("redis connection failed"),
			wantErr:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAccessTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

			mockAccessTokenHandler.EXPECT().
				Members(tc.tenantID, tc.userID).
				Return(tc.returnTokens, tc.returnError).
				Times(1)

			tokenIndex := &TokenIndex{
				accessTokenSetHandler:  mockAccessTokenHandler,
				refreshTokenSetHandler: nil,
				logger:                 logging.NewLogger(shared_models.ModuleAuth),
			}

			tokens, err := tokenIndex.GetAccessTokens(tc.tenantID, tc.userID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, tokens)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedCount, len(tokens))
				assert.Equal(t, tc.returnTokens, tokens)
			}
		})
	}
}

// Tests for ClearAccessTokens

func TestTokenIndex_ClearAccessTokens(t *testing.T) {
	tenantID := "tenant-1"
	userID := "user-1"

	testCases := []struct {
		name        string
		tenantID    string
		userID      string
		returnError error
		wantErr     bool
	}{
		{
			name:        "successful clear access tokens",
			tenantID:    tenantID,
			userID:      userID,
			returnError: nil,
			wantErr:     false,
		},
		{
			name:        "clear access tokens with database error",
			tenantID:    tenantID,
			userID:      userID,
			returnError: errors.New("redis connection failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAccessTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

			mockAccessTokenHandler.EXPECT().
				Clear(tc.tenantID, tc.userID).
				Return(tc.returnError).
				Times(1)

			tokenIndex := &TokenIndex{
				accessTokenSetHandler:  mockAccessTokenHandler,
				refreshTokenSetHandler: nil,
				logger:                 logging.NewLogger(shared_models.ModuleAuth),
			}

			err := tokenIndex.ClearAccessTokens(tc.tenantID, tc.userID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Tests for AddRefreshToken

func TestTokenIndex_AddRefreshToken(t *testing.T) {
	tenantID := "tenant-1"
	userID := "user-1"
	tokenID := "refresh-token-123"

	testCases := []struct {
		name        string
		tenantID    string
		userID      string
		tokenID     string
		returnError error
		wantErr     bool
	}{
		{
			name:        "successful add refresh token",
			tenantID:    tenantID,
			userID:      userID,
			tokenID:     tokenID,
			returnError: nil,
			wantErr:     false,
		},
		{
			name:        "add refresh token with database error",
			tenantID:    tenantID,
			userID:      userID,
			tokenID:     tokenID,
			returnError: errors.New("redis connection failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRefreshTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

			opts := map[string]any{
				"ttl":      refreshTokenTTL,
				"ttl_unit": refreshTokenTTLUnit,
			}

			mockRefreshTokenHandler.EXPECT().
				Add(tc.tenantID, tc.userID, tc.tokenID, opts).
				Return(tc.returnError).
				Times(1)

			tokenIndex := &TokenIndex{
				accessTokenSetHandler:  nil,
				refreshTokenSetHandler: mockRefreshTokenHandler,
				logger:                 logging.NewLogger(shared_models.ModuleAuth),
			}

			err := tokenIndex.AddRefreshToken(tc.tenantID, tc.userID, tc.tokenID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Tests for RemoveRefreshToken

func TestTokenIndex_RemoveRefreshToken(t *testing.T) {
	tenantID := "tenant-1"
	userID := "user-1"
	tokenID := "refresh-token-123"

	testCases := []struct {
		name        string
		tenantID    string
		userID      string
		tokenID     string
		returnError error
		wantErr     bool
	}{
		{
			name:        "successful remove refresh token",
			tenantID:    tenantID,
			userID:      userID,
			tokenID:     tokenID,
			returnError: nil,
			wantErr:     false,
		},
		{
			name:        "remove refresh token with database error",
			tenantID:    tenantID,
			userID:      userID,
			tokenID:     tokenID,
			returnError: errors.New("redis connection failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRefreshTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

			mockRefreshTokenHandler.EXPECT().
				Remove(tc.tenantID, tc.userID, tc.tokenID).
				Return(tc.returnError).
				Times(1)

			tokenIndex := &TokenIndex{
				accessTokenSetHandler:  nil,
				refreshTokenSetHandler: mockRefreshTokenHandler,
				logger:                 logging.NewLogger(shared_models.ModuleAuth),
			}

			err := tokenIndex.RemoveRefreshToken(tc.tenantID, tc.userID, tc.tokenID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Tests for GetRefreshTokens

func TestTokenIndex_GetRefreshTokens(t *testing.T) {
	tenantID := "tenant-1"
	userID := "user-1"

	testCases := []struct {
		name          string
		tenantID      string
		userID        string
		returnTokens  []string
		returnError   error
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "successful get refresh tokens with multiple tokens",
			tenantID:      tenantID,
			userID:        userID,
			returnTokens:  []string{"refresh-token-1", "refresh-token-2"},
			returnError:   nil,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "get refresh tokens with empty result",
			tenantID:      tenantID,
			userID:        userID,
			returnTokens:  []string{},
			returnError:   nil,
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:         "get refresh tokens with database error",
			tenantID:     tenantID,
			userID:       userID,
			returnTokens: nil,
			returnError:  errors.New("redis connection failed"),
			wantErr:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRefreshTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

			mockRefreshTokenHandler.EXPECT().
				Members(tc.tenantID, tc.userID).
				Return(tc.returnTokens, tc.returnError).
				Times(1)

			tokenIndex := &TokenIndex{
				accessTokenSetHandler:  nil,
				refreshTokenSetHandler: mockRefreshTokenHandler,
				logger:                 logging.NewLogger(shared_models.ModuleAuth),
			}

			tokens, err := tokenIndex.GetRefreshTokens(tc.tenantID, tc.userID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, tokens)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedCount, len(tokens))
				assert.Equal(t, tc.returnTokens, tokens)
			}
		})
	}
}

// Tests for ClearRefreshTokens

func TestTokenIndex_ClearRefreshTokens(t *testing.T) {
	tenantID := "tenant-1"
	userID := "user-1"

	testCases := []struct {
		name        string
		tenantID    string
		userID      string
		returnError error
		wantErr     bool
	}{
		{
			name:        "successful clear refresh tokens",
			tenantID:    tenantID,
			userID:      userID,
			returnError: nil,
			wantErr:     false,
		},
		{
			name:        "clear refresh tokens with database error",
			tenantID:    tenantID,
			userID:      userID,
			returnError: errors.New("redis connection failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRefreshTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

			mockRefreshTokenHandler.EXPECT().
				Clear(tc.tenantID, tc.userID).
				Return(tc.returnError).
				Times(1)

			tokenIndex := &TokenIndex{
				accessTokenSetHandler:  nil,
				refreshTokenSetHandler: mockRefreshTokenHandler,
				logger:                 logging.NewLogger(shared_models.ModuleAuth),
			}

			err := tokenIndex.ClearRefreshTokens(tc.tenantID, tc.userID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Integration test - Multiple operations

func TestTokenIndex_MultipleOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tenantID := "tenant-1"
	userID := "user-1"
	token1 := "token-1"
	token2 := "token-2"

	mockAccessTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)
	mockRefreshTokenHandler := handlers_mocks.NewMockSetHandler(ctrl)

	tokenIndex := NewTokenIndex(mockAccessTokenHandler, mockRefreshTokenHandler)

	// Add first access token
	mockAccessTokenHandler.EXPECT().
		Add(tenantID, userID, token1, map[string]any{
			"ttl":      accessTokenTTL,
			"ttl_unit": accessTokenTTLUnit,
		}).
		Return(nil).
		Times(1)

	err := tokenIndex.AddAccessToken(tenantID, userID, token1)
	require.NoError(t, err)

	// Add second access token
	mockAccessTokenHandler.EXPECT().
		Add(tenantID, userID, token2, map[string]any{
			"ttl":      accessTokenTTL,
			"ttl_unit": accessTokenTTLUnit,
		}).
		Return(nil).
		Times(1)

	err = tokenIndex.AddAccessToken(tenantID, userID, token2)
	require.NoError(t, err)

	// Get all access tokens
	mockAccessTokenHandler.EXPECT().
		Members(tenantID, userID).
		Return([]string{token1, token2}, nil).
		Times(1)

	tokens, err := tokenIndex.GetAccessTokens(tenantID, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(tokens))

	// Remove one token
	mockAccessTokenHandler.EXPECT().
		Remove(tenantID, userID, token1).
		Return(nil).
		Times(1)

	err = tokenIndex.RemoveAccessToken(tenantID, userID, token1)
	require.NoError(t, err)

	// Clear all tokens
	mockAccessTokenHandler.EXPECT().
		Clear(tenantID, userID).
		Return(nil).
		Times(1)

	err = tokenIndex.ClearAccessTokens(tenantID, userID)
	require.NoError(t, err)
}
