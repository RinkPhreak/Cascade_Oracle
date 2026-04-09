package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type AccountUseCase struct {
	accountRepo port.AccountRepository
	proxyRepo   port.ProxyRepository
}

func NewAccountUseCase(ar port.AccountRepository, pr port.ProxyRepository) *AccountUseCase {
	return &AccountUseCase{
		accountRepo: ar,
		proxyRepo:   pr,
	}
}

func (u *AccountUseCase) TransitionState(ctx context.Context, accountID uuid.UUID, state domain.AccountState, reason string) error {
	acc, err := u.accountRepo.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}

	acc.TransitionState(state)
	if err := u.accountRepo.UpdateAccount(ctx, acc); err != nil {
		return err
	}

	evtID, _ := uuid.NewV7()
	event := &domain.AccountEvent{
		ID:        evtID,
		AccountID: acc.ID,
		EventType: "state_transition",
		Payload:   string(state) + ": " + reason,
		CreatedAt: time.Now(),
	}
	return u.accountRepo.CreateAccountEvent(ctx, event)
}

func (u *AccountUseCase) HandleServiceNotification(ctx context.Context, accountID uuid.UUID, message string) error {
	evtID, _ := uuid.NewV7()
	event := &domain.AccountEvent{
		ID:        evtID,
		AccountID: accountID,
		EventType: "service_notice",
		Payload:   message,
		CreatedAt: time.Now(),
	}
	return u.accountRepo.CreateAccountEvent(ctx, event)
}

func (u *AccountUseCase) MarkProxyDegraded(ctx context.Context, proxyID uuid.UUID) error {
	proxy, err := u.proxyRepo.GetByID(ctx, proxyID)
	if err != nil {
		return err
	}

	proxy.MarkDegraded()
	return u.proxyRepo.Update(ctx, proxy)
}
