package usecase

import (
	"context"

	"cascade/internal/application/port"
)

const poolCriticalKey = "cascade:pool:critical"

type MassBanDetector struct {
	accountRepo   port.AccountRepository
	cache         port.Cache
	minThreshold  int
}

func NewMassBanDetector(repo port.AccountRepository, cache port.Cache, minThreshold int) *MassBanDetector {
	return &MassBanDetector{
		accountRepo:  repo,
		cache:        cache,
		minThreshold: minThreshold,
	}
}

func (m *MassBanDetector) EvaluatePool(ctx context.Context) error {
	count, err := m.accountRepo.CountActiveAccounts(ctx)
	if err != nil {
		return err
	}

	if count < m.minThreshold {
		// Set pool critical flag without expiration (persists until recovered)
		return m.cache.Set(ctx, poolCriticalKey, "1", 0)
	}

	exists, err := m.cache.Exists(ctx, poolCriticalKey)
	if err != nil {
		return err
	}
	if exists {
		return m.cache.Del(ctx, poolCriticalKey)
	}

	return nil
}
