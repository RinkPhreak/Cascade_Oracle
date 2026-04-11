package dto

import "time"

// CreateCampaignRequest represents the parameters for creating a new campaign.
// @Description Payload containing name, optional schedule, and mandatory templates for telegram/sms channels.
type CreateCampaignRequest struct {
	Name        string            `json:"name" validate:"required"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	Templates   map[string]string `json:"templates"` // <--- Убрали validate:"required"
}

// CampaignResponse represents the serialized view of a Campaign entity.
// @Description Summary view of a created or fetched campaign.
type CampaignResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CampaignStatsResponse struct {
	Total          int            `json:"total"`
	Completed      int            `json:"completed"`
	Replied        int            `json:"replied"`
	Failed         int            `json:"failed"`
	TGAttempted    int            `json:"tg_attempted"`
	SMSAttempted   int            `json:"sms_attempted"`
	ErrorBreakdown map[string]int `json:"error_breakdown"`
}

type CampaignTaskResponse struct {
	ID            string `json:"id"`
	ContactID     string `json:"contact_id"`
	Channel       string `json:"channel"`
	Status        string `json:"status"`
	AttemptNumber int    `json:"attempt_number"`
	UpdatedAt     string `json:"updated_at"`
}
