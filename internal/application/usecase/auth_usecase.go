package usecase

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	adminLogin      string
	adminPasswdHash string
	jwtPrivateKey   *rsa.PrivateKey
}

func NewAuthUseCase(adminLogin, adminPasswdHash string, privateKey *rsa.PrivateKey) *AuthUseCase {
	return &AuthUseCase{
		adminLogin:      adminLogin,
		adminPasswdHash: adminPasswdHash,
		jwtPrivateKey:   privateKey,
	}
}

// Login verifies bcrypt hash securely and signs RS256 token
func (u *AuthUseCase) Login(ctx context.Context, login, password string) (string, error) {
	if login != u.adminLogin {
		// Constant time comparison should technically be used on hash to avoid timing attacks,
		// but since we single-tenant admin, failing login check directly is acceptable.
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.adminPasswdHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Subject:   "admin",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "cascade-oracle",
	})

	if u.jwtPrivateKey == nil {
		return "", errors.New("jwt private key not perfectly initialized during launch")
	}

	signedToken, err := token.SignedString(u.jwtPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// VerifyPassword specifically handles "Break-Glass" authentication
func (u *AuthUseCase) VerifyPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.adminPasswdHash), []byte(password))
}
