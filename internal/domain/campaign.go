package domain

import (
	"time"

	"github.com/google/uuid"
)

type CampaignStatus string

const (
	CampaignStatusDraft    CampaignStatus = "draft"
	CampaignStatusActive   CampaignStatus = "active"
	CampaignStatusPaused   CampaignStatus = "paused"
	CampaignStatusFinished CampaignStatus = "finished"
)

type Campaign struct {
	ID          uuid.UUID
	Name        string
	Status      CampaignStatus
	ScheduledAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c *Campaign) IsScheduledForFuture(now time.Time) bool {
	return c.ScheduledAt != nil && c.ScheduledAt.After(now)
}

func (c *Campaign) TransitionStatus(status CampaignStatus) {
	c.Status = status
	c.UpdatedAt = time.Now()
}

type MessageTemplate struct {
	ID        uuid.UUID
	CampaignID uuid.UUID
	Channel   string // "telegram", "sms"
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CampaignContactStatus string

const (
	CampaignContactPending    CampaignContactStatus = "pending"
	CampaignContactInProgress CampaignContactStatus = "in_progress"
	CampaignContactReplied    CampaignContactStatus = "replied"
	CampaignContactCompleted  CampaignContactStatus = "completed"
	CampaignContactFailed     CampaignContactStatus = "failed"
)

type CampaignContact struct {
	CampaignID uuid.UUID
	ContactID  uuid.UUID
	Status     CampaignContactStatus
}

func (cc *CampaignContact) TransitionStatus(status CampaignContactStatus) {
	cc.Status = status
}
