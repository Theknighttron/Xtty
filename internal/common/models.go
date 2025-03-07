package common

import "time"

// MessageType defines different types of message in the system
type MessageType int

const (
	// Message types
	TypeText MessageType = iota
	TypeFriendRequest
	TypeFriendAccept
	TypeFriendReject
	TypeStatusUpdate
	TypeReadReceipt
	TypeTypingIndicator
	TypeSystemNotification
)

// userstatus represents the online status of the user
type UserStatus int

const (
	StatusOffline UserStatus = iota
	StatusOnline
	StatusAway
	StatusBusy
)

// User represents a user in the system
type User struct {
	Username   string     `json:"username"`
	PublicKey  []byte     `json:"public_key"`
	Status     UserStatus `json:"status"`
	LastSeen   time.Time  `json:"last_seen"`
	FriendList []string   `json:"friend_list,omitempty"`
}

// Message represents a message in the system (both encrypted and system messages)
type Message struct {
	ID          string      `json:"id"`
	SenderID    string      `json:"sender_id"`
	RecipientID string      `json:"recipient_id"`
	Type        MessageType `json:"type"`
	Timestamp   time.Time   `json:"timestamp"`

	// For encrypted Messages
	EncryptedContent []byte `json:"encrypted_content,omitempty"`
	EncryptedKey     []byte `json:"encrypted_key,omitempty"`
	Signature        []byte `json:"signature,omitempty"`

	// For system messages
	Content string `json:"content,omitempty"`
}

// FriendRequests represents a friendship request
type FriendRequests struct {
	FromUser  string    `json:"from_user"`
	ToUser    string    `json:"to_user"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// serverConfig holds configuration for the server
type ServerConfi struct {
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	MessageTTL        time.Duration `json:"message_ttl"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}
