package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"cascade/internal/application/port"
	"cascade/internal/delivery/http/dto"
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

func (u *AccountUseCase) ListAccounts(ctx context.Context) ([]*domain.Account, error) {
	return u.accountRepo.GetAll(ctx)
}

func (u *AccountUseCase) ListProxies(ctx context.Context) ([]*domain.Proxy, error) {
	return u.proxyRepo.GetAll(ctx)
}

func (u *AccountUseCase) AddProxy(ctx context.Context, req dto.CreateProxyRequest) (*domain.Proxy, error) {
	// Build the address string internally
	address := fmt.Sprintf("%s:%d", req.Host, req.Port)
	if req.Username != "" {
		address = fmt.Sprintf("%s:%d:%s:%s", req.Host, req.Port, req.Username, req.Password)
	}

	id, _ := uuid.NewV7()
	proxy := &domain.Proxy{
		ID:        id,
		Address:   address,
		Status:    domain.ProxyHealthy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.proxyRepo.Create(ctx, proxy); err != nil {
		return nil, err
	}
	return proxy, nil
}

func (u *AccountUseCase) RegisterAccount(ctx context.Context, phone string) (*domain.Account, error) {
	id, _ := uuid.NewV7()
	acc := &domain.Account{
		ID:              id,
		Phone:           phone,
		Status:          domain.StateWarmingUp,
		Channel:         "telegram",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 1. Save to database
	if err := u.accountRepo.CreateAccount(ctx, acc); err != nil {
		return nil, err
	}

	// 2. IMPORTANT: In production, trigger messenger (gotd) here.
	// For integration fix, we just return the created account.
	return acc, nil
}
