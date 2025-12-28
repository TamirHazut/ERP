package models

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAccessTokenClaims_Validate(t *testing.T) {
	futureTime := jwt.NewNumericDate(time.Now().Add(1 * time.Hour))
	pastTime := jwt.NewNumericDate(time.Now().Add(-1 * time.Hour))

	tests := []struct {
		name           string
		claims         *AccessTokenClaims
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "valid access token claims",
			claims: &AccessTokenClaims{
				UserID:      "user-123",
				TenantID:    "tenant-123",
				Username:    "testuser",
				Roles:       []string{"admin"},
				Permissions: []string{"users:read", "users:write"},
				TokenType:   "access",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr: false,
		},
		{
			name: "missing userID",
			claims: &AccessTokenClaims{
				TenantID:    "tenant-123",
				Username:    "testuser",
				Roles:       []string{"admin"},
				Permissions: []string{"users:read"},
				TokenType:   "access",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "UserID",
		},
		{
			name: "missing tenantID",
			claims: &AccessTokenClaims{
				UserID:      "user-123",
				Username:    "testuser",
				Roles:       []string{"admin"},
				Permissions: []string{"users:read"},
				TokenType:   "access",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "TenantID",
		},
		{
			name: "missing username",
			claims: &AccessTokenClaims{
				UserID:      "user-123",
				TenantID:    "tenant-123",
				Roles:       []string{"admin"},
				Permissions: []string{"users:read"},
				TokenType:   "access",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "Username",
		},
		{
			name: "nil permissions",
			claims: &AccessTokenClaims{
				UserID:    "user-123",
				TenantID:  "tenant-123",
				Username:  "testuser",
				Roles:     []string{"admin"},
				TokenType: "access",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "Permissions",
		},
		{
			name: "empty roles",
			claims: &AccessTokenClaims{
				UserID:      "user-123",
				TenantID:    "tenant-123",
				Username:    "testuser",
				Permissions: []string{"users:read"},
				TokenType:   "access",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "Roles",
		},
		{
			name: "wrong token type",
			claims: &AccessTokenClaims{
				UserID:      "user-123",
				TenantID:    "tenant-123",
				Username:    "testuser",
				Roles:       []string{"admin"},
				Permissions: []string{"users:read"},
				TokenType:   "refresh",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "TokenType",
		},
		{
			name: "missing expiresAt",
			claims: &AccessTokenClaims{
				UserID:      "user-123",
				TenantID:    "tenant-123",
				Username:    "testuser",
				Roles:       []string{"admin"},
				Permissions: []string{"users:read"},
				TokenType:   "access",
			},
			wantErr:        true,
			expectedErrMsg: "ExpiresAt",
		},
		{
			name: "expired token",
			claims: &AccessTokenClaims{
				UserID:      "user-123",
				TenantID:    "tenant-123",
				Username:    "testuser",
				Roles:       []string{"admin"},
				Permissions: []string{"users:read"},
				TokenType:   "access",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: pastTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.Validate()
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

func TestAccessTokenClaims_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		claims   *AccessTokenClaims
		expected bool
	}{
		{
			name: "not expired",
			claims: &AccessTokenClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			},
			expected: false,
		},
		{
			name: "expired",
			claims: &AccessTokenClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.claims.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRefreshTokenClaims_Validate(t *testing.T) {
	futureTime := jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour))
	pastTime := jwt.NewNumericDate(time.Now().Add(-1 * time.Hour))

	tests := []struct {
		name           string
		claims         *RefreshTokenClaims
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "valid refresh token claims",
			claims: &RefreshTokenClaims{
				UserID:    "user-123",
				TokenType: "refresh",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr: false,
		},
		{
			name: "missing userID",
			claims: &RefreshTokenClaims{
				TokenType: "refresh",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "UserID",
		},
		{
			name: "wrong token type",
			claims: &RefreshTokenClaims{
				UserID:    "user-123",
				TokenType: "access",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: futureTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "TokenType",
		},
		{
			name: "missing expiresAt",
			claims: &RefreshTokenClaims{
				UserID:    "user-123",
				TokenType: "refresh",
			},
			wantErr:        true,
			expectedErrMsg: "ExpiresAt",
		},
		{
			name: "expired token",
			claims: &RefreshTokenClaims{
				UserID:    "user-123",
				TokenType: "refresh",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: pastTime,
				},
			},
			wantErr:        true,
			expectedErrMsg: "expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.Validate()
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
