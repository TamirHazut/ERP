package redis_models

type KeyPrefix string

// Redis Key Patterns (constants for consistency)
const (
	// Session keys
	RedisKeySession      = "sessions:%s:%s"      // sessions:{tenant_id}:{session_id}
	RedisKeyUserSessions = "user_sessions:%s:%s" // user_sessions:{tenant_id}:{user_id} -> set of session_ids

	// Token keys
	RedisKeyToken          = "tokens:%s:%s:%s"         // tokens:{tenant_id}:{user_id}:{token_id}
	RedisKeyRefreshToken   = "refresh_tokens:%s:%s:%s" // refresh_tokens:{tenant_id}:{user_id}:{token_id}
	RedisKeyRevokedToken   = "revoked_tokens:%s:%s:%s" // revoked_tokens:{tenant_id}:{user_id}:{token_id}
	RedisKeyBlacklistToken = "blacklist:%s:%s:%s"      // blacklist:{tenant_id}:{user_id}:{token_jti}

	RedisKeyUserAccessTokens  = "user_access_tokens:%s:%s"  // user_access_tokens:{tenant_id}:{user_id} -> set of token_ids
	RedisKeyUserRefreshTokens = "user_refresh_tokens:%s:%s" // user_refresh_tokens:{tenant_id}:{user_id} -> set of token_ids

	// Permission & Role cache
	RedisKeyUserPermissions = "permissions:%s:%s" // permissions:{tenant_id}:{user_id}
	RedisKeyUserRoles       = "roles:%s:%s"       // roles:{tenant_id}:{user_id}
	RedisKeyRolePermissions = "role_perms:%s:%s"  // role_perms:{tenant_id}:{role_id}

	// Rate limiting
	RedisKeyRateLimit       = "rate_limit:%s:%s:%s" // rate_limit:{tenant_id}:{user_id}:{endpoint}
	RedisKeyTenantRateLimit = "tenant_limit:%s:%s"  // tenant_limit:{tenant_id}:{endpoint}
	RedisKeyIPRateLimit     = "ip_limit:%s:%s:%s"   // ip_limit:{tenant_id}:{ip_address}:{endpoint}

	// Cache keys
	RedisKeyQueryCache   = "query_cache:%s:%s"   // query_cache:{tenant_id}:{query_hash}
	RedisKeyUserCache    = "user_cache:%s:%s"    // user_cache:{tenant_id}:{user_id}
	RedisKeyTenantCache  = "tenant_cache:%s"     // tenant_cache:{tenant_id}
	RedisKeyProductCache = "product_cache:%s:%s" // product_cache:{tenant_id}:{product_id}
	RedisKeyOrderCache   = "order_cache:%s:%s"   // order_cache:{tenant_id}:{order_id}

	// Locks (for distributed locking)
	RedisKeyLock = "lock:%s:%s" // lock:{tenant_id}:{resource_id}

	// Temporary data
	RedisKeyPasswordReset = "pwd_reset:%s:%s"    // pwd_reset:{tenant_id}:{token}
	RedisKeyEmailVerify   = "email_verify:%s:%s" // email_verify:{tenant_id}:{token}
	RedisKeyMFACode       = "mfa_code:%s:%s"     // mfa_code:{tenant_id}:{user_id}
	RedisKeyInviteToken   = "invite:%s:%s"       // invite:{tenant_id}:{token}

	// Analytics & Metrics
	RedisKeyLoginAttempts = "login_attempts:%s:%s" // login_attempts:{tenant_id}:{user_id}
	RedisKeyActiveUsers   = "active_users:%s"      // active_users:{tenant_id} -> set
	RedisKeyOnlineUsers   = "online_users:%s"      // online_users:{tenant_id} -> sorted set

	// Feature flags cache
	RedisKeyFeatureFlag    = "feature_flag:%s:%s" // feature_flag:{tenant_id}:{flag_key}
	RedisKeyTenantFeatures = "tenant_features:%s" // tenant_features:{tenant_id}

	// Config cache
	RedisKeyServiceConfig = "config:%s:%s:%s" // config:{tenant_id}:{service_name}:{environment}
)
