package domain

import (
	"time"

	"github.com/google/uuid"
)

type OperatorSession struct {
	ID           uuid.UUID
	RefreshToken string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
}

func (s *OperatorSession) IsValid(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	if now.After(s.ExpiresAt) {
		return false
	}
	return true
}

func (s *OperatorSession) Revoke(now time.Time) {
	s.RevokedAt = &now
}
