package service

import "erp.localhost/infra/model/shared"

type ConfigService struct {
	configs map[shared.Module]ConfigService
}

func NewConfigService() *ConfigService {
	return &ConfigService{
		configs: make(map[shared.Module]ConfigService),
	}
}
