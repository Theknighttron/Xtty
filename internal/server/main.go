package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Theknighttron/Xtty/internal/common"
	"github.com/gorilla/websocket"
)

type Server struct {
	config      common.ServerConfig
	clients     map[*websocket.Conn]bool
	clientsLock sync.RWMutex
	users       map[string]common.User
	usersLock   sync.RWMutex
}

func NewServer(config common.ServerConfig) *Server {
	return &Server{
		config:  config,
		clients: make(map[*websocket.Conn]bool),
		users:   make(map[string]common.User),
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	s.clientsLock.Lock()
	s.clients[conn] = true
	s.clientsLock.Unlock()

	log.Println("New WebSocket connection established")

	// Handle incoming messages
	go s.handleMessages(conn)
}

func (s *Server) handleMessages(conn *websocket.Conn) {
	defer func() {
		s.clientsLock.Lock()
		delete(s.clients, conn)
		s.clientsLock.Unlock()
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		var packet common.Packet
		if err := json.Unmarshal(message, &packet); err != nil {
			log.Printf("Error unmarshaling packet: %v", err)
			continue
		}

		log.Printf("Received packet: %+v", packet)
	}
}

func (s *Server) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username  string `json:"username"`
		PublicKey []byte `json:"public_key"`
	}

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	// Validate the username
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	if _, exists := s.users[req.Username]; exists {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	// Create a new user
	user := common.User{
		Username:  req.Username,
		PublicKey: req.PublicKey,
		Status:    common.StatusOffline,
		LastSeen:  time.Now(),
	}

	// Store the user
	s.users[req.Username] = user

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// AuthenticateWebSocket authenticates WebSocket connections
func (s *Server) AuthenticateWebSocket(conn *websocket.Conn) (string, error) {
	// Read the first message (authentication token)
	_, message, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}

	var auth struct {
		Username string `json:"username"`
		Token    string `json:"token"`
	}
	if err := json.Unmarshal(message, &auth); err != nil {
		return "", err
	}

	// Validate the token (for now, just check if the user exists)
	s.usersLock.RLock()
	defer s.usersLock.RUnlock()
	if _, exists := s.users[auth.Username]; !exists {
		return "", errors.New("user not found")
	}

	return auth.Username, nil
}

func (s *Server) HandleStatusCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server is running"))
}
