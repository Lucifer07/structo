package cha

import (
	"encoding/base64"

	"github.com/google/uuid"
	"golang.org/x/crypto/chacha20"
)

type EncryptData struct {
	nonce []byte
	key   []byte
}

func NewEncryptor() *EncryptData {
	nonce := make([]byte, chacha20.NonceSize)
	key, _ := uuid.NewV7()

	return &EncryptData{
		key:   []byte(key.String()[0:8] + key.String()[9:13] + key.String()[14:18] + key.String()[19:23] + key.String()[24:]),
		nonce: nonce,
	}
}

func (e *EncryptData) Encrypt(plaintextBytes []byte) (string, error) {

	cipher, err := chacha20.NewUnauthenticatedCipher(e.key, e.nonce)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plaintextBytes))
	cipher.XORKeyStream(ciphertext, plaintextBytes)

	// Encode ciphertext to Base64
	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)

	return encodedCiphertext, nil
}

// Decrypt using ChaCha20
func (e *EncryptData) Decrypt(encodedCiphertext string) (string, error) {

	cipher, err := chacha20.NewUnauthenticatedCipher(e.key, e.nonce)
	if err != nil {
		return "", err
	}

	// Decode Base64 ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}
