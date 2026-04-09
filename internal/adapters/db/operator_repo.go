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

type operatorSessionModel struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid"`
	Login        string
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

func (operatorSessionModel) TableName() string { return "operator_sessions" }

// -- Repository --

type gormOperatorRepo struct {
	db *gorm.DB
}

func NewOperatorRepository(db *gorm.DB) port.OperatorRepository {
	return &gormOperatorRepo{db: db}
}

// -- Mappings --

func toDomainOperatorSession(m *operatorSessionModel) *domain.OperatorSession {
	return &domain.OperatorSession{
		ID:           m.ID,
		Login:        m.Login,
		RefreshToken: m.RefreshToken,
		ExpiresAt:    m.ExpiresAt,
		CreatedAt:    m.CreatedAt,
	}
}

func fromDomainOperatorSession(s *domain.OperatorSession) *operatorSessionModel {
	return &operatorSessionModel{
		ID:           s.ID,
		Login:        s.Login,
		RefreshToken: s.RefreshToken,
		ExpiresAt:    s.ExpiresAt,
		CreatedAt:    s.CreatedAt,
	}
}

// -- Methods --

func (r *gormOperatorRepo) SaveSession(ctx context.Context, session *domain.OperatorSession) error {
	m := fromDomainOperatorSession(session)
	return ExtractDB(ctx, r.db).Create(m).Error
}

func (r *gormOperatorRepo) GetSessionByToken(ctx context.Context, token string) (*domain.OperatorSession, error) {
	var m operatorSessionModel
	if err := ExtractDB(ctx, r.db).Where("refresh_token = ?", token).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainOperatorSession(&m), nil
}

func (r *gormOperatorRepo) UpdateSession(ctx context.Context, session *domain.OperatorSession) error {
	m := fromDomainOperatorSession(session)
	return ExtractDB(ctx, r.db).Save(m).Error
}
