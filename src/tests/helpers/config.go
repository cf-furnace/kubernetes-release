package helpers

import (
	"encoding/json"
	"errors"
	"os"

	"k8s.io/kubernetes/pkg/client/restclient"
)

type Config struct {
	APIServer string `json:"api_server"`
	CertFile  string `json:"cert_file"`
	KeyFile   string `json:"key_file"`
	CAFile    string `json:"ca_file"`
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
		Host: c.APIServer,
		TLSClientConfig: restclient.TLSClientConfig{
			CertFile: c.CertFile,
			KeyFile:  c.KeyFile,
			CAFile:   c.CAFile,
		},
	}
}
