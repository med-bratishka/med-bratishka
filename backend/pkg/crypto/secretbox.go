package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

type SecretBox struct {
	aead cipher.AEAD
}

func NewSecretBox(key string) (*SecretBox, error) {
	if key == "" {
		return nil, fmt.Errorf("empty secret box key")
	}
	sum := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SecretBox{aead: aead}, nil
}

func (b *SecretBox) EncryptString(plain string, aad []byte) (string, error) {
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := b.aead.Seal(nil, nonce, []byte(plain), aad)
	payload := append(nonce, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func (b *SecretBox) DecryptString(encoded string, aad []byte) (string, error) {
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	nonceSize := b.aead.NonceSize()
	if len(payload) <= nonceSize {
		return "", fmt.Errorf("invalid ciphertext")
	}
	plain, err := b.aead.Open(nil, payload[:nonceSize], payload[nonceSize:], aad)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
