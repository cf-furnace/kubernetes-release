package helpers

import (
	"encoding/json"
	"errors"
	"os"

	"k8s.io/kubernetes/pkg/client/restclient"
)

type Config struct {
	APIServer                 string `json:"api_server"`
	Username                  string `json:"username"`
	Password                  string `json:"password"`
	SkipCertificateValidation bool   `json:"skip_certificate_validation"`
	CertFile                  string `json:"cert_file"`
	KeyFile                   string `json:"key_file"`
	CAFile                    string `json:"ca_file"`
}

func Load() (*Config, error) {
	path := os.Getenv("CONFIG")
	if path == "" {
		return nil, errors.New("$CONFIG must point to a test configuration file")
	}

	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	config := &Config{}
	if err := json.NewDecoder(configFile).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) ClientConfig() *restclient.Config {
	return &restclient.Config{
		Host:     c.APIServer,
		Username: c.Username,
		Password: c.Password,
		Insecure: c.SkipCertificateValidation,
		TLSClientConfig: restclient.TLSClientConfig{
			CertFile: c.CertFile,
			KeyFile:  c.KeyFile,
			CAFile:   c.CAFile,
		},
	}
}
