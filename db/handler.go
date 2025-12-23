package db

type DBHandler interface {
	Create(db string, data any) (string, error)
	Find(db string, filter map[string]any) ([]any, error)
	Update(db string, id string, data any) error
	Delete(db string, id string) error
}
