# MongoDB

## Structure

### Collection: auth_db 
- tenants
- users
- roles
- permissions
### Collection: core_db
- products
- orders
- vendors
- inventory
### Collection: config_db
- configurations
- environment_settings
- feature_flags

# Redis

## Structure

### Auth:
- sessions:{session_id} → user data
- tokens:{token_id} → token metadata
- refresh_tokens:{user_id} → refresh token

### Gateway (caching):
- query_cache:{query_hash} → response
- rate_limit:{user_id} → request count

