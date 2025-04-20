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

var (
	rooms   = make(map[string]*Room) // In-memory room storage
	roomsMu sync.RWMutex             // Protects concurrent access to rooms
)

type Room struct {
	Clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

// Convert http connection into websocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room") // retrieve roomcode from url query string
	if roomCode == "" {
		http.Error(w, "Room code required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Get or Create Room
	roomsMu.Lock()
	room, exists := rooms[roomCode] // check if the room exists
	// if not create the room
	if !exists {
		room = &Room{Clients: make(map[*websocket.Conn]bool)}
		rooms[roomCode] = room
	}
	roomsMu.Unlock()

	// Add client to room
	room.mu.Lock()
	room.Clients[conn] = true
	room.mu.Unlock()

	log.Printf("Client joined room %s", roomCode)

	// Message relay loop
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Broadcast to all other clients in the room
		room.mu.Lock()
		for client := range room.Clients {
			if client != conn {
				if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.Printf("Failed to relay message: %v", err)
					delete(room.Clients, client)
				}
			}
		}
		room.mu.Unlock()
	}

	room.mu.Lock()
	delete(room.Clients, conn)
	// if the room is empty close the websocket connection
	if len(room.Clients) == 0 {
		roomsMu.Lock()
		delete(rooms, roomCode)
		roomsMu.Unlock()
	}
	room.mu.Unlock()
	conn.Close()
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
		return
	}

	// Validate the username
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Check if the username is already taken
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
