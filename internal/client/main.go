package client

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// RunClient runs the client
func RunClient() error {
	// Define command line flags
	serverHost := flag.String("host", "localhost", "Server host")
	serverPort := flag.Int("port", 8080, "Server port")
	username := flag.String("username", "", "Username")
	configPath := flag.String("config", GetDefaultConfigPath(), "Path to config file")
	createConfig := flag.Bool("create-config", false, "Create a new config file")

	// Parse flags
	flag.Parse()

	// Check if we need to create a new config
	if *createConfig {
		if *username == "" {
			return fmt.Errorf("username is required when creating a new config")
		}

		// Create new config
		config, err := CreateNewConfig(*username, *serverHost, *serverPort)
		if err != nil {
			return fmt.Errorf("failed to create new config: %v", err)
		}

		// Save config
		if err := SaveConfig(config, *configPath); err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}

		fmt.Printf("Created new config file at %s\n", *configPath)
		return nil
	}

	// Load config
	config, err := LoadConfig(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found, use --create-config to create one")
		}
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Override config with command line flags if provided
	if *serverHost != "localhost" {
		config.ServerHost = *serverHost
	}
	if *serverPort != 8080 {
		config.ServerPort = *serverPort
	}
	if *username != "" {
		config.Username = *username
	}

	// Create client
	client, err := NewClient(config, nil)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Connect to server
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	// Create UI
	ui := NewUI(client)

	// Set up log file
	logDir := filepath.Dir(*configPath)
	logFile, err := os.OpenFile(filepath.Join(logDir, "xtty-client.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Run UI
	if err := ui.Run(); err != nil {
		return fmt.Errorf("UI error: %v", err)
	}

	// Disconnect from server
	client.Disconnect()

	return nil
}
