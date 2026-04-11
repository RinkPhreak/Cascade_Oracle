package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type WaterfallUseCase struct {
	campaignRepo port.CampaignRepository
	contactRepo  port.ContactRepository
	accountRepo  port.AccountRepository
	proxyRepo    port.ProxyRepository
	attemptRepo  port.AttemptRepository
	tgClient     port.TelegramClient
	smsClient    port.SMSClient
	cache        port.Cache
	enqueuer     port.TaskEnqueuer
	crypto       port.CryptoService
}

func NewWaterfallUseCase(
	cmp port.CampaignRepository,
	cnt port.ContactRepository,
	acc port.AccountRepository,
	prx port.ProxyRepository,
	att port.AttemptRepository,
	tg port.TelegramClient,
	sms port.SMSClient,
	ch port.Cache,
	eq port.TaskEnqueuer,
	crypto port.CryptoService,
) *WaterfallUseCase {
	return &WaterfallUseCase{
		campaignRepo: cmp,
		contactRepo:  cnt,
		accountRepo:  acc,
		proxyRepo:    prx,
		attemptRepo:  att,
		tgClient:     tg,
		smsClient:    sms,
		cache:        ch,
		enqueuer:     eq,
		crypto:       crypto,
	}
}

// ProcessContact is the main entry point called by the Asynq worker.
// PERFORMANCE FIX: Instead of blocking on sleeps, it dispatches to step handlers
// that return immediately and re-enqueue the next step with a delay.
func (u *WaterfallUseCase) ProcessContact(ctx context.Context, payload domain.WaterfallPayload) error {
	if u.isSuspended(ctx) {
		return domain.ErrSystemSuspended
	}

	switch payload.Step {
	case domain.StepInit:
		return u.stepInit(ctx, payload)
	case domain.StepPresenceCheck:
		return u.stepPresenceCheck(ctx, payload)
	case domain.StepSend:
		return u.stepSend(ctx, payload)
	case domain.StepRetry:
		return u.stepRetry(ctx, payload)
	default:
		return fmt.Errorf("unknown waterfall step: %d", payload.Step)
	}
}

// stepInit resolves channels, picks an account, checks cache, and either
// fast-paths or schedules the presence check with a delay.
func (u *WaterfallUseCase) stepInit(ctx context.Context, payload domain.WaterfallPayload) error {
	contact, err := u.contactRepo.GetByID(ctx, payload.ContactID)
	if err != nil {
		return err
	}

	campaign, err := u.campaignRepo.GetByID(ctx, payload.CampaignID)
	if err != nil {
		return err
	}

	if contact.HasReplied {
		return u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactReplied)
	}

	prefChannel, _ := u.contactRepo.GetPreference(ctx, contact.ID)
	channels := u.reorderWaterfall(prefChannel)

	for _, channel := range channels {
		if channel == "telegram" && u.isPoolCritical(ctx) {
			continue
		}

		if channel == "telegram" {
			account, err := u.accountRepo.GetLeastBusyActiveAccount(ctx, channel)
			if err != nil {
				continue
			}
			if !account.CanSend(time.Now()) {
				continue
			}
			prx, err := u.proxyRepo.GetByID(ctx, account.ProxyID)
			if err != nil || !prx.IsHealthy() {
				continue
			}

			plainPhone, err := u.crypto.Decrypt(contact.Phone)
			if err != nil {
				continue
			}

			// Check presence cache before scheduling the delay
			keyPos := "tg:check:positive:" + plainPhone
			if exists, _ := u.cache.Exists(ctx, keyPos); exists {
				// Cached positive — skip presence check, go directly to send with pre-send delay
				return u.enqueueDeferred(ctx, domain.WaterfallPayload{
					CampaignID: payload.CampaignID,
					ContactID:  payload.ContactID,
					Step:       domain.StepSend,
					Channel:    channel,
					AccountID:  account.ID,
				}, time.Duration(8+rand.Intn(17))*time.Second)
			}

			keyNeg := "tg:check:negative:" + plainPhone
			if exists, _ := u.cache.Exists(ctx, keyNeg); exists {
				// Cached negative — skip telegram
				continue
			}

			if account.DailyCheckCount >= 100 {
				continue
			}

			// PERFORMANCE FIX: Schedule presence check after humanizing delay instead of blocking here
			return u.enqueueDeferred(ctx, domain.WaterfallPayload{
				CampaignID: payload.CampaignID,
				ContactID:  payload.ContactID,
				Step:       domain.StepPresenceCheck,
				Channel:    channel,
				AccountID:  account.ID,
			}, time.Duration(45+rand.Intn(30))*time.Second)
		}

		// SMS channel — no presence check needed
		return u.enqueueDeferred(ctx, domain.WaterfallPayload{
			CampaignID: payload.CampaignID,
			ContactID:  payload.ContactID,
			Step:       domain.StepSend,
			Channel:    channel,
		}, time.Duration(8+rand.Intn(17))*time.Second)
	}

	// All channels exhausted at init
	u.campaignRepo.UpdateCampaignContactStatus(ctx, payload.CampaignID, payload.ContactID, domain.CampaignContactFailed)
	return nil
}

// stepPresenceCheck performs the actual ImportContacts to check TG presence.
// BLOCKER 2 FIX: Does NOT call DeleteContacts here. Imported user IDs are passed
// forward in the payload and cleaned up after send completes.
func (u *WaterfallUseCase) stepPresenceCheck(ctx context.Context, payload domain.WaterfallPayload) error {
	contact, err := u.contactRepo.GetByID(ctx, payload.ContactID)
	if err != nil {
		return err
	}

	plainPhone, err := u.crypto.Decrypt(contact.Phone)
	if err != nil {
		return err
	}

	account, err := u.accountRepo.GetAccountByID(ctx, payload.AccountID)
	if err != nil {
		return err
	}

	imported, err := u.tgClient.ImportContacts(ctx, account.ID, []string{plainPhone})
	if err != nil {
		// Presence check failed — escalate to SMS by re-entering stepInit without TG
		slog.Warn("presence check ImportContacts failed", "error", err, "account_id", account.ID)
		u.campaignRepo.UpdateCampaignContactStatus(ctx, payload.CampaignID, payload.ContactID, domain.CampaignContactFailed)
		return nil
	}

	var importedUserIDs []int64
	for _, imp := range imported {
		importedUserIDs = append(importedUserIDs, imp.UserID)
	}

	found := len(imported) > 0

	// Update cache and daily check counter
	if found {
		u.cache.Set(ctx, "tg:check:positive:"+plainPhone, "1", 6*time.Hour)
	} else {
		u.cache.Set(ctx, "tg:check:negative:"+plainPhone, "1", 20*time.Minute)
	}
	account.IncrementDailyCheck()
	_ = u.accountRepo.UpdateAccount(ctx, account)

	if !found {
		// Not on Telegram — record failure, let the channel loop in a new init handle SMS
		// Create a failed attempt record
		ikey := u.generateIdempotencyKey(contact.ID, payload.CampaignID, payload.Channel)
		attID, _ := uuid.NewV7()
		attempt := &domain.SendAttempt{
			ID:             attID,
			IdempotencyKey: ikey,
			AttemptNumber:  1,
			CampaignID:     payload.CampaignID,
			ContactID:      contact.ID,
			Channel:        payload.Channel,
			AccountID:      &account.ID,
			Status:         domain.AttemptStatusFailed,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		u.failAttempt(ctx, attempt, "NOT_FOUND", "User not on Telegram")
		_ = u.attemptRepo.Upsert(ctx, attempt)

		// Try SMS fallback by scheduling a new send step
		return u.enqueueDeferred(ctx, domain.WaterfallPayload{
			CampaignID: payload.CampaignID,
			ContactID:  payload.ContactID,
			Step:       domain.StepSend,
			Channel:    "sms",
		}, time.Duration(8+rand.Intn(17))*time.Second)
	}

	// Found on Telegram — schedule the actual send with delay
	// BLOCKER 2 FIX: Pass importedUserIDs forward; they'll be cleaned up after send
	return u.enqueueDeferred(ctx, domain.WaterfallPayload{
		CampaignID:      payload.CampaignID,
		ContactID:       payload.ContactID,
		Step:            domain.StepSend,
		Channel:         payload.Channel,
		AccountID:       payload.AccountID,
		ImportedUserIDs: importedUserIDs,
	}, time.Duration(8+rand.Intn(17))*time.Second)
}

// stepSend performs a single send attempt without blocking sleeps.
func (u *WaterfallUseCase) stepSend(ctx context.Context, payload domain.WaterfallPayload) error {
	return u.doSend(ctx, payload, 0)
}

// stepRetry performs a retry send attempt.
func (u *WaterfallUseCase) stepRetry(ctx context.Context, payload domain.WaterfallPayload) error {
	return u.doSend(ctx, payload, payload.RetryIndex)
}

// doSend is the unified send logic for both initial and retry attempts.
func (u *WaterfallUseCase) doSend(ctx context.Context, payload domain.WaterfallPayload, retryIdx int) error {
	contact, err := u.contactRepo.GetByID(ctx, payload.ContactID)
	if err != nil {
		return err
	}

	campaign, err := u.campaignRepo.GetByID(ctx, payload.CampaignID)
	if err != nil {
		return err
	}

	plainPhone, err := u.crypto.Decrypt(contact.Phone)
	if err != nil {
		u.cleanupImportedContacts(ctx, payload)
		u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactFailed)
		return nil
	}

	// NITPICK 2 FIX: When name is empty, use an impersonal greeting instead of "User"
	// which signals spam and increases ban risk.
	plainName, _ := u.crypto.Decrypt(contact.Name)

	template, err := u.campaignRepo.GetTemplate(ctx, campaign.ID, payload.Channel)
	if err != nil {
		u.cleanupImportedContacts(ctx, payload)
		u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactFailed)
		return nil
	}

	// Render template with name-aware greeting
	text := template.Content
	if plainName != "" {
		text = strings.ReplaceAll(text, "{{Greeting}}", "Здравствуйте, "+plainName+"!")
		text = strings.ReplaceAll(text, "{{Name}}", plainName)
	} else {
		text = strings.ReplaceAll(text, "{{Greeting}}", "Добрый день!")
		text = strings.ReplaceAll(text, "{{Name}}", "")
	}
	text = strings.ReplaceAll(text, "{{Phone}}", plainPhone)

	// Upsert attempt record
	ikey := u.generateIdempotencyKey(contact.ID, campaign.ID, payload.Channel)
	var attempt *domain.SendAttempt

	if payload.AttemptID != uuid.Nil {
		attempt, err = u.attemptRepo.GetByID(ctx, payload.AttemptID)
		if err != nil {
			return err
		}
	} else {
		attID, _ := uuid.NewV7()
		attempt = &domain.SendAttempt{
			ID:             attID,
			IdempotencyKey: ikey,
			AttemptNumber:  retryIdx + 1,
			CampaignID:     campaign.ID,
			ContactID:      contact.ID,
			Channel:        payload.Channel,
			Status:         domain.AttemptStatusInProgress,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if payload.AccountID != uuid.Nil {
			attempt.AccountID = &payload.AccountID
		}
		if err := u.attemptRepo.Upsert(ctx, attempt); err != nil {
			return err
		}
	}

	// Check idempotency guard
	existing, err := u.attemptRepo.GetByIdempotencyKey(ctx, ikey)
	if err == nil && existing.Status == domain.AttemptStatusDelivered {
		u.cleanupImportedContacts(ctx, payload)
		return nil
	}

	attempt.AttemptNumber = retryIdx + 1

	var latency int
	var sendErr error

	if payload.Channel == "telegram" {
		account, accErr := u.accountRepo.GetAccountByID(ctx, payload.AccountID)
		if accErr != nil {
			u.cleanupImportedContacts(ctx, payload)
			u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactFailed)
			return nil
		}

		hourlyKey := fmt.Sprintf("cascade:account:%s:hourly", account.ID.String())
		hCount, _ := u.cache.Increment(ctx, hourlyKey)
		if hCount == 1 {
			u.cache.Expire(ctx, hourlyKey, time.Hour)
		}

		if hCount > 8 {
			slog.Warn("hourly burst cap reached", "account_id", account.ID, "contact_id", attempt.ContactID)
			u.failAttempt(ctx, attempt, "RATE_LIMIT", domain.ErrRateLimit.Error())
			u.cleanupImportedContacts(ctx, payload)
			u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactFailed)
			return nil
		}

		latency, sendErr = u.tgClient.Send(ctx, account.ID, plainPhone, text)

		// Handle fatal TG errors
		if sendErr != nil {
			if errors.Is(sendErr, domain.ErrPeerFlood) {
				acc, _ := u.accountRepo.GetAccountByID(ctx, account.ID)
				if acc != nil {
					acc.SetCooldown(time.Now().Add(24 * time.Hour))
					u.accountRepo.UpdateAccount(ctx, acc)
				}
				u.failAttempt(ctx, attempt, "PEER_FLOOD", sendErr.Error())
				u.cleanupImportedContacts(ctx, payload)
				u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactFailed)
				return nil
			}
			if errors.Is(sendErr, domain.ErrAuthUnregistered) {
				acc, _ := u.accountRepo.GetAccountByID(ctx, account.ID)
				if acc != nil {
					acc.TransitionState(domain.StateBanned)
					u.accountRepo.UpdateAccount(ctx, acc)
				}
				u.failAttempt(ctx, attempt, "BANNED", sendErr.Error())
				u.cleanupImportedContacts(ctx, payload)
				u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactFailed)
				return nil
			}
			if errors.Is(sendErr, domain.ErrUserNotFound) {
				u.failAttempt(ctx, attempt, "NOT_FOUND", sendErr.Error())
				u.cleanupImportedContacts(ctx, payload)
				// Fallback to SMS
				return u.enqueueDeferred(ctx, domain.WaterfallPayload{
					CampaignID: payload.CampaignID,
					ContactID:  payload.ContactID,
					Step:       domain.StepSend,
					Channel:    "sms",
				}, time.Duration(8+rand.Intn(17))*time.Second)
			}
		}
	} else {
		// SMS channel
		latency, sendErr = u.smsClient.Send(ctx, plainPhone, text)
	}

	if sendErr == nil {
		// Success!
		attempt.MarkDelivered(latency)
		u.attemptRepo.Update(ctx, attempt)

		if payload.AccountID != uuid.Nil {
			acc, _ := u.accountRepo.GetAccountByID(ctx, payload.AccountID)
			if acc != nil {
				acc.IncrementDailySend()
				u.accountRepo.UpdateAccount(ctx, acc)
			}
		}

		pref := &domain.ContactChannelPreference{
			ContactID:        contact.ID,
			PreferredChannel: payload.Channel,
			UpdatedAt:        time.Now(),
		}
		_ = u.contactRepo.SavePreference(ctx, pref)

		// BLOCKER 2 FIX: Clean up imported contacts AFTER successful send
		u.cleanupImportedContacts(ctx, payload)

		u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactCompleted)
		return nil
	}

	// Send failed — check if we can retry
	maxRetries := 2
	backoffs := []time.Duration{30 * time.Second, 90 * time.Second}

	if retryIdx < maxRetries {
		// PERFORMANCE FIX: Schedule retry with backoff delay instead of blocking sleep
		return u.enqueueDeferred(ctx, domain.WaterfallPayload{
			CampaignID:      payload.CampaignID,
			ContactID:       payload.ContactID,
			Step:            domain.StepRetry,
			Channel:         payload.Channel,
			AccountID:       payload.AccountID,
			AttemptID:       attempt.ID,
			ImportedUserIDs: payload.ImportedUserIDs,
			RetryIndex:      retryIdx + 1,
		}, backoffs[retryIdx])
	}

	// All retries exhausted
	u.failAttempt(ctx, attempt, "EXHAUSTED", sendErr.Error())
	u.cleanupImportedContacts(ctx, payload)
	u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactFailed)
	return nil
}

// cleanupImportedContacts deletes imported contacts from the TG account's contact list.
// BLOCKER 2 FIX: This is called at the END of the processing flow (success or failure),
// not immediately after presence check. This prevents the phantom contact problem where
// DeleteContacts removes the access_hash needed by ResolvePhone in Send.
func (u *WaterfallUseCase) cleanupImportedContacts(ctx context.Context, payload domain.WaterfallPayload) {
	if len(payload.ImportedUserIDs) > 0 && payload.AccountID != uuid.Nil {
		if err := u.tgClient.DeleteContacts(ctx, payload.AccountID, payload.ImportedUserIDs); err != nil {
			slog.Warn("failed to cleanup imported contacts", "error", err, "account_id", payload.AccountID)
		}
	}
}

// enqueueDeferred schedules the next step with a delay. The worker returns immediately.
// PERFORMANCE FIX: Replaces all safeSleep calls that were blocking Asynq worker goroutines.
func (u *WaterfallUseCase) enqueueDeferred(ctx context.Context, payload domain.WaterfallPayload, delay time.Duration) error {
	processAt := time.Now().Add(delay)
	return u.enqueuer.EnqueueWaterfall(ctx, payload, &processAt)
}

func (u *WaterfallUseCase) isSuspended(ctx context.Context) bool {
	halted, _ := u.cache.Exists(ctx, "cascade:system:halted")
	if halted {
		return true
	}
	memCrit, _ := u.cache.Exists(ctx, "cascade:memory:critical")
	if memCrit {
		return true
	}
	return false
}

func (u *WaterfallUseCase) isPoolCritical(ctx context.Context) bool {
	critical, _ := u.cache.Exists(ctx, "cascade:pool:critical")
	return critical
}

func (u *WaterfallUseCase) reorderWaterfall(prefChannel string) []string {
	if prefChannel == "sms" {
		return []string{"sms"} // If specifically asked for sms over TG (maybe previously hit on sms)
	}
	return []string{"telegram", "sms"}
}

func (u *WaterfallUseCase) generateIdempotencyKey(contact, campaign uuid.UUID, channel string) uuid.UUID {
	return uuid.NewMD5(uuid.NameSpaceURL, []byte(contact.String()+campaign.String()+channel))
}

func (u *WaterfallUseCase) failAttempt(ctx context.Context, attempt *domain.SendAttempt, code, msg string) {
	attempt.MarkFailed(code, msg, 0)
	_ = u.attemptRepo.Update(ctx, attempt)
}
