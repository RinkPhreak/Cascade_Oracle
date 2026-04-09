package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Generate bcrypt hash for "admin"
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	// Generate RSA keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}))

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}))

	// Escape newlines to make it a single line with literal \n as expected by the env loader block
	privStr := strings.ReplaceAll(privPEM, "\n", "\\n")
	pubStr := strings.ReplaceAll(pubPEM, "\n", "\\n")

	// Read .env file
	contentBytes, err := os.ReadFile(".env")
	if err != nil {
		panic(err)
	}
	content := string(contentBytes)

	// Replace the lines
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "ADMIN_PASSWORD_HASH=") {
			lines[i] = fmt.Sprintf("ADMIN_PASSWORD_HASH='%s'", string(hash))
		} else if strings.HasPrefix(line, "JWT_PRIVATE_KEY=") {
			lines[i] = fmt.Sprintf("JWT_PRIVATE_KEY=\"%s\"", privStr)
		} else if strings.HasPrefix(line, "JWT_PUBLIC_KEY=") {
			lines[i] = fmt.Sprintf("JWT_PUBLIC_KEY=\"%s\"", pubStr)
		}
	}

	err = os.WriteFile(".env", []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("SUCCESS")
}
