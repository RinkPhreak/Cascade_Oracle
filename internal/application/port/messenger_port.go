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
	Init(ctx context.Context, accountID uuid.UUID) error
	Send(ctx context.Context, accountID uuid.UUID, phone string, content string) (latencyMs int, err error)
	ImportContacts(ctx context.Context, accountID uuid.UUID, phones []string) ([]ImportedContact, error)
	DeleteContacts(ctx context.Context, accountID uuid.UUID, userIDs []int64) error
	Ping(ctx context.Context, accountID uuid.UUID) error
	VerifySession(ctx context.Context, accountID uuid.UUID) (string, error)
	StopClient(accountID uuid.UUID)
	// VerifySessionWithCredentials verifies a session using custom API credentials and DC routing.
	// This is used during account import when the session was created with different API_ID than system defaults.
	VerifySessionWithCredentials(ctx context.Context, accountID uuid.UUID, appID int, appHash string, dcID int, dcAddr string) (string, error)
}

// SMSClient abstracts the sms.ru HTTP integration.
type SMSClient interface {
	Send(ctx context.Context, phone string, content string) (latencyMs int, err error)
}
