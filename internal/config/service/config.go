package service

import (
	shared_models "erp.localhost/internal/shared/models"
)

type ConfigService struct {
	configs map[shared_models.Module]ConfigService
}
