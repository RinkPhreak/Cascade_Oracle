package port

import (
	"context"
	"time"

	"cascade/internal/domain"
	"github.com/google/uuid"
)

// TaskEnqueuer enforces bounded logic over Asynq payload dispatching.
type TaskEnqueuer interface {
	EnqueueWaterfall(ctx context.Context, payload domain.WaterfallPayload, processAt *time.Time) error
	CancelCampaignTasks(ctx context.Context, campaignID uuid.UUID) error
}
