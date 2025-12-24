package redis

type KeyPrefix string

// Redis Key Patterns (constants for consistency)
const (
	// Session keys
	RedisKeySession      = "sessions:%s"      // sessions:{session_id}
	RedisKeyUserSessions = "user_sessions:%s" // user_sessions:{user_id} -> set of session_ids

	// Token keys
	RedisKeyToken          = "tokens:%s"         // tokens:{token_id}
	RedisKeyRefreshToken   = "refresh_tokens:%s" // refresh_tokens:{user_id}
	RedisKeyRevokedToken   = "revoked_tokens:%s" // revoked_tokens:{token_id}
	RedisKeyBlacklistToken = "blacklist:%s"      // blacklist:{token_jti}

	// Permission & Role cache
	RedisKeyUserPermissions = "permissions:%s" // permissions:{user_id}
	RedisKeyUserRoles       = "roles:%s"       // roles:{user_id}
	RedisKeyRolePermissions = "role_perms:%s"  // role_perms:{role_id}

	// Rate limiting
	RedisKeyRateLimit       = "rate_limit:%s:%s"   // rate_limit:{user_id}:{endpoint}
	RedisKeyTenantRateLimit = "tenant_limit:%s:%s" // tenant_limit:{tenant_id}:{endpoint}
	RedisKeyIPRateLimit     = "ip_limit:%s:%s"     // ip_limit:{ip_address}:{endpoint}

	// Cache keys
	RedisKeyQueryCache   = "query_cache:%s"   // query_cache:{query_hash}
	RedisKeyUserCache    = "user_cache:%s"    // user_cache:{user_id}
	RedisKeyTenantCache  = "tenant_cache:%s"  // tenant_cache:{tenant_id}
	RedisKeyProductCache = "product_cache:%s" // product_cache:{product_id}
	RedisKeyOrderCache   = "order_cache:%s"   // order_cache:{order_id}

	// Locks (for distributed locking)
	RedisKeyLock = "lock:%s" // lock:{resource_id}

	// Temporary data
	RedisKeyPasswordReset = "pwd_reset:%s"    // pwd_reset:{token}
	RedisKeyEmailVerify   = "email_verify:%s" // email_verify:{token}
	RedisKeyMFACode       = "mfa_code:%s"     // mfa_code:{user_id}
	RedisKeyInviteToken   = "invite:%s"       // invite:{token}

	// Analytics & Metrics
	RedisKeyLoginAttempts = "login_attempts:%s" // login_attempts:{user_id}
	RedisKeyActiveUsers   = "active_users:%s"   // active_users:{tenant_id} -> set
	RedisKeyOnlineUsers   = "online_users"      // online_users -> sorted set

	// Feature flags cache
	RedisKeyFeatureFlag    = "feature_flag:%s"    // feature_flag:{flag_key}
	RedisKeyTenantFeatures = "tenant_features:%s" // tenant_features:{tenant_id}

	// Config cache
	RedisKeyServiceConfig = "config:%s:%s" // config:{service_name}:{environment}
)
