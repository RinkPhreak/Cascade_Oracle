package domain

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID         uuid.UUID
	PhoneHash  string
	Phone      string  // AES-256-GCM Encrypted
	Name       string  // AES-256-GCM Encrypted
	ExtraData  *string // Encrypted JSON
	HasReplied bool
	RepliedAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time // Right to Erasure / Anonymised
}

func (c *Contact) MarkReplied(repliedAt time.Time) {
	c.HasReplied = true
	c.RepliedAt = &repliedAt
	c.UpdatedAt = time.Now()
}

func (c *Contact) IsAnonymised() bool {
	return c.DeletedAt != nil
}

type ContactChannelPreference struct {
	ContactID        uuid.UUID
	PreferredChannel string
	UpdatedAt        time.Time
}

type ContactReply struct {
	ID        uuid.UUID
	ContactID uuid.UUID
	AccountID uuid.UUID
	Channel   string  // telegram, sms
	Message   *string // Encrypted
	RepliedAt time.Time
	CreatedAt time.Time
}

type RepliedLead struct {
	ID        uuid.UUID
	Phone     string
	Name      string
	Message   string
	Channel   string
	RepliedAt time.Time
}
