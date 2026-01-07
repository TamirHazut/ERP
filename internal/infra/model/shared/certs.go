package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	CACertName = "ca-cert.pem"
	CertName   = "cert.pem"
	KeyName    = "key.pem"
)

type Certs struct {
	CACert string `bson:"ca_cert" json:"ca_cert"`
	Cert   string `bson:"cert" json:"cert"`
	Key    string `bson:"key" json:"key"`
}

func NewCerts() *Certs {
	// 1. Get absolute path of the current file's directory
	_, filename, _, ok := runtime.Caller(1) // get the file of the function who called this function ("NewCerts")
	if !ok {
		return nil
	}
	relativePath, err := getRelativeDir(filename)
	if err != nil {
		return nil
	}
	certsPath := fmt.Sprintf("%s/../resources/certs", relativePath)
	return &Certs{
		CACert: fmt.Sprintf("%s/%s", certsPath, CACertName),
		Cert:   fmt.Sprintf("%s/%s", certsPath, CertName),
		Key:    fmt.Sprintf("%s/%s", certsPath, KeyName),
	}
}

func (c *Certs) IsValidCerts() bool {
	if c.CACert == "" || c.Cert == "" || c.Key == "" {
		return false
	}
	// Check if files exists and are accessable
	files := []string{c.CACert, c.Cert, c.Key}
	for _, filename := range files {
		_, err := os.Stat(filename)
		if err != nil {
			return false
		}
	}

	return true
}

func getRelativeDir(filename string) (string, error) {
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
