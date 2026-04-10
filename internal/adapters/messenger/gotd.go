package messenger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"golang.org/x/net/proxy"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type gotdClientPool struct {
	accountRepo port.AccountRepository
	appID       int
	appHash     string
	proxyRepo   port.ProxyRepository

	mu      sync.RWMutex
	clients map[uuid.UUID]*telegram.Client
	cancels map[uuid.UUID]context.CancelFunc
}

func NewTelegramClientPool(repo port.AccountRepository, proxyRepo port.ProxyRepository, appID int, appHash string) port.TelegramClient {
	return &gotdClientPool{
		accountRepo: repo,
		proxyRepo:   proxyRepo,
		appID:       appID,
		appHash:     appHash,
		clients:     make(map[uuid.UUID]*telegram.Client),
		cancels:     make(map[uuid.UUID]context.CancelFunc),
	}
}

func (p *gotdClientPool) Init(ctx context.Context, accountID uuid.UUID) error {
	_, err := p.getClient(ctx, accountID)
	return err
}

func (p *gotdClientPool) getClient(ctx context.Context, accountID uuid.UUID) (*telegram.Client, error) {
	p.mu.RLock()
	client, exists := p.clients[accountID]
	p.mu.RUnlock()

	if exists {
		return client, nil
	}

	acc, err := p.accountRepo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	_ = acc

	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists = p.clients[accountID]; exists {
		return client, nil
	}

	sessionStorage := &gotdDBSessionStorage{
		accountID: accountID,
		repo:      p.accountRepo,
	}

	opts := telegram.Options{
		SessionStorage: sessionStorage,
	}

	// 1. Setup Proxy if available
	if acc.ProxyID != uuid.Nil {
		prx, err := p.proxyRepo.GetByID(ctx, acc.ProxyID)
		if err == nil {
			dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%s:%d", prx.Host, prx.Port), &proxy.Auth{
				User:     prx.Username,
				Password: prx.Password,
			}, proxy.Direct)
			if err == nil {
				contextDialer := dialer.(proxy.ContextDialer)
				opts.Resolver = dcs.Plain(dcs.PlainOptions{
					Dial: contextDialer.DialContext,
				})
			}
		}
	}

	// 2. Setup Device Config from account credentials if JSON
	if acc.Credentials != "" && strings.HasPrefix(acc.Credentials, "{") {
		var meta struct {
			AppVersion    string `json:"app_version"`
			DeviceModel   string `json:"device_model"`
			SystemVersion string `json:"system_version"`
		}
		if err := json.Unmarshal([]byte(acc.Credentials), &meta); err == nil {
			opts.Device = telegram.DeviceConfig{
				DeviceModel:   meta.DeviceModel,
				SystemVersion: meta.SystemVersion,
				AppVersion:    meta.AppVersion,
			}
		}
	}

	client = telegram.NewClient(p.appID, p.appHash, opts)

	runCtx, cancel := context.WithCancel(context.Background())

	// Tier 1.1 Fix: Supervisor Goroutine loop for mtproto long-lived connection.
	go func() {
		defer func() {
			// Prevent random MTProto panics from crashing the cascade binary
			recover()
		}()
		// RunUntilCanceled restores dropped connections natively and gracefully dies on ctx cancel.
		_ = telegram.RunUntilCanceled(runCtx, client)
	}()

	p.clients[accountID] = client
	p.cancels[accountID] = cancel

	return client, nil
}

// StopClient enables graceful disconnect of a telegram node and clears it from pool
func (p *gotdClientPool) StopClient(accountID uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if cancel, exists := p.cancels[accountID]; exists {
		cancel()
		delete(p.cancels, accountID)
		delete(p.clients, accountID)
	}
}

func (p *gotdClientPool) isMockMode() bool {
	return os.Getenv("MESSENGER_MODE") == "mock"
}

func (p *gotdClientPool) ImportContacts(ctx context.Context, accountID uuid.UUID, phones []string) ([]port.ImportedContact, error) {
	if p.isMockMode() {
		res := make([]port.ImportedContact, 0)
		for _, ph := range phones {
			if ph != "" && ph != "404" {
				res = append(res, port.ImportedContact{
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

	api := tg.NewClient(client)
	var inputContacts []tg.InputPhoneContact
	for i, phone := range phones {
		inputContacts = append(inputContacts, tg.InputPhoneContact{
			ClientID:  int64(i + 1),
			Phone:     phone,
			FirstName: "CascadeLead_" + uuid.New().String()[:8],
		})
	}

	// Use the native imported method without Run wrapper
	imported, err := api.ContactsImportContacts(ctx, inputContacts)
	if err != nil {
		// Tier 1.2: Strict Typed errors via tgerr
		if tgerr.Is(err, "PEER_FLOOD") {
			return nil, domain.ErrPeerFlood
		}
		if tgerr.Is(err, "AUTH_KEY_UNREGISTERED") {
			p.StopClient(accountID)
			return nil, domain.ErrAuthUnregistered
		}
		return nil, err
	}

	var results []port.ImportedContact
	for _, user := range imported.Users {
		u, ok := user.(*tg.User)
		if ok && !u.Deleted {
			results = append(results, port.ImportedContact{
				UserID:   u.ID,
				Phone:    u.Phone,
			})
		}
	}
	return results, nil
}

func (p *gotdClientPool) Ping(ctx context.Context, accountID uuid.UUID) error {
	_, err := p.getClient(ctx, accountID)
	return err
}

func (p *gotdClientPool) VerifySession(ctx context.Context, accountID uuid.UUID) (string, error) {
	if p.isMockMode() {
		return "+79991234567", nil
	}

	client, err := p.getClient(ctx, accountID)
	if err != nil {
		return "", err
	}

	api := tg.NewClient(client)
	self, err := api.UsersGetFullUser(ctx, &tg.InputUserSelf{})
	if err != nil {
		return "", fmt.Errorf("failed to get self info: %w", err)
	}

	// self is already *tg.UsersUserFull, no need for type assertion.
	for _, u := range self.Users {
		user, ok := u.(*tg.User)
		if ok && user.Self {
			return user.Phone, nil
		}
	}

	return "", fmt.Errorf("self user not found in response")
}

func (p *gotdClientPool) DeleteContacts(ctx context.Context, accountID uuid.UUID, userIDs []int64) error {
	if p.isMockMode() {
		return nil
	}

	client, err := p.getClient(ctx, accountID)
	if err != nil {
		return err
	}

	api := tg.NewClient(client)
	var inputs []tg.InputUserClass
	for _, uID := range userIDs {
		inputs = append(inputs, &tg.InputUser{UserID: uID})
	}
	_, err = api.ContactsDeleteContacts(ctx, inputs)
	return err
}

func (p *gotdClientPool) Send(ctx context.Context, accountID uuid.UUID, phone string, text string) (int, error) {
	start := time.Now()
	if p.isMockMode() {
		if phone == "404" {
			return 0, domain.ErrPeerFlood
		}
		return 50, nil
	}

	client, err := p.getClient(ctx, accountID)
	if err != nil {
		return 0, err
	}

	api := tg.NewClient(client)
	sender := message.NewSender(api)

	resolv := peer.DefaultResolver(api)
	peerEntity, err := resolv.ResolvePhone(ctx, phone)
	if err != nil {
		return 0, domain.ErrUserNotFound
	}

	_, err = sender.To(peerEntity).Text(ctx, text)

	if err != nil {
		if tgerr.Is(err, "PEER_FLOOD") {
			return 0, domain.ErrPeerFlood
		}
		if tgerr.Is(err, "AUTH_KEY_UNREGISTERED") {
			p.StopClient(accountID)
			return 0, domain.ErrAuthUnregistered
		}
		return 0, err
	}

	return int(time.Since(start).Milliseconds()), nil
}

type gotdDBSessionStorage struct {
	accountID uuid.UUID
	repo      port.AccountRepository
}

func (s *gotdDBSessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	session, err := s.repo.GetSession(ctx, s.accountID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil // unblocks clear gotd instantiation defaults
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
