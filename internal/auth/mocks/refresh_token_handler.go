package mocks

import (
	"erp.localhost/internal/auth/models"
)

// MockRefreshTokenKeyHandler is a mock implementation of RefreshTokenKeyHandler for testing
type MockRefreshTokenKeyHandler struct {
	StoreFunc          func(tenantID string, userID string, tokenID string, refreshToken models.RefreshToken) error
	GetOneFunc         func(tenantID string, userID string, tokenID string) (*models.RefreshToken, error)
	GetAllFunc         func(tenantID string, userID string) ([]models.RefreshToken, error)
	ValidateFunc       func(tenantID string, userID string, tokenID string) (*models.RefreshToken, error)
	RevokeFunc         func(tenantID string, userID string, tokenID string, revokedBy string) error
	RevokeAllFunc      func(tenantID string, userID string, revokedBy string) error
	UpdateLastUsedFunc func(tenantID string, userID string, tokenID string) error
	DeleteFunc         func(tenantID string, userID string, tokenID string) error
}

func (m *MockRefreshTokenKeyHandler) Store(tenantID string, userID string, tokenID string, refreshToken models.RefreshToken) error {
	if m.StoreFunc != nil {
		return m.StoreFunc(tenantID, userID, tokenID, refreshToken)
	}
	return nil
}

func (m *MockRefreshTokenKeyHandler) GetOne(tenantID string, userID string, tokenID string) (*models.RefreshToken, error) {
	if m.GetOneFunc != nil {
		return m.GetOneFunc(tenantID, userID, tokenID)
	}
	return nil, nil
}

func (m *MockRefreshTokenKeyHandler) GetAll(tenantID string, userID string) ([]models.RefreshToken, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(tenantID, userID)
	}
	return nil, nil
}

func (m *MockRefreshTokenKeyHandler) Validate(tenantID string, userID string, tokenID string) (*models.RefreshToken, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(tenantID, userID, tokenID)
	}
	return nil, nil
}

func (m *MockRefreshTokenKeyHandler) Revoke(tenantID string, userID string, tokenID string, revokedBy string) error {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(tenantID, userID, tokenID, revokedBy)
	}
	return nil
}

func (m *MockRefreshTokenKeyHandler) RevokeAll(tenantID string, userID string, revokedBy string) error {
	if m.RevokeAllFunc != nil {
		return m.RevokeAllFunc(tenantID, userID, revokedBy)
	}
	return nil
}

func (m *MockRefreshTokenKeyHandler) UpdateLastUsed(tenantID string, userID string, tokenID string) error {
	if m.UpdateLastUsedFunc != nil {
		return m.UpdateLastUsedFunc(tenantID, userID, tokenID)
	}
	return nil
}

func (m *MockRefreshTokenKeyHandler) Delete(tenantID string, userID string, tokenID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(tenantID, userID, tokenID)
	}
	return nil
}
