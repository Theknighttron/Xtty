package main

import (
	"fmt"
	"log"

	"github.com/Theknighttron/Xtty/internal/client"
)

func main() {
	// Create a new user
	user := client.NewUser()

	// Generate room code and keys
	roomCode := client.GenerateRoomCode()
	if err := user.GenerateKeyPair(); err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	fmt.Printf("Your room code: %s\n", roomCode)
	fmt.Println("Share this with your peer to connect")

	// Connect to server
	serverURL := "ws://localhost:8080"
	if err := user.Connect(serverURL, roomCode); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	// Send public key
	if err := user.SendKeyExchange(); err != nil {
		log.Fatalf("Key exchange failed: %v", err)
	}

	defer user.Cleanup()

	// Start UI
	ui := client.NewUI(user)
	if err := ui.Run(); err != nil {
		log.Printf("UI error: %v", err)
	}
}
