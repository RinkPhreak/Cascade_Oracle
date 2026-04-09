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

type campaignModel struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	Name        string
	Status      string
	ScheduledAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
func (campaignModel) TableName() string { return "campaigns" }

type messageTemplateModel struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	CampaignID uuid.UUID `gorm:"type:uuid"`
	Channel    string
	Content    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
func (messageTemplateModel) TableName() string { return "message_templates" }

type campaignContactModel struct {
	CampaignID uuid.UUID `gorm:"primaryKey;type:uuid"`
	ContactID  uuid.UUID `gorm:"primaryKey;type:uuid"`
	Status     string
	CreatedAt  time.Time
}
func (campaignContactModel) TableName() string { return "campaign_contacts" }

// -- Repository --

type gormCampaignRepo struct {
	db *gorm.DB
}

func NewCampaignRepository(db *gorm.DB) port.CampaignRepository {
	return &gormCampaignRepo{db: db}
}

// -- Mappings --

func fromDomainCampaign(c *domain.Campaign) *campaignModel {
	return &campaignModel{
		ID:          c.ID,
		Name:        c.Name,
		Status:      string(c.Status),
		ScheduledAt: c.ScheduledAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toDomainCampaign(m *campaignModel) *domain.Campaign {
	return &domain.Campaign{
		ID:          m.ID,
		Name:        m.Name,
		Status:      domain.CampaignStatus(m.Status),
		ScheduledAt: m.ScheduledAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// -- Methods --

func (r *gormCampaignRepo) Create(ctx context.Context, camp *domain.Campaign) error {
	m := fromDomainCampaign(camp)
	return ExtractDB(ctx, r.db).Create(m).Error
}

func (r *gormCampaignRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Campaign, error) {
	var m campaignModel
	if err := ExtractDB(ctx, r.db).First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainCampaign(&m), nil
}

func (r *gormCampaignRepo) Update(ctx context.Context, camp *domain.Campaign) error {
	m := fromDomainCampaign(camp)
	return ExtractDB(ctx, r.db).Save(m).Error
}

func (r *gormCampaignRepo) CreateTemplate(ctx context.Context, tpl *domain.MessageTemplate) error {
	m := &messageTemplateModel{
		ID:         tpl.ID,
		CampaignID: tpl.CampaignID,
		Channel:    tpl.Channel,
		Content:    tpl.Content,
		CreatedAt:  tpl.CreatedAt,
		UpdatedAt:  tpl.UpdatedAt,
	}
	return ExtractDB(ctx, r.db).Create(m).Error
}

func (r *gormCampaignRepo) GetTemplate(ctx context.Context, campaignID uuid.UUID, channel string) (*domain.MessageTemplate, error) {
	var m messageTemplateModel
	if err := ExtractDB(ctx, r.db).Where("campaign_id = ? AND channel = ?", campaignID, channel).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.MessageTemplate{
		ID:         m.ID,
		CampaignID: m.CampaignID,
		Channel:    m.Channel,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}, nil
}

func (r *gormCampaignRepo) AssignContactsToCampaign(ctx context.Context, campaignID uuid.UUID, contactIDs []uuid.UUID, status domain.CampaignContactStatus) error {
	if len(contactIDs) == 0 {
		return nil
	}
	var models []campaignContactModel
	for _, cid := range contactIDs {
		models = append(models, campaignContactModel{
			CampaignID: campaignID,
			ContactID:  cid,
			Status:     string(status),
			CreatedAt:  time.Now(),
		})
	}
	// GORM CreateInBatches with ON CONFLICT DO NOTHING (natively ignores dupes if contact already queued)
	return ExtractDB(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(models, 1000).Error
}

func (r *gormCampaignRepo) UpdateCampaignContactStatus(ctx context.Context, campaignID, contactID uuid.UUID, status domain.CampaignContactStatus) error {
	return ExtractDB(ctx, r.db).Model(&campaignContactModel{}).
		Where("campaign_id = ? AND contact_id = ?", campaignID, contactID).
		Update("status", string(status)).Error
}

func (r *gormCampaignRepo) ListCampaigns(ctx context.Context, status string) ([]*domain.Campaign, error) {
	var models []campaignModel
	query := ExtractDB(ctx, r.db)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	var res []*domain.Campaign
	for _, m := range models {
		mCopy := m
		res = append(res, toDomainCampaign(&mCopy))
	}
	return res, nil
}

func (r *gormCampaignRepo) ListCampaignContacts(ctx context.Context, campaignID uuid.UUID) ([]*domain.CampaignContact, error) {
	var models []campaignContactModel
	if err := ExtractDB(ctx, r.db).Where("campaign_id = ?", campaignID).Find(&models).Error; err != nil {
		return nil, err
	}
	var res []*domain.CampaignContact
	for _, m := range models {
		res = append(res, &domain.CampaignContact{
			CampaignID: m.CampaignID,
			ContactID:  m.ContactID,
			Status:     domain.CampaignContactStatus(m.Status),
		})
	}
	return res, nil
}
