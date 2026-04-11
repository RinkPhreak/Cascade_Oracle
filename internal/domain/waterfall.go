package domain

import "github.com/google/uuid"

// WaterfallPayload defines the structure of the data serialized into the queue
// mechanism (Asynq) when triggering the waterfall process for a lead.
// Step constants for the deferred state machine.
// Workers return immediately after scheduling the next step with a delay,
// instead of blocking on sleeps inside the Asynq goroutine.
const (
	StepInit          = 0 // Entry: resolve channel, pick account, check presence cache
	StepPresenceCheck = 1 // After pre-check delay: ImportContacts to verify TG presence
	StepSend          = 2 // After pre-send delay: actually send the message
	StepRetry         = 3 // After backoff delay: retry sending
)

type WaterfallPayload struct {
	CampaignID uuid.UUID `json:"campaign_id"`
	ContactID  uuid.UUID `json:"contact_id"`

	// Step-machine state — allows deferred scheduling instead of blocking sleeps
	Step            int       `json:"step"`
	Channel         string    `json:"channel,omitempty"`
	AccountID       uuid.UUID `json:"account_id,omitempty"`
	AttemptID       uuid.UUID `json:"attempt_id,omitempty"`
	ImportedUserIDs []int64   `json:"imported_user_ids,omitempty"`
	RetryIndex      int       `json:"retry_index,omitempty"`
}
