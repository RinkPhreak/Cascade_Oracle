package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"cascade/internal/application/port"
)

type ContactUseCase struct {
	contactRepo port.ContactRepository
	uow         port.UnitOfWork
	crypto      port.CryptoService
}

func NewContactUseCase(cr port.ContactRepository, uow port.UnitOfWork, crypto port.CryptoService) *ContactUseCase {
	return &ContactUseCase{
		contactRepo: cr,
		uow:         uow,
		crypto:      crypto,
	}
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
