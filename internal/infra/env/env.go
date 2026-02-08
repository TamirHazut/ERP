package env

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	infra_error "erp.localhost/internal/infra/error"
)

// getEnv gets an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnv gets an environment variable or returns a default value
func GetEnvAsInt(key string, defaultValue int) int {
	value := GetEnv(key, "")
	if valInt, err := strconv.Atoi(value); err == nil {
		return valInt
	}
	return defaultValue
}

func LoadEnvironmentVariablesFromFile(filePath string) (map[string]string, *infra_error.AppError) {
	envVariables := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envVariables[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	if err := scanner.Err(); err != nil {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	return envVariables, nil
}
