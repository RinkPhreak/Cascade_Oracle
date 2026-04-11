package usecase

import (
	"context"
	"time"

	"cascade/internal/application/port"
	"cascade/internal/domain"

	"github.com/google/uuid"
)

type OperatorAuthUseCase struct {
	repo   port.OperatorRepository
	crypto port.CryptoService
}

func NewOperatorAuthUseCase(repo port.OperatorRepository, crypto port.CryptoService) *OperatorAuthUseCase {
	return &OperatorAuthUseCase{repo: repo, crypto: crypto}
}

func (u *OperatorAuthUseCase) CreateSession(ctx context.Context, login string, password string, expectedHash string, expectedLogin string) (*domain.OperatorSession, error) {
	if login != expectedLogin {
		return nil, domain.ErrNotFound
	}
	if err := u.crypto.ComparePassword(expectedHash, password); err != nil {
		return nil, domain.ErrNotFound
	}

	sessionID, _ := uuid.NewV7()
	tokenUUID, _ := uuid.NewV7() // High-entropy verifiable refresh token

	session := &domain.OperatorSession{
		ID:           sessionID,
		RefreshToken: tokenUUID.String(),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := u.repo.SaveSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (u *OperatorAuthUseCase) ValidateSession(ctx context.Context, token string) (*domain.OperatorSession, error) {
	session, err := u.repo.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if !session.IsValid(time.Now()) {
		return nil, domain.ErrNotFound
	}
	return session, nil
}

func (u *OperatorAuthUseCase) RevokeSession(ctx context.Context, token string) error {
	session, err := u.repo.GetSessionByToken(ctx, token)
	if err != nil {
		return err
	}
	session.Revoke(time.Now())
	return u.repo.UpdateSession(ctx, session)
}
