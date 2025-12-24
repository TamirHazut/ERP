package db

// Generic Repository
type Repository[T any] struct {
	dbHandler DBHandler
	dbName    string
}

func NewRepository[T any](dbHandler DBHandler, dbName string) (*Repository[T], error) {
	return &Repository[T]{
		dbHandler: dbHandler,
		dbName:    dbName,
	}, nil
}

func (r *Repository[T]) Create(item T) (string, error) {
	return r.dbHandler.Create(r.dbName, item)
}

func (r *Repository[T]) Find(filter map[string]any) ([]T, error) {
	items, err := r.dbHandler.Find(r.dbName, filter)
	if err != nil {
		return nil, err
	}
	res := []T{}
	for _, item := range items {
		res = append(res, item.(T))
	}
	return res, nil
}

func (r *Repository[T]) Update(id string, item T) error {
	return r.dbHandler.Update(r.dbName, id, item)
}

func (r *Repository[T]) Delete(id string) error {
	return r.dbHandler.Delete(r.dbName, id)
}
