package client

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config for client configurations
type Config struct {
	Username   string `json:"username"`
	ServerHost string `json:"server_host"`
	ServerPort string `json:"server_port"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// Load the Config from the configurations file
func LoadConfig(configPath string) (*Config, error) {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	// check if the config file exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, errors.New("config file does not exist")
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save the configuration to the config file
func saveConfig(config *Config, configPath string) error {
	// Create config directory if it doesnt exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	// Marshall config to JSON
	data, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}

	// Write config file
	return os.WriteFile(configPath, data, 0600)
}
