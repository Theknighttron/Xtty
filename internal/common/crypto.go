package common

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

// Create a new RSA key pair
func GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

// Encodes the private key to PEM(Private Enhance Mail) format
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)

	return privateKeyPEM
}

// Encodes the private key to PEM(Private Enhance Mail) format
func EncodePublicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: publicKeyBytes,
		},
	)

	return publicKeyPEM, nil
}

// ParsePrivateKeyFromPEM parses a PEM encoded private key
func ParsePrivateKeyFromPEM(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// ParsePublicKeyFromPEM parses a PEM encoded public key
func ParsePublicKeyFromPEM(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return publicKey, nil
}

// Encrypt a message using hybrid encryption
func EncryptMessage(message []byte, publicKey *rsa.PublicKey) ([]byte, []byte, error) {
	// Generate a random key AES key
	aeskey := make([]byte, 32) // 256 bits

	if _, err := io.ReadFull(rand.Reader, aeskey); err != nil {
		return nil, nil, err
	}

	// Encrypt the AES key with RSA
	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		aeskey,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt the message with AES-GCM
	block, err := aes.NewCipher(aeskey)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	encryptedMessage := aesgcm.Seal(nonce, nonce, message, nil)

	return encryptedMessage, encryptedKey, nil
}

// Decrypts a message using hybrid encryption (RSA + AES)
func DecryptMessage(encryptedMessage, encryptedKey []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	// Decrypt the AES key with RSA
	aesKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Decrypt the message with AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(encryptedMessage) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := encryptedMessage[:nonceSize], encryptedMessage[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Signs a message with a private key
func SignMessage(message []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	// Hash the message
	hash := sha256.Sum256(message)

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(
		rand.Reader,
		privateKey,
		crypto.SHA256,
		hash[:],
	)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// VerifySignature verifies a message signature
func VerifySignature(message, signature []byte, publicKey *rsa.PublicKey) error {
	// Hash the message
	hash := sha256.Sum256(message)

	// Verify the signature
	return rsa.VerifyPKCS1v15(
		publicKey,
		crypto.SHA256,
		hash[:],
		signature,
	)
}
