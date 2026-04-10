package usecase

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gotd/td/session"

	"cascade/internal/application/port"
	"cascade/internal/delivery/http/dto"
	"cascade/internal/domain"
)

type AccountUseCase struct {
	accountRepo port.AccountRepository
	proxyRepo   port.ProxyRepository
	tgClient    port.TelegramClient
}

func NewAccountUseCase(ar port.AccountRepository, pr port.ProxyRepository, tgc port.TelegramClient) *AccountUseCase {
	return &AccountUseCase{
		accountRepo: ar,
		proxyRepo:   pr,
		tgClient:    tgc,
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
	id, _ := uuid.NewV7()
	proxy := &domain.Proxy{
		ID:        id,
		Host:      req.Host,
		Port:      req.Port,
		Username:  req.Username,
		Password:  req.Password,
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
		ID:        id,
		Phone:     phone,
		Status:    domain.StateWarmingUp,
		Channel:   "telegram",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1. Save to database
	if err := u.accountRepo.CreateAccount(ctx, acc); err != nil {
		return nil, err
	}

	// 2. Trigger the actual Telegram connection (MTProto Handshake/Session Init)
	if err := u.tgClient.Init(ctx, acc.ID); err != nil {
		return nil, fmt.Errorf("failed to init telegram session: %w", err)
	}

	return acc, nil
}

func (u *AccountUseCase) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	// 1. Stop the telegram client in the pool
	u.tgClient.StopClient(id)
	
	// 2. Delete from DB
	return u.accountRepo.DeleteAccount(ctx, id)
}

func (u *AccountUseCase) ImportAccount(ctx context.Context, files map[string][]byte, proxyReq dto.CreateProxyRequest, comment string) (*domain.Account, error) {
	// 1. Validate Proxy (Strictly Mandatory)
	if proxyReq.Host == "" || proxyReq.Port == 0 {
		return nil, fmt.Errorf("proxy is mandatory for account import")
	}

	var sessionData []byte
	var jsonMeta []byte
	var zipData []byte

	// 2. Extract files
	for name, data := range files {
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".zip" {
			zipData = data
			break
		}
		if ext == ".session" {
			sessionData = data
		}
		if ext == ".json" {
			jsonMeta = data
		}
	}

	if zipData != nil {
		r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
		if err != nil {
			return nil, fmt.Errorf("failed to read zip: %w", err)
		}
		for _, f := range r.File {
			ext := strings.ToLower(filepath.Ext(f.Name))
			if ext == ".session" {
				rc, _ := f.Open()
				sessionData, _ = io.ReadAll(rc)
				rc.Close()
			}
			if ext == ".json" {
				rc, _ := f.Open()
				jsonMeta, _ = io.ReadAll(rc)
				rc.Close()
			}
		}
	}

	if sessionData == nil {
		return nil, fmt.Errorf("missing .session file")
	}

	// 3. Parse Telethon Session
	tmpPath, err := createTempSessionFile(sessionData)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpPath)

	teleSession, err := parseTelethonSession(ctx, tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse telethon session: %w", err)
	}

	// 4. Prepare Proxy
	proxy, err := u.AddProxy(ctx, proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to save proxy: %w", err)
	}

	// 5. Initialize Device Metadata
	var meta struct {
		AppVersion    string `json:"app_version"`
		DeviceModel   string `json:"device_model"`
		SystemVersion string `json:"system_version"`
		Comment       string `json:"comment"`
	}
	meta.Comment = comment
	if jsonMeta != nil {
		_ = json.Unmarshal(jsonMeta, &meta)
	}
	credBytes, _ := json.Marshal(meta)

	// 6. Initialize gotd Session Data
	gotdSessionData := &session.Data{
		DC:      teleSession.DCID,
		Addr:    teleSession.ServerAddress,
		AuthKey: teleSession.AuthKey,
	}
	
	marshaled, err := json.Marshal(gotdSessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gotd session: %w", err)
	}

	// 7. Create Account Record (UPSERT)
	accID, _ := uuid.NewV7()
	acc := &domain.Account{
		ID:          accID,
		Phone:       "PENDING_" + accID.String()[:8], // Temporary unique phone
		Status:      domain.StateWarmingUp,
		Channel:     "telegram",
		ProxyID:     proxy.ID,
		Credentials: string(credBytes),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save session first so verification can use it
	if err := u.accountRepo.SaveSession(ctx, &domain.Session{
		AccountID:   acc.ID,
		SessionData: string(marshaled),
		UpdatedAt:   time.Now(),
	}); err != nil {
		return nil, err
	}

	// 8. Verify with gotd (Strict Timeout)
	verifyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	phone, err := u.tgClient.VerifySession(verifyCtx, acc.ID)
	if err != nil {
		return nil, fmt.Errorf("session verification failed (proxy or auth error): %w", err)
	}

	acc.Phone = phone
	acc.Status = domain.StateActive

	if err := u.accountRepo.CreateAccount(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}
