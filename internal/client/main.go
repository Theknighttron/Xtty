package client

import (
	"flag"
	"fmt"
	"time"
)

// RunClient handles the complete client lifecycle
func RunClient() error {
	// Define command line flags
	joinCode := flag.String("join", "", "Room code to join")
	username := flag.String("username", "", "Your username")
	flag.Parse()

	// Validate username
	if *username == "" {
		return fmt.Errorf("username is required (use -username YOURNAME)")
	}

	// Initialize client
	c := NewUser(*username)
	defer c.Cleanup()

	var roomCode string
	if *joinCode == "" {
		// Create new room
		roomCode = GenerateRoomCode()
		fmt.Printf("Your room code: %s\n", roomCode)
		fmt.Println("Share this with your peer to connect")
	} else {
		// Join existing room
		roomCode = *joinCode
		fmt.Printf("Joining room: %s\n", roomCode)
	}

	// Generate encryption keys
	if err := c.GenerateKeyPair(); err != nil {
		return fmt.Errorf("failed to generate keys: %v", err)
	}

	// Connect to server
	serverURL := "ws://localhost:8080"
	if err := c.Connect(serverURL, roomCode); err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	// If joining existing room, wait for key exchange
	if *joinCode != "" {
		select {
		case <-c.KeyExchangeDone:
			// Key exchange completed
		case <-time.After(10 * time.Second):
			return fmt.Errorf("timed out waiting for peer")
		}
	}

	// Initialize and run UI
	ui := NewUI(c)
	if err := ui.Run(); err != nil {
		return fmt.Errorf("ui error: %v", err)
	}

	return nil
}
