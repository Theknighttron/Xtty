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
	roomCode := flag.String("room", "", "Room code to join (optional)")

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

	// Load config if it exists
	var config *Config
	var err error

	if _, err := os.Stat(*configPath); err == nil {
		config, err = LoadConfig(*configPath)
		if err != nil {
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
	} else {
		// Use defaults if no config file
		config = &Config{
			ServerHost: *serverHost,
			ServerPort: *serverPort,
			Username:   *username,
		}

		if config.Username == "" {
			return fmt.Errorf("username is required (use --username flag or create a config file)")
		}
	}

	// Set up log file
	logDir := filepath.Dir(*configPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "xtty-client.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Create user
	user := NewUser()

	// Generate key pair
	if err := user.GenerateKeyPair(); err != nil {
		return fmt.Errorf("failed to generate key pair: %v", err)
	}

	// Determine room code
	useRoomCode := *roomCode
	if useRoomCode == "" {
		useRoomCode = GenerateRoomCode()
	}

	// Construct server URL
	serverURL := fmt.Sprintf("ws://%s:%d", config.ServerHost, config.ServerPort)

	// Connect to server
	if err := user.Connect(serverURL, useRoomCode); err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	// Exchange keys
	if err := user.SendKeyExchange(); err != nil {
		return fmt.Errorf("failed to exchange keys: %v", err)
	}

	// Create UI
	ui := NewUI(user)

	// Run UI
	if err := ui.Run(); err != nil {
		return fmt.Errorf("UI error: %v", err)
	}

	// Clean up
	user.Cleanup()

	return nil
}
