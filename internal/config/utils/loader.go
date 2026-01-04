package utils

import (
	"bufio"
	"os"
	"strings"

	"erp.localhost/internal/infra/logging"
)

func LoadEnvironmentVariablesFromFile(filePath string, logger *logging.Logger) map[string]string {
	envVariables := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("failed to open file", "error", err)
		return nil
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
			logger.Warn("invalid line in config file", "line", line)
			continue
		}
		envVariables[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	if err := scanner.Err(); err != nil {
		logger.Error("failed to read file", "error", err)
		return nil
	}
	return envVariables
}
