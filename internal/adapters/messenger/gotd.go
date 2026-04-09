package messenger

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type gotdClientPool struct {
	accountRepo port.AccountRepository
	appID       int
	appHash     string

	mu      sync.RWMutex
	clients map[uuid.UUID]*telegram.Client
}

func NewTelegramClientPool(repo port.AccountRepository, appID int, appHash string) port.TelegramClient {
	return &gotdClientPool{
		accountRepo: repo,
		appID:       appID,
		appHash:     appHash,
		clients:     make(map[uuid.UUID]*telegram.Client),
	}
}

// getClient dynamically loads and configures an MTProto client using robust pool caching 
func (p *gotdClientPool) getClient(ctx context.Context, accountID uuid.UUID) (*telegram.Client, error) {
	p.mu.RLock()
	client, exists := p.clients[accountID]
	p.mu.RUnlock()

	if exists {
		return client, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double check
	if client, exists = p.clients[accountID]; exists {
		return client, nil
	}

	acc, err := p.accountRepo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	_ = acc // Note: Production proxy configuration logic would consume acc.ProxyID here

	sessionStorage := &gotdDBSessionStorage{
		accountID: accountID,
		repo:      p.accountRepo,
	}

	client = telegram.NewClient(p.appID, p.appHash, telegram.Options{
		SessionStorage: sessionStorage,
	})

	p.clients[accountID] = client
	return client, nil
}

func (p *gotdClientPool) isMockMode() bool {
	return os.Getenv("MESSENGER_MODE") == "mock"
}

func (p *gotdClientPool) ImportContacts(ctx context.Context, accountID uuid.UUID, phones []string) ([]domain.TelegramUser, error) {
	if p.isMockMode() {
		res := make([]domain.TelegramUser, 0)
		for _, ph := range phones {
			if ph != "" && ph != "404" {
				res = append(res, domain.TelegramUser{
					UserID: int64(len(ph) * 1000), 
					Phone:  ph,
				})
			}
		}
		return res, nil
	}

	client, err := p.getClient(ctx, accountID)
	if err != nil {
		return nil, err
	}

	var results []domain.TelegramUser
	err = client.Run(ctx, func(tctx context.Context) error {
		api := tg.NewClient(client)

		var inputContacts []tg.InputPhoneContact
		for i, phone := range phones {
			inputContacts = append(inputContacts, tg.InputPhoneContact{
				ClientID:  int64(i + 1),
				Phone:     phone,
				FirstName: "CascadeLead_" + uuid.New().String()[:8], // Minimum viable 152-FZ safe name
			})
		}

		imported, err := api.ContactsImportContacts(tctx, inputContacts)
		if err != nil {
			return err
		}

		for _, user := range imported.Users {
			u, ok := user.(*tg.User)
			if ok && !u.Deleted {
				results = append(results, domain.TelegramUser{
					UserID:   u.ID,
					Username: u.Username,
					Phone:    u.Phone,
				})
			}
		}
		return nil
	})

	return results, err
}

func (p *gotdClientPool) DeleteContacts(ctx context.Context, accountID uuid.UUID, userIDs []int64) error {
	if p.isMockMode() {
		return nil
	}

	client, err := p.getClient(ctx, accountID)
	if err != nil {
		return err
	}

	return client.Run(ctx, func(tctx context.Context) error {
		api := tg.NewClient(client)
		var inputs []tg.InputUserClass
		for _, uID := range userIDs {
			inputs = append(inputs, &tg.InputUser{UserID: uID})
		}
		_, err := api.ContactsDeleteContacts(tctx, inputs)
		return err
	})
}

func (p *gotdClientPool) Send(ctx context.Context, accountID uuid.UUID, phone string, text string) (int, error) {
	start := time.Now()
	if p.isMockMode() {
		if phone == "404" {
			return 0, errors.New("PEER_FLOOD")
		}
		return 50, nil
	}

	client, err := p.getClient(ctx, accountID)
	if err != nil {
		return 0, err
	}

	err = client.Run(ctx, func(tctx context.Context) error {
		api := tg.NewClient(client)
		sender := message.NewSender(api)
		
		resolv := peer.DefaultResolver(api)
		peerEntity, err := resolv.ResolvePhone(tctx, phone)
		if err != nil {
			return domain.ErrUserNotFound
		}

		_, err = sender.To(peerEntity).Text(tctx, text)
		
		// Map MTProto exceptions to standard rules
		if err != nil {
			errStr := err.Error()
			if errStr == "PEER_FLOOD" || errStr == "AUTH_KEY_UNREGISTERED" {
				return errors.New(errStr) // Handled dynamically by usecase
			}
		}

		return err
	})

	return int(time.Since(start).Milliseconds()), err
}

// gotdDBSessionStorage connects gotd's session blob semantics to our PostgreSQL adapter interface.
type gotdDBSessionStorage struct {
	accountID uuid.UUID
	repo      port.AccountRepository
}

func (s *gotdDBSessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	session, err := s.repo.GetSession(ctx, s.accountID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil // gotd standard specifies nil byte array unblocks clean init
		}
		return nil, err
	}
	return []byte(session.SessionData), nil
}

func (s *gotdDBSessionStorage) StoreSession(ctx context.Context, data []byte) error {
	return s.repo.SaveSession(ctx, &domain.Session{
		AccountID:   s.accountID,
		SessionData: string(data),
		UpdatedAt:   time.Now(),
	})
}
