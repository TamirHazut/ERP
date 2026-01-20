package redis

type KeyPrefix string

// Redis Key Patterns (constants for consistency)
const (
	// Session keys
	RedisKeySession      = "sessions"      // sessions:{tenant_id}:{session_id}
	RedisKeyUserSessions = "user_sessions" // user_sessions:{tenant_id}:{user_id} -> set of session_ids

	// Token keys
	RedisKeyToken          = "tokens"         // tokens:{tenant_id}:{user_id}
	RedisKeyRefreshToken   = "refresh_tokens" // refresh_tokens:{tenant_id}:{user_id}
	RedisKeyRevokedToken   = "revoked_tokens" // revoked_tokens:{tenant_id}:{user_id}
	RedisKeyBlacklistToken = "blacklist"      // blacklist:{tenant_id}:{user_id}

	RedisKeyUserAccessTokens  = "user_access_tokens"  // user_access_tokens:{tenant_id}:{user_id} -> set of token_ids
	RedisKeyUserRefreshTokens = "user_refresh_tokens" // user_refresh_tokens:{tenant_id}:{user_id} -> set of token_ids

	// Permission & Role cache
	RedisKeyUserPermissions = "permissions" // permissions:{tenant_id}:{user_id}
	RedisKeyUserRoles       = "roles"       // roles:{tenant_id}:{user_id}
	RedisKeyRolePermissions = "role_perms"  // role_perms:{tenant_id}:{role_id}

	// Rate limiting
	RedisKeyRateLimit       = "rate_limit"   // rate_limit:{tenant_id}:{user_id}:{endpoint}
	RedisKeyTenantRateLimit = "tenant_limit" // tenant_limit:{tenant_id}:{endpoint}
	RedisKeyIPRateLimit     = "ip_limit"     // ip_limit:{tenant_id}:{ip_address}:{endpoint}

	// Cache keys
	RedisKeyQueryCache   = "query_cache"   // query_cache:{tenant_id}:{query_hash}
	RedisKeyUserCache    = "user_cache"    // user_cache:{tenant_id}:{user_id}
	RedisKeyTenantCache  = "tenant_cache"  // tenant_cache:{tenant_id}
	RedisKeyProductCache = "product_cache" // product_cache:{tenant_id}:{product_id}
	RedisKeyOrderCache   = "order_cache"   // order_cache:{tenant_id}:{order_id}

	// Locks (for distributed locking)
	RedisKeyLock = "lock" // lock:{tenant_id}:{resource_id}

	// Temporary data
	RedisKeyPasswordReset = "pwd_reset"    // pwd_reset:{tenant_id}:{token}
	RedisKeyEmailVerify   = "email_verify" // email_verify:{tenant_id}:{token}
	RedisKeyMFACode       = "mfa_code"     // mfa_code:{tenant_id}:{user_id}
	RedisKeyInviteToken   = "invite"       // invite:{tenant_id}:{token}

	// Analytics & Metrics
	RedisKeyLoginAttempts = "login_attempts" // login_attempts:{tenant_id}:{user_id}
	RedisKeyActiveUsers   = "active_users"   // active_users:{tenant_id} -> set
	RedisKeyOnlineUsers   = "online_users"   // online_users:{tenant_id} -> sorted set

	// Feature flags cache
	RedisKeyFeatureFlag    = "feature_flag"    // feature_flag:{tenant_id}:{flag_key}
	RedisKeyTenantFeatures = "tenant_features" // tenant_features:{tenant_id}

	// Config cache
	RedisKeyServiceConfig = "config" // config:{tenant_id}:{service_name}:{environment}
)
