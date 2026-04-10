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
	var m proxyModel
	if err := ExtractDB(ctx, r.db).First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.Proxy{
		ID:        m.ID,
		Host:      m.Host,
		Port:      m.Port,
		Username:  m.Username,
		Password:  m.Password,
		Status:    domain.ProxyStatus(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}, nil
}

func (r *gormProxyRepo) Create(ctx context.Context, proxy *domain.Proxy) error {
	m := &proxyModel{
		ID:        proxy.ID,
		Host:      proxy.Host,
		Port:      proxy.Port,
		Username:  proxy.Username,
		Password:  proxy.Password,
		Status:    string(proxy.Status),
		CreatedAt: proxy.CreatedAt,
		UpdatedAt: proxy.UpdatedAt,
	}
	return ExtractDB(ctx, r.db).Create(m).Error
}

func (r *gormProxyRepo) GetAll(ctx context.Context) ([]*domain.Proxy, error) {
	var models []proxyModel
	if err := ExtractDB(ctx, r.db).Find(&models).Error; err != nil {
		return nil, err
	}
	var res []*domain.Proxy
	for _, m := range models {
		res = append(res, &domain.Proxy{
			ID:        m.ID,
			Host:      m.Host,
			Port:      m.Port,
			Username:  m.Username,
			Password:  m.Password,
			Status:    domain.ProxyStatus(m.Status),
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	return res, nil
}

func (r *gormProxyRepo) Update(ctx context.Context, proxy *domain.Proxy) error {
	m := &proxyModel{
		ID:        proxy.ID,
		Host:      proxy.Host,
		Port:      proxy.Port,
		Username:  proxy.Username,
		Password:  proxy.Password,
		Status:    string(proxy.Status),
		CreatedAt: proxy.CreatedAt,
		UpdatedAt: proxy.UpdatedAt,
	}
	return ExtractDB(ctx, r.db).Save(m).Error
}
