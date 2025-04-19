package main

import (
	"log"

	"github.com/Theknighttron/Xtty/internal/client"
)

func main() {
	if err := client.RunClient(); err != nil {
		log.Fatalf("Client error: %v", err)
	}
}
