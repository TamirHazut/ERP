package keyshandlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenIndex(t *testing.T) {
	// Note: This test requires a running Redis instance
	// If Redis is not available, it will return nil
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	require.NotNil(t, index)
	assert.NotNil(t, index.client)
	assert.NotNil(t, index.logger)
}

func TestTokenIndex_AddAccessToken(t *testing.T) {
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer index.Close()

	err := index.AddAccessToken("tenant-1", "user-1", "token-1")
	require.NoError(t, err)

	// Verify it was added
	tokens, err := index.GetAccessTokens("tenant-1", "user-1")
	require.NoError(t, err)
	assert.Contains(t, tokens, "token-1")
}

func TestTokenIndex_RemoveAccessToken(t *testing.T) {
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer index.Close()

	// Add first
	err := index.AddAccessToken("tenant-1", "user-1", "token-1")
	require.NoError(t, err)

	// Remove
	err = index.RemoveAccessToken("tenant-1", "user-1", "token-1")
	require.NoError(t, err)

	// Verify it was removed
	tokens, err := index.GetAccessTokens("tenant-1", "user-1")
	require.NoError(t, err)
	assert.NotContains(t, tokens, "token-1")
}

func TestTokenIndex_GetAccessTokens(t *testing.T) {
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer index.Close()

	// Add multiple tokens
	err := index.AddAccessToken("tenant-1", "user-1", "token-1")
	require.NoError(t, err)
	err = index.AddAccessToken("tenant-1", "user-1", "token-2")
	require.NoError(t, err)

	// Get all tokens
	tokens, err := index.GetAccessTokens("tenant-1", "user-1")
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Contains(t, tokens, "token-1")
	assert.Contains(t, tokens, "token-2")
}

func TestTokenIndex_AddRefreshToken(t *testing.T) {
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer index.Close()

	err := index.AddRefreshToken("tenant-1", "user-1", "refresh-1")
	require.NoError(t, err)

	// Verify it was added
	tokens, err := index.GetRefreshTokens("tenant-1", "user-1")
	require.NoError(t, err)
	assert.Contains(t, tokens, "refresh-1")
}

func TestTokenIndex_RemoveRefreshToken(t *testing.T) {
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer index.Close()

	// Add first
	err := index.AddRefreshToken("tenant-1", "user-1", "refresh-1")
	require.NoError(t, err)

	// Remove
	err = index.RemoveRefreshToken("tenant-1", "user-1", "refresh-1")
	require.NoError(t, err)

	// Verify it was removed
	tokens, err := index.GetRefreshTokens("tenant-1", "user-1")
	require.NoError(t, err)
	assert.NotContains(t, tokens, "refresh-1")
}

func TestTokenIndex_GetRefreshTokens(t *testing.T) {
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer index.Close()

	// Add multiple tokens
	err := index.AddRefreshToken("tenant-1", "user-1", "refresh-1")
	require.NoError(t, err)
	err = index.AddRefreshToken("tenant-1", "user-1", "refresh-2")
	require.NoError(t, err)

	// Get all tokens
	tokens, err := index.GetRefreshTokens("tenant-1", "user-1")
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Contains(t, tokens, "refresh-1")
	assert.Contains(t, tokens, "refresh-2")
}

func TestTokenIndex_ClearAccessTokens(t *testing.T) {
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer index.Close()

	// Add tokens
	err := index.AddAccessToken("tenant-1", "user-1", "token-1")
	require.NoError(t, err)
	err = index.AddAccessToken("tenant-1", "user-1", "token-2")
	require.NoError(t, err)

	// Clear all
	err = index.ClearAccessTokens("tenant-1", "user-1")
	require.NoError(t, err)

	// Verify cleared
	tokens, err := index.GetAccessTokens("tenant-1", "user-1")
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestTokenIndex_ClearRefreshTokens(t *testing.T) {
	index := NewTokenIndex()
	if index == nil {
		t.Skip("Redis not available, skipping test")
	}
	defer index.Close()

	// Add tokens
	err := index.AddRefreshToken("tenant-1", "user-1", "refresh-1")
	require.NoError(t, err)
	err = index.AddRefreshToken("tenant-1", "user-1", "refresh-2")
	require.NoError(t, err)

	// Clear all
	err = index.ClearRefreshTokens("tenant-1", "user-1")
	require.NoError(t, err)

	// Verify cleared
	tokens, err := index.GetRefreshTokens("tenant-1", "user-1")
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

