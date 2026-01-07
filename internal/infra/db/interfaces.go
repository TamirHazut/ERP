package db

//go:generate mockgen -destination=mock/mock_db_handler.go -package=mock erp.localhost/internal/infra/db DBHandler

type DBHandler interface {
	Close() error
	Create(db string, data any, opts ...map[string]any) (string, error)
	FindOne(db string, filter map[string]any) (any, error)
	FindAll(db string, filter map[string]any) ([]any, error)
	Update(db string, filter map[string]any, data any, opts ...map[string]any) error
	Delete(db string, filter map[string]any) error
}
