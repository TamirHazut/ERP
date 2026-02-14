package db

import (
	infra_error "erp.localhost/infra/error"
)

//go:generate mockgen -destination=mock/mock_db_handler.go -package=mock erp.localhost/infra/db DBHandler

type DBHandler interface {
	Close() *infra_error.AppError
	Create(db string, data any, opts ...map[string]any) (string, *infra_error.AppError)
	FindOne(db string, filter map[string]any, result any) *infra_error.AppError
	FindAll(db string, filter map[string]any, result any) *infra_error.AppError
	Update(db string, filter map[string]any, data any, opts ...map[string]any) *infra_error.AppError
	Delete(db string, filter map[string]any) *infra_error.AppError
}
