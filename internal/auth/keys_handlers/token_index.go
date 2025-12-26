package keyshandlers

import (
	"context"
	"fmt"

	erp_errors "erp.localhost/internal/errors"
	logging "erp.localhost/internal/logging"
	redis_client "github.com/redis/go-redis/v9"
)

var (
	redisContext = context.Background()
)

// TokenIndex manages token indices using Redis Sets
// Key pattern: user_access_tokens:{tenant_id}:{user_id} -> Set of token_ids
// Key pattern: user_refresh_tokens:{tenant_id}:{user_id} -> Set of token_ids
type TokenIndex struct {
	client *redis_client.Client
	logger *logging.Logger
}

// NewTokenIndex creates a new TokenIndex
func NewTokenIndex() *TokenIndex {
	logger := logging.NewLogger(logging.ModuleAuth)
	
	// Get Redis client from a handler (we'll need to access the underlying client)
	// For now, we'll create our own connection
	uri := "redis://:supersecretredis@localhost:6379"
	options, err := redis_client.ParseURL(uri)
	if err != nil {
		logger.Error("Failed to parse Redis URI", "error", err)
		return nil
	}

	client := redis_client.NewClient(options)
	if err := client.Ping(redisContext).Err(); err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return nil
	}

	return &TokenIndex{
		client: client,
		logger: logger,
	}
}

// AddAccessToken adds a token_id to the user's access token index
func (t *TokenIndex) AddAccessToken(tenantID string, userID string, tokenID string) error {
	key := fmt.Sprintf("user_access_tokens:%s:%s", tenantID, userID)
	err := t.client.SAdd(redisContext, key, tokenID).Err()
	if err != nil {
		t.logger.Error("Failed to add access token to index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	
	// Set TTL on the index (match longest token expiry, e.g., 1 hour)
	// This ensures the index is cleaned up even if individual tokens expire
	t.client.Expire(redisContext, key, 2*60*60) // 2 hours
	
	t.logger.Debug("Access token added to index", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// RemoveAccessToken removes a token_id from the user's access token index
func (t *TokenIndex) RemoveAccessToken(tenantID string, userID string, tokenID string) error {
	key := fmt.Sprintf("user_access_tokens:%s:%s", tenantID, userID)
	err := t.client.SRem(redisContext, key, tokenID).Err()
	if err != nil {
		t.logger.Error("Failed to remove access token from index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	
	t.logger.Debug("Access token removed from index", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// GetAccessTokens returns all token_ids for a user's access tokens
func (t *TokenIndex) GetAccessTokens(tenantID string, userID string) ([]string, error) {
	key := fmt.Sprintf("user_access_tokens:%s:%s", tenantID, userID)
	tokenIDs, err := t.client.SMembers(redisContext, key).Result()
	if err != nil {
		if err == redis_client.Nil {
			return []string{}, nil
		}
		t.logger.Error("Failed to get access tokens from index", "error", err, "tenantID", tenantID, "userID", userID)
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	
	return tokenIDs, nil
}

// AddRefreshToken adds a token_id to the user's refresh token index
func (t *TokenIndex) AddRefreshToken(tenantID string, userID string, tokenID string) error {
	key := fmt.Sprintf("user_refresh_tokens:%s:%s", tenantID, userID)
	err := t.client.SAdd(redisContext, key, tokenID).Err()
	if err != nil {
		t.logger.Error("Failed to add refresh token to index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	
	// Set TTL on the index (match longest token expiry, e.g., 7 days)
	t.client.Expire(redisContext, key, 7*24*60*60) // 7 days
	
	t.logger.Debug("Refresh token added to index", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// RemoveRefreshToken removes a token_id from the user's refresh token index
func (t *TokenIndex) RemoveRefreshToken(tenantID string, userID string, tokenID string) error {
	key := fmt.Sprintf("user_refresh_tokens:%s:%s", tenantID, userID)
	err := t.client.SRem(redisContext, key, tokenID).Err()
	if err != nil {
		t.logger.Error("Failed to remove refresh token from index", "error", err, "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	
	t.logger.Debug("Refresh token removed from index", "tenantID", tenantID, "userID", userID, "tokenID", tokenID)
	return nil
}

// GetRefreshTokens returns all token_ids for a user's refresh tokens
func (t *TokenIndex) GetRefreshTokens(tenantID string, userID string) ([]string, error) {
	key := fmt.Sprintf("user_refresh_tokens:%s:%s", tenantID, userID)
	tokenIDs, err := t.client.SMembers(redisContext, key).Result()
	if err != nil {
		if err == redis_client.Nil {
			return []string{}, nil
		}
		t.logger.Error("Failed to get refresh tokens from index", "error", err, "tenantID", tenantID, "userID", userID)
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	
	return tokenIDs, nil
}

// ClearAccessTokens removes all access tokens from the index (used when revoking all)
func (t *TokenIndex) ClearAccessTokens(tenantID string, userID string) error {
	key := fmt.Sprintf("user_access_tokens:%s:%s", tenantID, userID)
	err := t.client.Del(redisContext, key).Err()
	if err != nil {
		t.logger.Error("Failed to clear access tokens index", "error", err, "tenantID", tenantID, "userID", userID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	
	t.logger.Debug("Access tokens index cleared", "tenantID", tenantID, "userID", userID)
	return nil
}

// ClearRefreshTokens removes all refresh tokens from the index (used when revoking all)
func (t *TokenIndex) ClearRefreshTokens(tenantID string, userID string) error {
	key := fmt.Sprintf("user_refresh_tokens:%s:%s", tenantID, userID)
	err := t.client.Del(redisContext, key).Err()
	if err != nil {
		t.logger.Error("Failed to clear refresh tokens index", "error", err, "tenantID", tenantID, "userID", userID)
		return erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	
	t.logger.Debug("Refresh tokens index cleared", "tenantID", tenantID, "userID", userID)
	return nil
}

// Close closes the Redis connection
func (t *TokenIndex) Close() error {
	return t.client.Close()
}

