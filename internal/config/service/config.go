package service

import shared_models "erp.localhost/internal/infra/model/shared"

type ConfigService struct {
	configs map[shared_models.Module]ConfigService
}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}
