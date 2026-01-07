package env

import (
	"os"
	"testing"

	"erp.localhost/internal/infra/logging"
	shared_models "erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
)

var testEnvFile = `
TEST=test
TEST2=test2
TEST3=test3
`

func TestLoadEnvironmentVariablesFromFile(t *testing.T) {
	testCase := []struct {
		name     string
		envFile  string
		expected map[string]string
	}{
		{
			name:     "valid env file",
			envFile:  testEnvFile,
			expected: map[string]string{"TEST": "test", "TEST2": "test2", "TEST3": "test3"},
		},
		{
			name:     "invalid env file",
			envFile:  "invalid.env",
			expected: map[string]string{},
		},
		{
			name:     "empty env file",
			envFile:  "",
			expected: map[string]string{},
		},
		{
			name: "env file with comments",
			envFile: `
		# This is a comment
		TEST=test
		TEST2=test2
		TEST3=test3
		`,
			expected: map[string]string{"TEST": "test", "TEST2": "test2", "TEST3": "test3"},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			tempFile, err := os.CreateTemp("", "test.env")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())
			if _, err := tempFile.WriteString(tc.envFile); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			tempFile.Close()
			envVariables := LoadEnvironmentVariablesFromFile(tempFile.Name(), logging.NewLogger(shared_models.ModuleConfig))
			assert.NotNil(t, envVariables)
			assert.Equal(t, tc.expected, envVariables)
			os.Remove(tempFile.Name())
		})
	}
}
