package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

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
	// TODO: Implement user registration
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Registration not implemented yet"))
}

func (s *Server) HandleStatusCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server is running"))
}

