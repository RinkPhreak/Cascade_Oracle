package db

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type gormProxyRepo struct {
	db *gorm.DB
}

func NewProxyRepository(db *gorm.DB) port.ProxyRepository {
	return &gormProxyRepo{db: db}
}

func (r *gormProxyRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Proxy, error) {
	return &domain.Proxy{ID: id, Status: "active"}, nil
}

func (r *gormProxyRepo) Save(ctx context.Context, proxy *domain.Proxy) error {
	return nil
}

func (r *gormProxyRepo) Create(ctx context.Context, proxy *domain.Proxy) error {
	return nil
}

func (r *gormProxyRepo) GetAll(ctx context.Context) ([]*domain.Proxy, error) {
	return nil, nil
}

func (r *gormProxyRepo) Update(ctx context.Context, proxy *domain.Proxy) error {
	return nil
}
