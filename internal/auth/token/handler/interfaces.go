package handler

//go:generate mockgen -destination=mock/mock_token_handler.go -package=mock erp.localhost/internal/auth/token/handler TokenHandler

type TokenHandler[T any] interface {
	Store(tenantID string, userID string, tokenID string, value T) error
	GetOne(tenantID string, userID string, tokenID string) (*T, error)
	GetAll(tenantID string, userID string) ([]T, error)
	Validate(tenantID string, userID string, tokenID string) (*T, error)
	Revoke(tenantID string, userID string, tokenID string, revokedBy string) error
	RevokeAll(tenantID string, userID string, revokedBy string) error
	Delete(tenantID string, userID string, tokenID string) error
}
