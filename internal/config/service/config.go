package service

import shared_models "erp.localhost/internal/infra/models/shared"

type ConfigService struct {
	configs map[shared_models.Module]ConfigService
}
