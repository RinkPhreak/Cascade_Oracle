package port

import (
	"context"

	"cascade/internal/domain"
)

// SessionStorage defines the interface for parsing external session formats.
type SessionStorage interface {
	ParseTelethonSession(ctx context.Context, data []byte) (*domain.TelethonSession, error)
}
