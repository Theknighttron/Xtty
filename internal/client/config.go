package client

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/Theknighttron/Xtty/internal/common"
)

// Config for client configurations
type Config struct {
	Username   string `json:"username"`
	ServerHost string `json:"server_host"`
	ServerPort int    `json:"server_port"`
	PrivateKey []byte `json:"private_key"`
	PublicKey  []byte `json:"public_key"`
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

// Create a new configurations with a new key pairs
func CreateNewConfig(username, serverHost string, serverPort int) (*Config, error) {
	// Generate new keypair
	privateKey, publicKey, err := common.GenerateKeyPair(2048)
	if err != nil {
		return nil, err
	}

	// Encode keys to PEM
	privateKeyPEM := common.EncodePrivateKeyToPEM(privateKey)
	publicKeyPEM, err := common.EncodePublicKeyToPEM(publicKey)
	if err != nil {
		return nil, err
	}

	// Create config
	config := &Config{
		Username:   username,
		ServerHost: serverHost,
		ServerPort: serverPort,
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
	}

	return config, nil
}

// Return the default path for the config file
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return filepath.Join(homeDir, ".xtty", "config.json")
}
