package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Theknighttron/Xtty/internal/client"
)

func main() {
	join := flag.String("join", "", "Room code to join")
	username := flag.String("username", "", "Your username")
	flag.Parse()

	if *username == "" {
		fmt.Println("Error: username is required")
		fmt.Println("Usage: go run main.go -username YOURNAME [-join ROOM_CODE]")
		os.Exit(1)
	}

	u := client.NewUser(*username)

	var roomCode string
	if *join == "" {
		roomCode = client.GenerateRoomCode()
		fmt.Printf("Your room code: %s\n", roomCode)
		fmt.Println("Share this with your peer to connect")
	} else {
		roomCode = *join
		fmt.Printf("Joining room: %s\n", roomCode)
	}

	if err := u.GenerateKeyPair(); err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	if err := u.Connect("ws://localhost:8080", roomCode); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer u.Cleanup()

	// Wait for key exchange if joining existing room
	if *join != "" {
		<-u.KeyExchangeDone
	}

	ui := client.NewUI(u)
	if err := ui.Run(); err != nil {
		log.Printf("UI error: %v", err)
	}
}
