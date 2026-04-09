package dto

import "time"

// CreateCampaignRequest represents the parameters for creating a new campaign.
// @Description Payload containing name, optional schedule, and mandatory templates for telegram/sms channels.
type CreateCampaignRequest struct {
	Name        string            `json:"name" validate:"required"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	Templates   map[string]string `json:"templates" validate:"required"`
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

// ErrorResponse represents a standardized API error shape.
// @Description Contains strict business error code and human-readable message.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
