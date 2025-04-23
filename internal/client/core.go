package client

import (
	"crypto/rand"
	"math/big"
)

const (
	roomCodeLength = 6
	letterBytes    = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghfnqrt23456789"
)

type Client struct {
}

// Generate secure random room codes
func GenerateRoomCode() string {
	b := make([]byte, roomCodeLength)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		b[i] = letterBytes[n.Int64()]
	}
	return string(b)

}

func NewClient() *Client {
	return &Client{
		Done: make(chan struct{}),
	}
}
