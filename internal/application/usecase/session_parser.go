package usecase

import (
	"context"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type SessionParser struct {
	storage port.SessionStorage
}

func NewSessionParser(storage port.SessionStorage) *SessionParser {
	return &SessionParser{
		storage: storage,
	}
}

func (p *SessionParser) Parse(ctx context.Context, data []byte) (*domain.TelethonSession, error) {
	return p.storage.ParseTelethonSession(ctx, data)
}
