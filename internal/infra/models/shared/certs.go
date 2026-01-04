package shared_models

import "os"

type Certs struct {
	CACert string `bson:"ca_cert" json:"ca_cert"`
	Certs  string `bson:"cert" json:"cert"`
	Key    string `bson:"key" json:"key"`
}

func (c *Certs) IsValidCerts() bool {
	if c.CACert == "" || c.Certs == "" || c.Key == "" {
		return false
	}
	// Check if files exists and are accessable
	files := []string{c.CACert, c.Certs, c.Key}
	for _, filename := range files {
		_, err := os.Stat(filename)
		if err != nil {
			return false
		}
	}

	return true
}
