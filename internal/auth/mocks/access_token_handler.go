package mocks

import (
	redis_models "erp.localhost/internal/db/redis/models"
)

// MockAccessTokenKeyHandler is a mock implementation of AccessTokenKeyHandler for testing
type MockAccessTokenKeyHandler struct {
	StoreFunc     func(tenantID string, userID string, tokenID string, metadata redis_models.TokenMetadata) error
	GetOneFunc    func(tenantID string, userID string, tokenID string) (*redis_models.TokenMetadata, error)
	GetAllFunc    func(tenantID string, userID string) ([]redis_models.TokenMetadata, error)
	ValidateFunc  func(tenantID string, userID string, tokenID string) (*redis_models.TokenMetadata, error)
	RevokeFunc    func(tenantID string, userID string, tokenID string, revokedBy string) error
	RevokeAllFunc func(tenantID string, userID string, revokedBy string) error
	DeleteFunc    func(tenantID string, userID string, tokenID string) error
}

func (m *MockAccessTokenKeyHandler) Store(tenantID string, userID string, tokenID string, metadata redis_models.TokenMetadata) error {
	if m.StoreFunc != nil {
		return m.StoreFunc(tenantID, userID, tokenID, metadata)
	}
	return nil
}

func (m *MockAccessTokenKeyHandler) GetOne(tenantID string, userID string, tokenID string) (*redis_models.TokenMetadata, error) {
	if m.GetOneFunc != nil {
		return m.GetOneFunc(tenantID, userID, tokenID)
	}
	return nil, nil
}

func (m *MockAccessTokenKeyHandler) GetAll(tenantID string, userID string) ([]redis_models.TokenMetadata, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(tenantID, userID)
	}
	return nil, nil
}

func (m *MockAccessTokenKeyHandler) Validate(tenantID string, userID string, tokenID string) (*redis_models.TokenMetadata, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(tenantID, userID, tokenID)
	}
	return nil, nil
}

func (m *MockAccessTokenKeyHandler) Revoke(tenantID string, userID string, tokenID string, revokedBy string) error {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(tenantID, userID, tokenID, revokedBy)
	}
	return nil
}

func (m *MockAccessTokenKeyHandler) RevokeAll(tenantID string, userID string, revokedBy string) error {
	if m.RevokeAllFunc != nil {
		return m.RevokeAllFunc(tenantID, userID, revokedBy)
	}
	return nil
}

func (m *MockAccessTokenKeyHandler) Delete(tenantID string, userID string, tokenID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(tenantID, userID, tokenID)
	}
	return nil
}
