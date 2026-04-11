package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"cascade/internal/application/usecase"
	"cascade/internal/domain"
)

// HandleWaterfallTask maps asynq bytes into domains and drives the cascade
func HandleWaterfallTask(uc *usecase.WaterfallUseCase) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload domain.WaterfallPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			slog.Error("failed to decode waterfall payload", "error", err)
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		err := uc.ProcessContact(ctx, payload)
		if err != nil {
			// If system is suspended, pass the error back so Asynq can Exponential Backoff & Retry
			if errors.Is(err, domain.ErrSystemSuspended) {
				slog.Warn("system suspended, deferring via asynq retry", "campaign_id", payload.CampaignID)
				return err
			}

			// If something else goes wrong contextually, wait.
			slog.Error("waterfall process runtime fault", "error", err, "contact_id", payload.ContactID)
			return err
		}

		return nil
	}
}
