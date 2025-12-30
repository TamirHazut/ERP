package mongo

//go:generate mockgen -destination=mocks/mock_collection_handler.go -package=mocks erp.localhost/internal/db/mongo CollectionHandler

type CollectionHandler[T any] interface {
	Create(item T) (string, error)
	FindOne(filter map[string]any) (*T, error)
	FindAll(filter map[string]any) ([]T, error)
	Update(filter map[string]any, item T) error
	Delete(filter map[string]any) error
}
