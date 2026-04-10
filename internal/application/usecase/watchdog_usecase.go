package usecase

import (
	"context"
	"log/slog"
	"time"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type StuckAttemptWatchdog struct {
	attemptRepo  port.AttemptRepository
	campaignRepo port.CampaignRepository
	enqueuer     port.TaskEnqueuer
}

func NewStuckAttemptWatchdog(repo port.AttemptRepository, campRepo port.CampaignRepository, enqueuer port.TaskEnqueuer) *StuckAttemptWatchdog {
	return &StuckAttemptWatchdog{
		attemptRepo:  repo,
		campaignRepo: campRepo,
		enqueuer:     enqueuer,
	}
}

func (w *StuckAttemptWatchdog) RecoverStuck(ctx context.Context, timeout time.Duration) error {
	limitTime := time.Now().Add(-timeout)
	
	stuckAttempts, err := w.attemptRepo.GetStuck(ctx, limitTime)
	if err != nil {
		return err
	}

	for _, attempt := range stuckAttempts {
		attempt.MarkFailed("STUCK_TIMEOUT", "Attempt remained in_progress past architectural limit", 0)
		if err := w.attemptRepo.Update(ctx, attempt); err != nil {
			slog.Error("failed to update stuck attempt", "attempt_id", attempt.ID, "error", err)
			continue
		}

		if attempt.Channel == "sms" {
			// SMS is the final fallback. If it's stuck, we fail the entire contact flow.
			w.campaignRepo.UpdateCampaignContactStatus(ctx, attempt.CampaignID, attempt.ContactID, domain.CampaignContactFailed)
			slog.Error("stuck sms attempt failed final step", "attempt_id", attempt.ID)
			continue
		}

		// Requeue for fallback (downstream cascade)
		fallbackChannel := attempt.Channel
		if fallbackChannel == "telegram" {
			fallbackChannel = "sms"
		}

		payload := domain.WaterfallPayload{
			CampaignID: attempt.CampaignID,
			ContactID:  attempt.ContactID,
			Channel:    fallbackChannel,
		}
		
		if err := w.enqueuer.EnqueueWaterfall(ctx, payload, nil); err != nil {
			slog.Error("failed to requeue stuck attempt", "attempt_id", attempt.ID, "error", err)
			continue
		}
	}

	return nil
}
