package client

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/gorilla/websocket"
)

const (
	roomCodeLength = 6
	letterBytes    = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // No confusing chars
)

type User struct {
	Conn            *websocket.Conn
	RoomCode        string
	KeyPair         *rsa.PrivateKey
	PeerPubKey      *rsa.PublicKey
	Messages        []Message
	Done            chan struct{}
	KeyExchangeDone chan struct{}
	Username        string
}

type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Sent      bool      `json:"sent"`
	Sender    string    `json:"sender"`
}

func GenerateRoomCode() string {
	b := make([]byte, roomCodeLength)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		b[i] = letterBytes[n.Int64()]
	}
	return string(b)
}

func NewUser(username string) *User {
	return &User{
		Done:            make(chan struct{}),
		KeyExchangeDone: make(chan struct{}),
		Username:        username,
	}
}

func (c *User) JoinRoom(serverURL, roomCode string) error {
	c.RoomCode = roomCode
	return c.Connect(serverURL, roomCode)
}

func (c *User) GenerateKeyPair() error {
	var err error
	c.KeyPair, err = rsa.GenerateKey(rand.Reader, 2048)
	return err
}

func (c *User) Connect(serverURL, roomCode string) error {
	conn, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("%s/ws?room=%s", serverURL, roomCode),
		nil,
	)
	if err != nil {
		return err
	}
	c.Conn = conn
	c.RoomCode = roomCode

	go c.readPump()
	return nil
}

func (c *User) readPump() {
	defer close(c.Done)

	// Send our public key immediately after connecting
	if err := c.SendKeyExchange(); err != nil {
		log.Printf("Failed to send key exchange: %v, err")
		return
	}

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}

		// Handle different message types
		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		switch data["type"] {
		case "key_exchange":
			c.handleKeyExchange(data)
			close(c.KeyExchangeDone)
		case "message":
			if c.PeerPubKey == nil {
				log.Println("Received message before key exchange")
				continue
			}
			c.handleEncryptedMessage(data)
		}
	}
}

func (c *User) SendKeyExchange() error {
	if c.KeyPair == nil {
		return fmt.Errorf("no key pair generated")
	}

	pubKeyBytes, err := json.Marshal(c.KeyPair.PublicKey)
	if err != nil {
		return err
	}

	msg := map[string]interface{}{
		"type": "key_exchange",
		"key":  string(pubKeyBytes),
	}

	return c.Conn.WriteJSON(msg)
}

func (c *User) handleKeyExchange(data map[string]interface{}) {
	keyStr, ok := data["key"].(string)
	if !ok {
		log.Println("Invalid key format")
		return
	}

	var pubKey rsa.PublicKey
	if err := json.Unmarshal([]byte(keyStr), &pubKey); err != nil {
		log.Printf("Failed to parse public key: %v", err)
		return
	}

	c.PeerPubKey = &pubKey
	log.Println("Peer public key received and verified")

	// Send confirmation message
	confirmation := map[string]interface{}{
		"type": "key_confirm",
		"from": c.Username,
	}
	c.Conn.WriteJSON(confirmation)
}

func (c *User) SendMessage(content string) error {
	if c.PeerPubKey == nil {
		select {
		case <-c.KeyExchangeDone:
			// Keys exchanged, continue
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout waiting for key exchange")
		}
	}

	encrypted, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		c.PeerPubKey,
		[]byte(content),
	)
	if err != nil {
		return err
	}

	msg := map[string]interface{}{
		"type":    "message",
		"content": encrypted,
	}

	c.Messages = append(c.Messages, Message{
		Content:   content,
		Timestamp: time.Now(),
		Sent:      true,
	})

	return c.Conn.WriteJSON(msg)
}

func (c *User) handleEncryptedMessage(data map[string]interface{}) {
	encrypted, ok := data["content"].(string)
	if !ok {
		log.Println("Invalid message format")
		return
	}

	decrypted, err := rsa.DecryptPKCS1v15(
		rand.Reader,
		c.KeyPair,
		[]byte(encrypted),
	)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		return
	}

	c.Messages = append(c.Messages, Message{
		Content:   string(decrypted),
		Timestamp: time.Now(),
		Sent:      false,
	})
}

func (c *User) Cleanup() {
	if c.Conn != nil {
		c.Conn.Close()
	}
	c.KeyPair = nil
	c.PeerPubKey = nil
	c.Messages = nil
}
