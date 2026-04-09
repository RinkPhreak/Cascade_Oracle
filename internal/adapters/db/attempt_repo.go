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

type sendAttemptModel struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid"`
	IdempotencyKey uuid.UUID `gorm:"uniqueIndex;type:uuid"`
	ContactID      uuid.UUID `gorm:"type:uuid"`
	CampaignID     uuid.UUID `gorm:"type:uuid"`
	AccountID      *uuid.UUID `gorm:"type:uuid"`
	ProxyID        *uuid.UUID `gorm:"type:uuid"`
	Channel        string
	Status         string
	ErrorCode      *string
	ErrorMessage   *string
	LatencyMs      *int
	AttemptNumber  int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (sendAttemptModel) TableName() string { return "send_attempts" }

// -- Repository --

type gormAttemptRepo struct {
	db *gorm.DB
}

func NewAttemptRepository(db *gorm.DB) port.AttemptRepository {
	return &gormAttemptRepo{db: db}
}

// -- Mappings --

func fromDomainAttempt(a *domain.SendAttempt) *sendAttemptModel {
	m := &sendAttemptModel{
		ID:             a.ID,
		IdempotencyKey: a.IdempotencyKey,
		ContactID:      a.ContactID,
		CampaignID:     a.CampaignID,
		AccountID:      a.AccountID,
		ProxyID:        a.ProxyID,
		Channel:        a.Channel,
		Status:         string(a.Status),
		AttemptNumber:  a.AttemptNumber,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
	if a.ErrorCode != "" {
		m.ErrorCode = &a.ErrorCode
	}
	if a.ErrorMessage != "" {
		m.ErrorMessage = &a.ErrorMessage
	}
	if a.LatencyMs > 0 {
		m.LatencyMs = &a.LatencyMs
	}
	return m
}

func toDomainAttempt(m *sendAttemptModel) *domain.SendAttempt {
	a := &domain.SendAttempt{
		ID:             m.ID,
		IdempotencyKey: m.IdempotencyKey,
		ContactID:      m.ContactID,
		CampaignID:     m.CampaignID,
		AccountID:      m.AccountID,
		ProxyID:        m.ProxyID,
		Channel:        m.Channel,
		Status:         domain.AttemptStatus(m.Status),
		AttemptNumber:  m.AttemptNumber,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if m.ErrorCode != nil {
		a.ErrorCode = *m.ErrorCode
	}
	if m.ErrorMessage != nil {
		a.ErrorMessage = *m.ErrorMessage
	}
	if m.LatencyMs != nil {
		a.LatencyMs = *m.LatencyMs
	}
	return a
}

// -- Methods --

func (r *gormAttemptRepo) Upsert(ctx context.Context, attempt *domain.SendAttempt) error {
	m := fromDomainAttempt(attempt)
	return ExtractDB(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).Create(m).Error
}

func (r *gormAttemptRepo) Update(ctx context.Context, attempt *domain.SendAttempt) error {
	m := fromDomainAttempt(attempt)
	return ExtractDB(ctx, r.db).Save(m).Error
}

func (r *gormAttemptRepo) GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (*domain.SendAttempt, error) {
	var m sendAttemptModel
	if err := ExtractDB(ctx, r.db).Where("idempotency_key = ?", key).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainAttempt(&m), nil
}

func (r *gormAttemptRepo) GetStuck(ctx context.Context, olderThan time.Time) ([]*domain.SendAttempt, error) {
	var models []sendAttemptModel
	err := ExtractDB(ctx, r.db).
		Where("status = ? AND updated_at < ?", domain.AttemptStatusInProgress, olderThan).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	var res []*domain.SendAttempt
	for _, m := range models {
		mCopy := m
		res = append(res, toDomainAttempt(&mCopy))
	}
	return res, nil
}
