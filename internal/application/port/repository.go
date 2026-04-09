package port

import (
	"context"
	"time"

	"cascade/internal/domain"

	"github.com/google/uuid"
)

// AccountRepository defines operations for managing Telegram accounts.
type AccountRepository interface {
	CreateAccount(ctx context.Context, account *domain.Account) error
	GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetAccountByPhone(ctx context.Context, phone string) (*domain.Account, error)
	UpdateAccount(ctx context.Context, account *domain.Account) error
	GetLeastBusyActiveAccount(ctx context.Context, channel string) (*domain.Account, error)
	CountActiveAccounts(ctx context.Context) (int, error)
	ResetDailyCounters(ctx context.Context) error

	CreateAccountEvent(ctx context.Context, event *domain.AccountEvent) error

	SaveSession(ctx context.Context, session *domain.Session) error
	GetSession(ctx context.Context, accountID uuid.UUID) (*domain.Session, error)
}

// ProxyRepository isolates operations for network proxies.
type ProxyRepository interface {
	Create(ctx context.Context, proxy *domain.Proxy) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Proxy, error)
	Update(ctx context.Context, proxy *domain.Proxy) error
	GetAll(ctx context.Context) ([]*domain.Proxy, error)
}

// ContactRepository manages the B2B leads directory.
type ContactRepository interface {
	Create(ctx context.Context, contact *domain.Contact) error
	CreateBatch(ctx context.Context, contacts []*domain.Contact) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error)
	GetByHash(ctx context.Context, phoneHash string) (*domain.Contact, error)
	Update(ctx context.Context, contact *domain.Contact) error

	SaveReply(ctx context.Context, reply *domain.ContactReply) error
	SavePreference(ctx context.Context, pref *domain.ContactChannelPreference) error
	GetPreference(ctx context.Context, contactID uuid.UUID) (string, error)

	FetchAnonymizeCandidates(ctx context.Context, retentionThreshold time.Time) ([]*domain.Contact, error)
}

// CampaignRepository manages message flows and scheduled templates.
type CampaignRepository interface {
	Create(ctx context.Context, campaign *domain.Campaign) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Campaign, error)
	Update(ctx context.Context, campaign *domain.Campaign) error
	FetchExecutable(ctx context.Context, timeZero time.Time) ([]*domain.Campaign, error)
	ListCampaigns(ctx context.Context) ([]*domain.Campaign, error)
	ListCampaignContacts(ctx context.Context, campaignID uuid.UUID) ([]*domain.CampaignContact, error)

	GetTemplate(ctx context.Context, campaignID uuid.UUID, channel string) (*domain.MessageTemplate, error)
	CreateTemplate(ctx context.Context, tpl *domain.MessageTemplate) error
	UpdateCampaignContactStatus(ctx context.Context, campaignID, contactID uuid.UUID, status domain.CampaignContactStatus) error
	AssignContactsToCampaign(ctx context.Context, campaignID uuid.UUID, contactIDs []uuid.UUID, status domain.CampaignContactStatus) error
}

// AttemptRepository enforces idempotency boundaries for queue delivery.
type AttemptRepository interface {
	Upsert(ctx context.Context, attempt *domain.SendAttempt) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SendAttempt, error)
	GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (*domain.SendAttempt, error)
	GetStuck(ctx context.Context, limitTime time.Time) ([]*domain.SendAttempt, error)
	Update(ctx context.Context, attempt *domain.SendAttempt) error
}

// OperatorRepository manages access token revocation.
type OperatorRepository interface {
	SaveSession(ctx context.Context, session *domain.OperatorSession) error
	GetSessionByToken(ctx context.Context, token string) (*domain.OperatorSession, error)
	UpdateSession(ctx context.Context, session *domain.OperatorSession) error
}

// UnitOfWork defines bounds for SQL transactions spanning multiple repositories.
type UnitOfWork interface {
	Execute(ctx context.Context, fn func(context.Context) error) error
}
