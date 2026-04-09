package port

import (
	"context"

	"github.com/google/uuid"
)

// ImportedContact represents a result from the MTProto import routine.
type ImportedContact struct {
	Phone  string
	UserID int64
}

// TelegramClient abstracts the underlying gotd/td MTProto engine.
type TelegramClient interface {
	Send(ctx context.Context, accountID uuid.UUID, phone string, content string) (latencyMs int, err error)
	ImportContacts(ctx context.Context, accountID uuid.UUID, phones []string) ([]ImportedContact, error)
	DeleteContacts(ctx context.Context, accountID uuid.UUID, userIDs []int64) error
	Ping(ctx context.Context, accountID uuid.UUID) error
}

// SMSClient abstracts the sms.ru HTTP integration.
type SMSClient interface {
	Send(ctx context.Context, phone string, content string) (latencyMs int, err error)
}
