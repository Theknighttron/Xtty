package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Theknighttron/Xtty/internal/common"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client
type Client struct {
	conn           *websocket.Conn
	config         *Config
	privateKey     interface{}
	messageHandler func(message *common.Message)
	shutdownCh     chan struct{}
}

// NewClient creates a new WebSocket client
func NewClient(config *Config, messageHandler func(message *common.Message)) (*Client, error) {
	// Parse private key
	privateKey, err := common.ParsePrivateKeyFromPEM(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	client := &Client{
		config:         config,
		privateKey:     privateKey,
		messageHandler: messageHandler,
		shutdownCh:     make(chan struct{}),
	}

	return client, nil
}

// Connects to the WebSocket server
func (c *Client) Connect() error {
	// Create WebSocket URL
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", c.config.ServerHost, c.config.ServerPort),
		Path:   "/ws",
	}

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %v", err)
	}

	c.conn = conn

	// Start handling messages
	go c.handleMessages()

	// Send auth packet
	authPacket := common.Packet{
		Type:      "auth",
		Timestamp: time.Now(),
		Data: map[string]string{
			"username":   c.config.Username,
			"public_key": string(c.config.PublicKey),
		},
	}

	if err := c.SendPacket(authPacket); err != nil {
		return fmt.Errorf("failed to send auth packet: %v", err)
	}

	log.Println("Connected to WebSocket server")

	return nil
}

// SendPacket sends a packet to the server
func (c *Client) SendPacket(packet common.Packet) error {
	data, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %v", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	return nil
}

// SendMessage sends a message to a user
func (c *Client) SendMessage(recipientID string, messageType common.MessageType, content string) error {
	// Get recipient's public key from server (to be implemented)
	// For now, we'll assume we already have it and it's hardcoded
	// In a real implementation, we'd request it from the server or cache it

	// Encrypt message
	// This would require fetching the recipient's public key first
	// For now, we'll send a simple unencrypted message

	message := common.Message{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		SenderID:    c.config.Username,
		RecipientID: recipientID,
		Type:        messageType,
		Timestamp:   time.Now(),
		Content:     content,
	}

	packet := common.Packet{
		Type:      "message",
		Timestamp: time.Now(),
		Data:      message,
	}

	return c.SendPacket(packet)
}

// handleMessages handles incoming WebSocket messages
func (c *Client) handleMessages() {
	defer c.conn.Close()

	for {
		select {
		case <-c.shutdownCh:
			return
		default:
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				return
			}

			var packet common.Packet
			if err := json.Unmarshal(data, &packet); err != nil {
				log.Printf("Error unmarshaling packet: %v", err)
				continue
			}

			switch packet.Type {
			case "message":
				var message common.Message
				messageData, err := json.Marshal(packet.Data)
				if err != nil {
					log.Printf("Error marshaling message data: %v", err)
					continue
				}

				if err := json.Unmarshal(messageData, &message); err != nil {
					log.Printf("Error unmarshaling message: %v", err)
					continue
				}

				if c.messageHandler != nil {
					c.messageHandler(&message)
				}
			case "error":
				log.Printf("Error from server: %v", packet.Data)
			default:
				log.Printf("Received packet of type %s", packet.Type)
			}
		}
	}
}

// Disconnect disconnects from the WebSocket server
func (c *Client) Disconnect() {
	close(c.shutdownCh)
	if c.conn != nil {
		c.conn.Close()
	}
}
