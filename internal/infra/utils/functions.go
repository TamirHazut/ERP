package utils

import (
	"os"
	"path/filepath"
)

func GetRelativeDir(filename string) (string, error) {
	functionFileDir := filepath.Dir(filename)

	// 2. Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 3. Calculate the relative path from CWD to the function's file directory
	relativePath, err := filepath.Rel(workingDir, functionFileDir)
	if err != nil {
		return "", err
	}

	return relativePath, nil
}
