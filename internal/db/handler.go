package db

type DBHandler interface {
	Create(db string, data any) (string, error)
	FindOne(db string, filter map[string]any) (any, error)
	FindAll(db string, filter map[string]any) ([]any, error)
	Update(db string, filter map[string]any, data any) error
	Delete(db string, filter map[string]any) error
}
