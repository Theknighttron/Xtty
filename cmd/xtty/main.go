package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Theknighttron/Xtty/internal/common"
	"github.com/Theknighttron/Xtty/internal/server"
)

var (
	host = flag.String("host", "localhost", "Server host")
	port = flag.Int("port", 8080, "Server port")
)

func main() {
	flag.Parse()

	xttyServer := server.NewServer(common.ServerConfig{
		Host:              *host,
		Port:              *port,
		MessageTTL:        7 * 24 * time.Hour,
		HeartbeatInterval: 30 * time.Second,
	})

	// Set up routes
	// http.HandleFunc("/ws", xttyServer.HandleWebSocket)
	http.HandleFunc("/register", xttyServer.HandleRegistration)
	http.HandleFunc("/status", xttyServer.HandleStatusCheck)

	// Create a server with grateful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *host, *port),
		Handler: nil, // use default serverMux
	}

	// channel to listen for Interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Start the server in go routine
	go func() {
		log.Printf("Starting Xtty server on %s:%d", *host, *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Error during server shutdown: %v", err)
	}

	// Wait for the server to finish processing requests
	log.Println("Server gracefully stopped")

}
