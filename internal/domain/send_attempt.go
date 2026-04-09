package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	AttemptStatusInProgress DeliveryStatus = "in_progress"
	AttemptStatusDelivered  DeliveryStatus = "delivered"
	AttemptStatusFailed     DeliveryStatus = "failed"
)

type SendAttempt struct {
	ID             uuid.UUID
	IdempotencyKey uuid.UUID
	AttemptNumber  int
	CampaignID     uuid.UUID
	ContactID      uuid.UUID
	AccountID      *uuid.UUID // Can be nil if failed before assigning proxy/account
	ProxyID        *uuid.UUID
	Channel        string // "telegram", "sms"
	Status         DeliveryStatus
	ErrorCode      *string
	ErrorMessage   *string
	LatencyMs      *int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *SendAttempt) MarkDelivered(latency int) {
	s.Status = AttemptStatusDelivered
	l := latency
	s.LatencyMs = &l
	s.UpdatedAt = time.Now()
}

func (s *SendAttempt) MarkFailed(code string, message string, latency int) {
	s.Status = AttemptStatusFailed
	c := code
	m := message
	l := latency
	s.ErrorCode = &c
	s.ErrorMessage = &m
	s.LatencyMs = &l
	s.UpdatedAt = time.Now()
}
