package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type IncomingMessageHandler struct {
	contactRepo port.ContactRepository
	crypto      port.CryptoService
}

func NewIncomingMessageHandler(cr port.ContactRepository, crypto port.CryptoService) *IncomingMessageHandler {
	return &IncomingMessageHandler{
		contactRepo: cr,
		crypto:      crypto,
	}
}

func (h *IncomingMessageHandler) HandleReply(ctx context.Context, accountID uuid.UUID, senderPhone string, messageText string, channel string) error {
	hash := h.crypto.HashPhone(senderPhone)
	
	contact, err := h.contactRepo.GetByHash(ctx, hash)
	if err != nil {
		if err == domain.ErrNotFound {
			// Unregistered sender inbound interaction logic bounds 
			return nil
		}
		return err
	}

	now := time.Now()

	encryptedMsg, err := h.crypto.Encrypt(messageText)
	if err != nil {
		return err
	}

	replyID, _ := uuid.NewV7()
	reply := &domain.ContactReply{
		ID:        replyID,
		ContactID: contact.ID,
		AccountID: accountID,
		Channel:   channel,
		Message:   &encryptedMsg,
		RepliedAt: now,
		CreatedAt: now,
	}

	if err := h.contactRepo.SaveReply(ctx, reply); err != nil {
		return err
	}

	contact.MarkReplied(now)
	return h.contactRepo.Update(ctx, contact)
}
