package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"

	"cascade/internal/application/port"
)

type aesCryptoService struct {
	aesKey []byte
	pepper []byte
	rsaKey *rsa.PrivateKey
}

func NewCryptoService(aesKeyHex string, pepperStr string, rsaPriv *rsa.PrivateKey) (port.CryptoService, error) {
	key, err := hex.DecodeString(aesKeyHex)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("AES key must be exactly 32 bytes for AES-256")
	}

	return &aesCryptoService{
		aesKey: key,
		pepper: []byte(pepperStr),
		rsaKey: rsaPriv,
	}, nil
}

func (s *aesCryptoService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *aesCryptoService) Decrypt(ciphertextB64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (s *aesCryptoService) HashPhone(phone string) string {
	hasher := sha256.New()
	hasher.Write([]byte(phone))
	hasher.Write(s.pepper)
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *aesCryptoService) GetRSAPublicKey() (*rsa.PublicKey, error) {
	if s.rsaKey == nil {
		return nil, errors.New("rsa private key not initialized")
	}
	return &s.rsaKey.PublicKey, nil
}

func (s *aesCryptoService) HashPassword(password string) (string, error) {
	return "", errors.New("not implemented for cascade aes")
}

func (s *aesCryptoService) ComparePassword(hash, password string) error {
	return errors.New("not implemented for cascade aes")
}
