package common_test

import (
	"github.com/Theknighttron/Xtty/internal/common"
	"testing"
)

func TestEncryptionDecryption(t *testing.T) {
	// Generate a key pair
	privateKey, publicKey, err := common.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Original message
	originalMessage := []byte("This is a secret message for testing")

	// Encrypt the message
	encryptedMessage, encryptedKey, err := common.EncryptMessage(originalMessage, publicKey)
	if err != nil {
		t.Fatalf("Failed to encrypt message: %v", err)
	}

	// Decrypt the message
	decryptedMessage, err := common.DecryptMessage(encryptedMessage, encryptedKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to decrypt message: %v", err)
	}

	// Compare the original and decrypted messages
	if string(decryptedMessage) != string(originalMessage) {
		t.Errorf("Decrypted message doesn't match the original message")
	}
}

func TestSignatureVerification(t *testing.T) {
	// Generate a key pair
	privateKey, publicKey, err := common.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Message to sign
	message := []byte("This message needs to be signed for authenticity")

	// Sign the message
	signature, err := common.SignMessage(message, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Verify the signature
	err = common.VerifySignature(message, signature, publicKey)
	if err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}

	// Try to verify with tampered message
	tamperedMessage := []byte("This message has been tampered with")
	err = common.VerifySignature(tamperedMessage, signature, publicKey)
	if err == nil {
		t.Errorf("Verification of tampered message should fail")
	}
}
