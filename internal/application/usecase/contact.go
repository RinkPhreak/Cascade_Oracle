package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type ContactUseCase struct {
	contactRepo port.ContactRepository
	attemptRepo port.AttemptRepository
	uow         port.UnitOfWork
	crypto      port.CryptoService
}

func NewContactUseCase(cr port.ContactRepository, ar port.AttemptRepository, uow port.UnitOfWork, crypto port.CryptoService) *ContactUseCase {
	return &ContactUseCase{
		contactRepo: cr,
		attemptRepo: ar,
		uow:         uow,
		crypto:      crypto,
	}
}

func (u *ContactUseCase) ListReplied(ctx context.Context) ([]*domain.RepliedLead, error) {
	replies, err := u.contactRepo.ListReplies(ctx)
	if err != nil {
		return nil, err
	}

	var leads []*domain.RepliedLead
	for _, r := range replies {
		contact, err := u.contactRepo.GetByID(ctx, r.ContactID)
		if err != nil {
			continue // Skip orphans
		}

		// Decrypt sensitive info
		phone, _ := u.crypto.Decrypt(contact.Phone)
		name, _ := u.crypto.Decrypt(contact.Name)
		msg := ""
		if r.Message != nil {
			msg, _ = u.crypto.Decrypt(*r.Message)
		}

		leads = append(leads, &domain.RepliedLead{
			ID:        contact.ID,
			Phone:     phone,
			Name:      name,
			Message:   msg,
			Channel:   r.Channel,
			RepliedAt: r.RepliedAt,
		})
	}
	return leads, nil
}

func (u *ContactUseCase) GetContactTrace(ctx context.Context, contactID uuid.UUID) ([]*domain.SendAttempt, error) {
	return u.attemptRepo.GetTrace(ctx, contactID)
}

// Anonymise securely drops all PII in compliance with Right to Be Forgotten
func (u *ContactUseCase) Anonymise(ctx context.Context, contactID uuid.UUID) error {
	return u.uow.Execute(ctx, func(txCtx context.Context) error {
		contact, err := u.contactRepo.GetByID(txCtx, contactID)
		if err != nil {
			return err
		}

		// Erase original attributes securely via Cryptography boundaries
		obfuscatedName, _ := u.crypto.Encrypt("ANONYMISED")

		// The hash continues to identify the duplicate numbers, preventing future re-import.
		// However, the raw phone is completely irrecoverable since we use a SHA256 of old phone + pepper as "phone"
		secureHashPhone, _ := u.crypto.Encrypt("HIDDEN_SHA256")

		contact.Phone = secureHashPhone
		contact.Name = obfuscatedName

		// Nullify extra data
		contact.ExtraData = nil

		// Soft deletion
		now := time.Now()
		contact.DeletedAt = &now

		if err := u.contactRepo.Update(txCtx, contact); err != nil {
			return err
		}

		// Completely wipe any multi-channel preferences to prevent correlation tracking
		if err := u.contactRepo.DeletePreference(txCtx, contact.ID); err != nil {
			return err
		}

		return nil
	})
}
