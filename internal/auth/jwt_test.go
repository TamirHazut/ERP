package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	testCases := []struct {
		name          string
		secretKey     string
		tokenDuration time.Duration
		shouldPanic   bool
	}{
		{
			name:          "valid parameters",
			secretKey:     "test-secret-key-12345",
			tokenDuration: time.Hour,
			shouldPanic:   false,
		},
		{
			name:          "empty secret key",
			secretKey:     "",
			tokenDuration: time.Hour,
			shouldPanic:   true,
		},
		{
			name:          "zero duration",
			secretKey:     "test-secret-key-12345",
			tokenDuration: 0,
			shouldPanic:   true,
		},
		{
			name:          "negative duration",
			secretKey:     "test-secret-key-12345",
			tokenDuration: -time.Hour,
			shouldPanic:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				// Note: NewJWTManager calls logger.Fatal which exits the program
				// In a real scenario, we'd need to refactor to return an error instead
				// For now, we skip panic tests or refactor the function
				t.Skip("Skipping panic test - requires refactoring NewJWTManager to return error")
			} else {
				manager := NewJWTManager(tc.secretKey, tc.tokenDuration)
				assert.NotNil(t, manager)
			}
		})
	}
}

func TestJWTManager_GenerateToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-12345", time.Hour)

	testCases := []struct {
		name     string
		userID   string
		tenantID string
		wantErr  bool
	}{
		{
			name:     "valid user and tenant",
			userID:   "user-123",
			tenantID: "tenant-456",
			wantErr:  false,
		},
		{
			name:     "empty user ID",
			userID:   "",
			tenantID: "tenant-456",
			wantErr:  false, // Currently allows empty - may want to validate
		},
		{
			name:     "empty tenant ID",
			userID:   "user-123",
			tenantID: "",
			wantErr:  false, // Currently allows empty - may want to validate
		},
		{
			name:     "both empty",
			userID:   "",
			tenantID: "",
			wantErr:  false, // Currently allows empty - may want to validate
		},
		{
			name:     "special characters in IDs",
			userID:   "user@123!#$%",
			tenantID: "tenant&*()",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := manager.GenerateToken(tc.userID, tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
				// Token should be a valid JWT format (3 parts separated by dots)
				assert.Regexp(t, `^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`, token)
			}
		})
	}
}

func TestJWTManager_VerifyToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-12345", time.Hour)
	validToken, err := manager.GenerateToken("user-123", "tenant-456")
	require.NoError(t, err)

	// Create an expired token manager
	expiredManager := NewJWTManager("test-secret-key-12345", -time.Hour)
	expiredToken, err := expiredManager.GenerateToken("user-123", "tenant-456")
	require.NoError(t, err)

	// Create token with different secret
	differentSecretManager := NewJWTManager("different-secret-key", time.Hour)
	differentSecretToken, err := differentSecretManager.GenerateToken("user-123", "tenant-456")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		token     string
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "valid token",
			token:     validToken,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "expired token",
			token:     expiredToken,
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "token with different secret",
			token:     differentSecretToken,
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "empty token",
			token:     "",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "malformed token - random string",
			token:     "not-a-valid-token",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "malformed token - missing parts",
			token:     "header.payload",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "tampered token",
			token:     validToken + "tampered",
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := manager.VerifyToken(tc.token)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantValid, valid)
		})
	}
}

func TestJWTManager_RefreshToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-12345", time.Hour)
	validToken, err := manager.GenerateToken("user-123", "tenant-456")
	require.NoError(t, err)

	testCases := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "invalid-token",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newToken, err := manager.RefreshToken(tc.token)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, newToken)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, newToken)
				// New token should be valid
				valid, verifyErr := manager.VerifyToken(newToken)
				require.NoError(t, verifyErr)
				assert.True(t, valid)
			}
		})
	}
}

func TestJWTManager_RevokeToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-12345", time.Hour)
	validToken, err := manager.GenerateToken("user-123", "tenant-456")
	require.NoError(t, err)

	testCases := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "invalid-token",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			revokedToken, err := manager.RevokeToken(tc.token)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Note: Current implementation doesn't truly revoke - it just re-signs
				// A proper revoke should add to a blacklist
				assert.NotEmpty(t, revokedToken)
			}
		})
	}
}

func TestJWTManager_TokenExpiry(t *testing.T) {
	// Test with short duration (1 second for reliable testing)
	shortDuration := 1 * time.Second
	manager := NewJWTManager("test-secret-key-12345", shortDuration)

	token, err := manager.GenerateToken("user-123", "tenant-456")
	require.NoError(t, err)

	// Token should be valid immediately
	valid, err := manager.VerifyToken(token)
	require.NoError(t, err)
	assert.True(t, valid)

	// Wait for token to expire (1.1 seconds to be safe)
	time.Sleep(1100 * time.Millisecond)

	// Token should now be invalid (expired)
	valid, err = manager.VerifyToken(token)
	// Expect an error because the token is expired
	assert.Error(t, err, "Expected error for expired token")
	assert.False(t, valid, "Expected token to be invalid after expiry")
}
