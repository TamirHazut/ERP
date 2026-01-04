package auth_models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRefreshToken_Validate(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(7 * 24 * time.Hour)
	pastTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name           string
		token          *RefreshToken
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "valid refresh token",
			token: &RefreshToken{
				Token:     "valid-token",
				UserID:    "user-123",
				TenantID:  "tenant-123",
				SessionID: "session-123",
				ExpiresAt: futureTime,
				CreatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "missing token",
			token: &RefreshToken{
				UserID:    "user-123",
				TenantID:  "tenant-123",
				SessionID: "session-123",
				ExpiresAt: futureTime,
				CreatedAt: now,
			},
			wantErr:        true,
			expectedErrMsg: "Token",
		},
		{
			name: "missing tenantID",
			token: &RefreshToken{
				Token:     "valid-token",
				UserID:    "user-123",
				SessionID: "session-123",
				ExpiresAt: futureTime,
				CreatedAt: now,
			},
			wantErr:        true,
			expectedErrMsg: "TenantID",
		},
		{
			name: "missing userID",
			token: &RefreshToken{
				Token:     "valid-token",
				TenantID:  "tenant-123",
				SessionID: "session-123",
				ExpiresAt: futureTime,
				CreatedAt: now,
			},
			wantErr:        true,
			expectedErrMsg: "UserID",
		},
		{
			name: "missing expiresAt",
			token: &RefreshToken{
				Token:     "valid-token",
				UserID:    "user-123",
				TenantID:  "tenant-123",
				SessionID: "session-123",
				CreatedAt: now,
			},
			wantErr:        true,
			expectedErrMsg: "ExpiresAt",
		},
		{
			name: "missing createdAt",
			token: &RefreshToken{
				Token:     "valid-token",
				UserID:    "user-123",
				TenantID:  "tenant-123",
				SessionID: "session-123",
				ExpiresAt: futureTime,
			},
			wantErr:        true,
			expectedErrMsg: "CreatedAt",
		},
		{
			name: "expired token",
			token: &RefreshToken{
				Token:     "valid-token",
				UserID:    "user-123",
				TenantID:  "tenant-123",
				SessionID: "session-123",
				ExpiresAt: pastTime,
				CreatedAt: now,
			},
			wantErr:        true,
			expectedErrMsg: "expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.token.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRefreshToken_IsValid(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(7 * 24 * time.Hour)
	pastTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name     string
		token    *RefreshToken
		expected bool
	}{
		{
			name: "valid and not revoked",
			token: &RefreshToken{
				ExpiresAt: futureTime,
				IsRevoked: false,
			},
			expected: true,
		},
		{
			name: "valid but revoked",
			token: &RefreshToken{
				ExpiresAt: futureTime,
				IsRevoked: true,
			},
			expected: false,
		},
		{
			name: "expired and not revoked",
			token: &RefreshToken{
				ExpiresAt: pastTime,
				IsRevoked: false,
			},
			expected: false,
		},
		{
			name: "expired and revoked",
			token: &RefreshToken{
				ExpiresAt: pastTime,
				IsRevoked: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRefreshToken_IsExpired(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(7 * 24 * time.Hour)
	pastTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name     string
		token    *RefreshToken
		expected bool
	}{
		{
			name: "not expired",
			token: &RefreshToken{
				ExpiresAt: futureTime,
			},
			expected: false,
		},
		{
			name: "expired",
			token: &RefreshToken{
				ExpiresAt: pastTime,
			},
			expected: true,
		},
		{
			name: "zero time (treated as expired)",
			token: &RefreshToken{
				ExpiresAt: time.Time{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}
