package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountState string

const (
	StateWarmingUp AccountState = "warming_up"
	StateActive    AccountState = "active"
	StateCooling   AccountState = "cooling_down"
	StateSuspended AccountState = "suspended"
	StateBanned    AccountState = "banned"
)

type ProxyStatus string

const (
	ProxyHealthy  ProxyStatus = "healthy"
	ProxyDegraded ProxyStatus = "degraded"
)

type Proxy struct {
	ID        uuid.UUID
	Address   string
	Status    ProxyStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Proxy) IsHealthy() bool {
	return p.Status == ProxyHealthy
}

func (p *Proxy) MarkDegraded() {
	p.Status = ProxyDegraded
	p.UpdatedAt = time.Now()
}

type Account struct {
	ID              uuid.UUID
	Phone           string
	Channel         string
	ProxyID         uuid.UUID
	Status          AccountState
	Credentials     string // AES-256-GCM Encrypted
	DailyCheckCount int
	DailySendCount  int
	CooldownUntil   *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (a *Account) TransitionState(newState AccountState) {
	a.Status = newState
	a.UpdatedAt = time.Now()
}

func (a *Account) SetCooldown(until time.Time) {
	a.Status = StateCooling
	a.CooldownUntil = &until
	a.UpdatedAt = time.Now()
}

func (a *Account) IncrementDailySend() {
	a.DailySendCount++
	a.UpdatedAt = time.Now()
}

func (a *Account) IncrementDailyCheck() {
	a.DailyCheckCount++
	a.UpdatedAt = time.Now()
}

func (a *Account) IsInCooldown(now time.Time) bool {
	if a.CooldownUntil == nil {
		return false
	}
	return a.CooldownUntil.After(now)
}

func (a *Account) CanSend(now time.Time) bool {
	if a.Status != StateActive && a.Status != StateWarmingUp {
		return false
	}
	return !a.IsInCooldown(now)
}

type AccountEvent struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	EventType string
	Payload   string
	CreatedAt time.Time
}

type Session struct {
	ID          uuid.UUID
	AccountID   uuid.UUID
	SessionData string // Encrypted gotd session
	UpdatedAt   time.Time
}
