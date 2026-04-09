package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

// -- DB Models --

type contactModel struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	PhoneHash  string
	Phone      string
	Name       string
	ExtraData  *string
	HasReplied bool
	RepliedAt  *time.Time
	DeletedAt  *time.Time
	CreatedAt  time.Time
}

func (contactModel) TableName() string { return "contacts" }

type contactReplyModel struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	ContactID  uuid.UUID `gorm:"type:uuid"`
	AccountID  uuid.UUID `gorm:"type:uuid"`
	Channel    string
	Message    *string
	RepliedAt  time.Time
	CreatedAt  time.Time
}

func (contactReplyModel) TableName() string { return "contact_replies" }

type contactChannelPreferenceModel struct {
	ContactID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	PreferredChannel string
	UpdatedAt        time.Time
}

func (contactChannelPreferenceModel) TableName() string { return "contact_channel_preferences" }

// -- Repository --

type gormContactRepo struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) port.ContactRepository {
	return &gormContactRepo{db: db}
}

// -- Mappings --

func fromDomainContact(c *domain.Contact) *contactModel {
	return &contactModel{
		ID:         c.ID,
		PhoneHash:  c.PhoneHash,
		Phone:      c.Phone,
		Name:       c.Name,
		ExtraData:  c.ExtraData,
		HasReplied: c.HasReplied,
		RepliedAt:  c.RepliedAt,
		DeletedAt:  c.DeletedAt,
		CreatedAt:  c.CreatedAt,
	}
}

func toDomainContact(m *contactModel) *domain.Contact {
	return &domain.Contact{
		ID:         m.ID,
		PhoneHash:  m.PhoneHash,
		Phone:      m.Phone,
		Name:       m.Name,
		ExtraData:  m.ExtraData,
		HasReplied: m.HasReplied,
		RepliedAt:  m.RepliedAt,
		DeletedAt:  m.DeletedAt,
		CreatedAt:  m.CreatedAt,
	}
}

// -- Methods --

func (r *gormContactRepo) Create(ctx context.Context, contact *domain.Contact) error {
	m := fromDomainContact(contact)
	return ExtractDB(ctx, r.db).Create(m).Error
}

func (r *gormContactRepo) CreateBatch(ctx context.Context, contacts []*domain.Contact) error {
	if len(contacts) == 0 {
		return nil
	}
	var models []contactModel
	for _, c := range contacts {
		models = append(models, *fromDomainContact(c))
	}
	// GORM's CreateInBatches avoids extreme parameter length errors.
	// ON CONFLICT DO NOTHING natively skips existing contacts on idempotency/hash collision.
	return ExtractDB(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(models, len(models)).Error
}

func (r *gormContactRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	var m contactModel
	if err := ExtractDB(ctx, r.db).First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainContact(&m), nil
}

func (r *gormContactRepo) GetByHash(ctx context.Context, phoneHash string) (*domain.Contact, error) {
	var m contactModel
	if err := ExtractDB(ctx, r.db).Where("phone_hash = ?", phoneHash).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainContact(&m), nil
}

func (r *gormContactRepo) Update(ctx context.Context, contact *domain.Contact) error {
	m := fromDomainContact(contact)
	return ExtractDB(ctx, r.db).Save(m).Error
}

func (r *gormContactRepo) SaveReply(ctx context.Context, reply *domain.ContactReply) error {
	rm := &contactReplyModel{
		ID:        reply.ID,
		ContactID: reply.ContactID,
		AccountID: reply.AccountID,
		Channel:   reply.Channel,
		Message:   reply.Message,
		RepliedAt: reply.RepliedAt,
		CreatedAt: reply.CreatedAt,
	}
	return ExtractDB(ctx, r.db).Create(rm).Error
}

func (r *gormContactRepo) SavePreference(ctx context.Context, pref *domain.ContactChannelPreference) error {
	pm := &contactChannelPreferenceModel{
		ContactID:        pref.ContactID,
		PreferredChannel: pref.PreferredChannel,
		UpdatedAt:        pref.UpdatedAt,
	}
	return ExtractDB(ctx, r.db).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(pm).Error
}

func (r *gormContactRepo) GetPreference(ctx context.Context, contactID uuid.UUID) (string, error) {
	var pm contactChannelPreferenceModel
	if err := ExtractDB(ctx, r.db).First(&pm, "contact_id = ?", contactID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil 
		}
		return "", err
	}
	return pm.PreferredChannel, nil
}

func (r *gormContactRepo) FetchAnonymizeCandidates(ctx context.Context, retentionThreshold time.Time) ([]*domain.Contact, error) {
	var models []contactModel
	err := ExtractDB(ctx, r.db).Where("deleted_at IS NULL AND created_at < ?", retentionThreshold).Find(&models).Error
	if err != nil {
		return nil, err
	}
	var res []*domain.Contact
	for _, m := range models {
		mCopy := m
		res = append(res, toDomainContact(&mCopy))
	}
	return res, nil
}

func (r *gormContactRepo) DeletePreference(ctx context.Context, contactID uuid.UUID) error {
	return ExtractDB(ctx, r.db).Where("contact_id = ?", contactID).Delete(&contactChannelPreferenceModel{}).Error
}
