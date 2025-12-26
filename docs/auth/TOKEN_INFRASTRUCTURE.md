# Token Infrastructure Design

## Overview

This document describes the token management infrastructure for the ERP system, including access tokens and refresh tokens stored in Redis.

## Design Decisions

### 1. **Per Tenant+User vs Per User**

**Decision: Per Tenant+User**

Tokens are scoped to both tenant and user for the following reasons:

- **Multi-tenant isolation**: A user can belong to multiple tenants and should have separate tokens for each
- **Security**: If a user's tokens are compromised in one tenant, other tenants remain secure
- **Compliance**: Better audit trail and access control per tenant
- **Scalability**: Easier to revoke all tokens for a specific tenant+user combination

### 2. **Key Patterns**

#### Access Tokens
- **Key Pattern**: `tokens:{tenant_id}:{token_id}`
- **Storage**: `TokenMetadata` model
- **TTL**: Matches JWT expiry (typically 15 minutes - 1 hour)

#### Refresh Tokens
- **Key Pattern**: `refresh_tokens:{tenant_id}:{user_id}:{token_id}`
- **Storage**: `RefreshToken` model (from `internal/auth/models`)
- **TTL**: 7 days (configurable)
- **Multiple tokens**: Users can have multiple refresh tokens (one per device/session)

## Components

### 1. AccessTokenKeyHandler

**Location**: `internal/auth/keys_handlers/access_token.go`

**Functions**:
- `Store(tenantID, tokenID, metadata)` - Store access token metadata
- `Get(tenantID, tokenID)` - Retrieve access token metadata
- `Validate(tenantID, tokenID)` - Check if token is valid (not revoked, not expired)
- `Revoke(tenantID, tokenID, revokedBy)` - Revoke a single access token
- `RevokeAll(tenantID, userID, revokedBy)` - Revoke all access tokens for a user (TODO: requires token index)
- `Delete(tenantID, tokenID)` - Hard delete a token

**Usage Example**:
```go
handler := keyshandlers.NewAccessTokenKeyHandler(redis.KeyPrefix("tokens"))

// Store token
metadata := redis_models.TokenMetadata{
    TokenID:   "token-123",
    JTI:       "jti-123",
    UserID:   "user-456",
    TenantID:  "tenant-789",
    TokenType: "access",
    IssuedAt:  time.Now(),
    ExpiresAt: time.Now().Add(15 * time.Minute),
}
err := handler.Store("tenant-789", "token-123", metadata)

// Validate token
metadata, err := handler.Validate("tenant-789", "token-123")
```

### 2. RefreshTokenKeyHandler

**Location**: `internal/auth/keys_handlers/refresh_token.go`

**Functions**:
- `Store(tenantID, userID, tokenID, refreshToken)` - Store refresh token
- `Get(tenantID, userID, tokenID)` - Retrieve refresh token
- `Validate(tenantID, userID, tokenID)` - Check if refresh token is valid
- `Revoke(tenantID, userID, tokenID)` - Revoke a single refresh token
- `RevokeAll(tenantID, userID)` - Revoke all refresh tokens for a user (TODO: requires token index)
- `UpdateLastUsed(tenantID, userID, tokenID)` - Update last used timestamp
- `Delete(tenantID, userID, tokenID)` - Hard delete a refresh token

**Usage Example**:
```go
handler := keyshandlers.NewRefreshTokenKeyHandler(redis.KeyPrefix("refresh_tokens"))

// Store refresh token
refreshToken := models.RefreshToken{
    Token:     "refresh-token-string",
    UserID:    "user-456",
    TenantID:  "tenant-789",
    SessionID: "session-123",
    ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
    CreatedAt: time.Now(),
}
err := handler.Store("tenant-789", "user-456", "token-123", refreshToken)

// Validate and use
token, err := handler.Validate("tenant-789", "user-456", "token-123")
if err == nil {
    // Update last used
    handler.UpdateLastUsed("tenant-789", "user-456", "token-123")
}
```

## Token Flow

### 1. Login Flow
```
1. User authenticates with credentials
2. Generate access token (JWT) with short expiry (15 min - 1 hour)
3. Generate refresh token (JWT or random string) with long expiry (7 days)
4. Store access token metadata in Redis: tokens:{tenant_id}:{token_id}
5. Store refresh token in Redis: refresh_tokens:{tenant_id}:{user_id}:{token_id}
6. Return both tokens to client
```

### 2. Access Token Validation Flow
```
1. Client sends access token in Authorization header
2. Verify JWT signature and expiry
3. Check Redis: tokens:{tenant_id}:{token_id}
   - If not found or revoked → reject
   - If expired → reject
4. If valid → allow request
```

### 3. Refresh Token Flow
```
1. Client sends refresh token
2. Validate refresh token in Redis: refresh_tokens:{tenant_id}:{user_id}:{token_id}
   - Check if exists, not revoked, not expired
3. Generate new access token
4. Optionally rotate refresh token (generate new one, revoke old one)
5. Update refresh token LastUsedAt
6. Return new access token (and optionally new refresh token)
```

### 4. Logout Flow
```
1. Revoke access token: tokens:{tenant_id}:{token_id}
2. Revoke refresh token: refresh_tokens:{tenant_id}:{user_id}:{token_id}
3. Optionally revoke all tokens for user: RevokeAll(tenantID, userID)
```

## Token Models

### TokenMetadata (Redis Model)
```go
type TokenMetadata struct {
    TokenID   string     // Unique token identifier
    JTI       string     // JWT ID (from JWT claims)
    UserID    string     // User who owns the token
    TenantID  string     // Tenant context
    TokenType string     // "access" or "refresh"
    IssuedAt  time.Time
    ExpiresAt time.Time
    Revoked   bool
    RevokedAt *time.Time
    RevokedBy string
    IPAddress string
    UserAgent string
    Scopes    []string
}
```

### RefreshToken (Auth Model)
```go
type RefreshToken struct {
    Token      string    // The actual token string
    UserID     string    // Owner of the token
    TenantID   string    // Tenant ID
    SessionID  string    // Session ID
    DeviceID   string    // Device identifier
    IPAddress  string    // IP when token was created
    UserAgent string    // Browser/app info
    ExpiresAt  time.Time // When token expires
    CreatedAt  time.Time // When token was created
    LastUsedAt time.Time // Last time token was used
    RevokedAt  time.Time // When token was revoked
    IsRevoked  bool      // Quick check if revoked
}
```

## Future Improvements

### 1. Token Index for Efficient RevokeAll
Currently, `RevokeAll` is not fully implemented because it requires scanning all tokens. To implement efficiently:

**Option A: Maintain a Set Index**
- Key: `user_tokens:{tenant_id}:{user_id}` → Redis Set of token_ids
- When storing a token, also add token_id to the set
- When revoking all, iterate over the set and revoke each token

**Option B: Use Redis Hash**
- Key: `user_tokens:{tenant_id}:{user_id}` → Redis Hash
- Field: `token_id`, Value: token metadata JSON
- Allows efficient retrieval and deletion

### 2. Token Rotation
Implement refresh token rotation for better security:
- When refreshing, generate a new refresh token
- Revoke the old refresh token
- Return both new access and refresh tokens

### 3. Device Management
Track and manage tokens per device:
- List all active devices/tokens for a user
- Revoke tokens for a specific device
- Show device information in user settings

### 4. Token Blacklist
For immediate revocation of access tokens (before expiry):
- Store revoked tokens in: `blacklist:{jti}` or `revoked_tokens:{token_id}`
- Check blacklist during token validation
- TTL matches original token expiry

## Security Considerations

1. **Token Storage**: Tokens are stored in Redis with appropriate TTLs
2. **Revocation**: Tokens can be immediately revoked (stored in Redis)
3. **Expiry**: Both access and refresh tokens have expiry times
4. **Multi-tenant Isolation**: Tokens are scoped per tenant+user
5. **Token Rotation**: Consider implementing refresh token rotation
6. **HTTPS Only**: Tokens should only be transmitted over HTTPS
7. **Secure Storage**: Client should store tokens securely (httpOnly cookies, secure storage)

## Testing

Unit tests should be created for:
- `AccessTokenKeyHandler` - all methods
- `RefreshTokenKeyHandler` - all methods
- Token validation logic
- Revocation logic

Use the mock Redis handler (`internal/db/redis/mock/handler.go`) for testing.

