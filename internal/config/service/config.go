package service

import model_shared "erp.localhost/internal/infra/model/shared"

type ConfigService struct {
	configs map[model_shared.Module]ConfigService
}

func NewConfigService() *ConfigService {
	return &ConfigService{
		configs: make(map[model_shared.Module]ConfigService),
	}
}
