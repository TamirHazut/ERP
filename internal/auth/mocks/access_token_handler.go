package mocks

import (
	redis_models "erp.localhost/internal/db/redis/models"
)

// MockAccessTokenKeyHandler is a mock implementation of AccessTokenKeyHandler for testing
type MockAccessTokenKeyHandler struct {
	StoreFunc     func(tenantID string, tokenID string, metadata redis_models.TokenMetadata) error
	GetFunc       func(tenantID string, tokenID string) (*redis_models.TokenMetadata, error)
	ValidateFunc  func(tenantID string, tokenID string) (*redis_models.TokenMetadata, error)
	RevokeFunc    func(tenantID string, tokenID string, revokedBy string) error
	RevokeAllFunc func(tenantID string, userID string, revokedBy string) error
	DeleteFunc    func(tenantID string, tokenID string) error
}

func (m *MockAccessTokenKeyHandler) Store(tenantID string, tokenID string, metadata redis_models.TokenMetadata) error {
	if m.StoreFunc != nil {
		return m.StoreFunc(tenantID, tokenID, metadata)
	}
	return nil
}

func (m *MockAccessTokenKeyHandler) Get(tenantID string, tokenID string) (*redis_models.TokenMetadata, error) {
	if m.GetFunc != nil {
		return m.GetFunc(tenantID, tokenID)
	}
	return nil, nil
}

func (m *MockAccessTokenKeyHandler) Validate(tenantID string, tokenID string) (*redis_models.TokenMetadata, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(tenantID, tokenID)
	}
	return nil, nil
}

func (m *MockAccessTokenKeyHandler) Revoke(tenantID string, tokenID string, revokedBy string) error {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(tenantID, tokenID, revokedBy)
	}
	return nil
}

func (m *MockAccessTokenKeyHandler) RevokeAll(tenantID string, userID string, revokedBy string) error {
	if m.RevokeAllFunc != nil {
		return m.RevokeAllFunc(tenantID, userID, revokedBy)
	}
	return nil
}

func (m *MockAccessTokenKeyHandler) Delete(tenantID string, tokenID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(tenantID, tokenID)
	}
	return nil
}
