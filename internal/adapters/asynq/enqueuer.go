package asynq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type asynqEnqueuer struct {
	client *asynq.Client
}

func NewAsynqEnqueuer(redisAddr string) port.TaskEnqueuer {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &asynqEnqueuer{client: client}
}

func (e *asynqEnqueuer) EnqueueWaterfall(ctx context.Context, payload domain.WaterfallPayload, processAt *time.Time) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask("cascade:waterfall:process", b)
	var opts []asynq.Option
	
	// If campaign is scheduled for the future
	if processAt != nil && processAt.After(time.Now()) {
		opts = append(opts, asynq.ProcessAt(*processAt))
	}

	opts = append(opts, asynq.Queue("default"), asynq.MaxRetry(25))

	_, err = e.client.EnqueueContext(ctx, task, opts...)
	return err
}

func (e *asynqEnqueuer) CancelCampaignTasks(ctx context.Context, campaignID uuid.UUID) error {
	// Implementing this fully requires querying Inspector to delete particular queued items.
	// For simplicity in Phase 1, we can rely on Campaign state logic blocking execution in workers.
	// But here's a placeholder returning nil.
	return nil
}
