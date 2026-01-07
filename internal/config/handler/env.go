package handler

import (
	"os"

	"erp.localhost/internal/config/env"
	"erp.localhost/internal/infra/logging/logger"
	model_shared "erp.localhost/internal/infra/model/shared"
)

var (
	envHandler *EnvHandler = initEnvHandler()
)

type EnvHandler struct {
	envVariables map[string]string
	logger       logger.Logger
}

func initEnvHandler() *EnvHandler {
	logger := logger.NewBaseLogger(model_shared.ModuleConfig)
	envFiles, err := os.ReadDir("configs/env")
	if err != nil {
		logger.Error("Failed to read env files", "error", err)
		return nil
	}
	envVariables := make(map[string]string)
	for _, envFile := range envFiles {
		envVariables := env.LoadEnvironmentVariablesFromFile(envFile.Name(), logger)
		for key, value := range envVariables {
			envVariables[key] = value
		}
	}
	return &EnvHandler{
		envVariables: envVariables,
		logger:       logger,
	}
}

func GetEnvHandler(envName string) string {
	if envHandler == nil {
		return "env"
	}
	envVariable, ok := envHandler.envVariables[envName]
	if !ok {
		env := getEnvFromOS(envName)
		if env == "" {
			envHandler.logger.Debug("Env variable not found", "env", envName)
			return env
		}
	}

	return envVariable
}

func getEnvFromOS(envName string) string {
	if env := os.Getenv(envName); env != "" {
		return env
	}
	return ""
}
