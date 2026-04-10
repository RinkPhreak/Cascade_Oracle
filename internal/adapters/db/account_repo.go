package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

// -- DB Models --

type proxyModel struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	Address   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (proxyModel) TableName() string { return "proxies" }

type accountModel struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid"`
	Phone           string
	Channel         string
	ProxyID         *uuid.UUID `gorm:"type:uuid"`
	State           string
	Credentials     string
	DailyCheckCount int
	DailySendCount  int
	CooldownUntil   *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (accountModel) TableName() string { return "accounts" }

type accountEventModel struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	AccountID uuid.UUID `gorm:"type:uuid"`
	EventType string
	Payload   string
	CreatedAt time.Time
}

func (accountEventModel) TableName() string { return "account_events" }

type sessionModel struct {
	AccountID   uuid.UUID `gorm:"primaryKey;column:account_id;type:uuid"`
	SessionData []byte
	UpdatedAt   time.Time
}

func (sessionModel) TableName() string { return "sessions" }

// -- Repository --

type gormAccountRepo struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) port.AccountRepository {
	return &gormAccountRepo{db: db}
}

// -- Mappings --

func toDomainAccount(am *accountModel) *domain.Account {
	if am == nil {
		return nil
	}
	var pID uuid.UUID
	if am.ProxyID != nil {
		pID = *am.ProxyID
	}
	return &domain.Account{
		ID:              am.ID,
		Phone:           am.Phone,
		Channel:         am.Channel,
		ProxyID:         pID,
		Status:          domain.AccountState(am.State),
		Credentials:     am.Credentials,
		DailyCheckCount: am.DailyCheckCount,
		DailySendCount:  am.DailySendCount,
		CooldownUntil:   am.CooldownUntil,
		CreatedAt:       am.CreatedAt,
		UpdatedAt:       am.UpdatedAt,
	}
}

func fromDomainAccount(a *domain.Account) *accountModel {
	am := &accountModel{
		ID:              a.ID,
		Phone:           a.Phone,
		Channel:         a.Channel,
		ProxyID:         nil,
		State:           string(a.Status),
		Credentials:     a.Credentials,
		DailyCheckCount: a.DailyCheckCount,
		DailySendCount:  a.DailySendCount,
		CooldownUntil:   a.CooldownUntil,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
	if a.ProxyID != uuid.Nil {
		am.ProxyID = &a.ProxyID
	}
	return am
}

// -- Methods --

func (r *gormAccountRepo) CreateAccount(ctx context.Context, account *domain.Account) error {
	m := fromDomainAccount(account)
	return ExtractDB(ctx, r.db).Create(m).Error
}

func (r *gormAccountRepo) GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	var m accountModel
	if err := ExtractDB(ctx, r.db).First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainAccount(&m), nil
}

func (r *gormAccountRepo) GetAccountByPhone(ctx context.Context, phone string) (*domain.Account, error) {
	var m accountModel
	if err := ExtractDB(ctx, r.db).Where("phone = ?", phone).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainAccount(&m), nil
}

func (r *gormAccountRepo) UpdateAccount(ctx context.Context, account *domain.Account) error {
	m := fromDomainAccount(account)
	return ExtractDB(ctx, r.db).Save(m).Error
}

func (r *gormAccountRepo) GetLeastBusyActiveAccount(ctx context.Context, channel string) (*domain.Account, error) {
	var m accountModel

	err := ExtractDB(ctx, r.db).
		Where("channel = ?", channel).
		Where("created_at < ?", time.Now().Add(-48*time.Hour)).
		Where("state = ? OR state = ?", domain.StateActive, domain.StateWarmingUp).
		Where("cooldown_until IS NULL OR cooldown_until <= ?", time.Now()).
		Order("daily_send_count asc, daily_check_count asc").
		First(&m).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainAccount(&m), nil
}

func (r *gormAccountRepo) CountActiveAccounts(ctx context.Context) (int, error) {
	var count int64
	err := ExtractDB(ctx, r.db).Model(&accountModel{}).
		Where("state = ?", domain.StateActive).Count(&count).Error
	return int(count), err
}

func (r *gormAccountRepo) ResetDailyCounters(ctx context.Context) error {
	return ExtractDB(ctx, r.db).Model(&accountModel{}).
		Updates(map[string]interface{}{"daily_check_count": 0, "daily_send_count": 0}).Error
}

func (r *gormAccountRepo) CreateAccountEvent(ctx context.Context, event *domain.AccountEvent) error {
	em := &accountEventModel{
		ID:        event.ID,
		AccountID: event.AccountID,
		EventType: event.EventType,
		Payload:   event.Payload,
		CreatedAt: event.CreatedAt,
	}
	return ExtractDB(ctx, r.db).Create(em).Error
}

func (r *gormAccountRepo) SaveSession(ctx context.Context, session *domain.Session) error {
	sm := &sessionModel{
		AccountID:   session.AccountID,
		SessionData: []byte(session.SessionData),
		UpdatedAt:   session.UpdatedAt,
	}
	return ExtractDB(ctx, r.db).Save(sm).Error
}

func (r *gormAccountRepo) GetSession(ctx context.Context, accountID uuid.UUID) (*domain.Session, error) {
	var m sessionModel
	if err := ExtractDB(ctx, r.db).First(&m, "account_id = ?", accountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.Session{
		ID:          uuid.Nil,
		AccountID:   m.AccountID,
		SessionData: string(m.SessionData),
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

func (r *gormAccountRepo) GetAll(ctx context.Context) ([]*domain.Account, error) {
	var models []accountModel
	if err := ExtractDB(ctx, r.db).Find(&models).Error; err != nil {
		return nil, err
	}
	var res []*domain.Account
	for _, m := range models {
		mCopy := m
		res = append(res, toDomainAccount(&mCopy))
	}
	return res, nil
}
