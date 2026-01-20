package token

import (
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JWTAccessClaims wraps AccessTokenClaims for JWT operations
type JWTAccessClaims struct {
	jwt.RegisteredClaims // Contains ID (jti), but we don't persist it

	// Custom claims from proto (NO token_id)
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// ToProtoClaims converts JWT claims to proto (jti is NOT included in proto)
func (c *JWTAccessClaims) ToProtoClaims() *authv1.AccessTokenClaims {
	return &authv1.AccessTokenClaims{
		// NO TokenId - not needed for single token per user
		UserId:    c.UserID,
		TenantId:  c.TenantID,
		Username:  c.Username,
		Email:     c.Email,
		Roles:     c.Roles,
		IssuedAt:  timestamppb.New(c.IssuedAt.Time),
		ExpiresAt: timestamppb.New(c.ExpiresAt.Time),
	}
}

// FromProtoClaims creates JWT claims from proto
func FromProtoClaims(claims *authv1.AccessTokenClaims, issuer string) *JWTAccessClaims {
	return &JWTAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // Generate jti for JWT standard
			Issuer:    issuer,
			Subject:   claims.UserId,
			ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt.AsTime()),
			IssuedAt:  jwt.NewNumericDate(claims.IssuedAt.AsTime()),
		},
		UserID:   claims.UserId,
		TenantID: claims.TenantId,
		Username: claims.Username,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}
}
