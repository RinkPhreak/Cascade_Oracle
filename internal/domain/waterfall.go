package domain

import "github.com/google/uuid"

// WaterfallPayload defines the structure of the data serialized into the queue 
// mechanism (Asynq) when triggering the waterfall process for a lead.
type WaterfallPayload struct {
	CampaignID uuid.UUID `json:"campaign_id"`
	ContactID  uuid.UUID `json:"contact_id"`
	Channel    string    `json:"channel,omitempty"`
}
