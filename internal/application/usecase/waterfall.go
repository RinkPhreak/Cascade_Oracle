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

func (u *WaterfallUseCase) ProcessContact(ctx context.Context, payload domain.WaterfallPayload) error {
	if u.isSuspended(ctx) {
		return domain.ErrSystemSuspended
	}

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
			continue // graceful cascade
		}

		var account *domain.Account
		var proxy *domain.Proxy

		if channel == "telegram" {
			account, err = u.accountRepo.GetLeastBusyActiveAccount(ctx, channel)
			if err != nil {
				continue // Exhausted active pool, fallback to SMS
			}
			if !account.CanSend(time.Now()) {
				continue
			}
			proxy, err = u.proxyRepo.GetByID(ctx, account.ProxyID)
			if err != nil || !proxy.IsHealthy() {
				u.accountRepo.UpdateAccount(ctx, account) // Or degrade proxy, simplified
				continue
			}
		}

		ikey := u.generateIdempotencyKey(contact.ID, campaign.ID, channel)
		
		attID, _ := uuid.NewV7()
		attempt := &domain.SendAttempt{
			ID:             attID,
			IdempotencyKey: ikey,
			AttemptNumber:  1,
			CampaignID:     campaign.ID,
			ContactID:      contact.ID,
			Channel:        channel,
			Status:         domain.AttemptStatusInProgress,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if account != nil {
			attempt.AccountID = &account.ID
		}
		if proxy != nil {
			attempt.ProxyID = &proxy.ID
		}

		if err := u.attemptRepo.Upsert(ctx, attempt); err != nil {
			return err
		}

		existing, err := u.attemptRepo.GetByIdempotencyKey(ctx, ikey)
		if err != nil {
			return err
		}
		if existing.Status == domain.AttemptStatusDelivered {
			return nil // guard
		}

		plainPhone, err := u.crypto.Decrypt(contact.Phone)
		if err != nil {
			continue
		}

		// Decrypt Name for templating
		plainName, _ := u.crypto.Decrypt(contact.Name)
		if plainName == "" { plainName = "User" }

		if channel == "telegram" {
			presence, pErr := u.checkTelegramPresence(ctx, account, plainPhone)
			if pErr != nil || !presence {
				u.failAttempt(ctx, existing, "NOT_FOUND", "User not on Telegram or cap reached")
				continue
			}
		}

		template, err := u.campaignRepo.GetTemplate(ctx, campaign.ID, channel)
		if err != nil {
			u.failAttempt(ctx, existing, "NO_TEMPLATE", "Template missing")
			continue
		}

		// Tier 2.4: Render Template
		text := template.Content
		text = strings.ReplaceAll(text, "{{Name}}", plainName)
		text = strings.ReplaceAll(text, "{{Phone}}", plainPhone)

		result, resultErr := u.sendWithRetry(ctx, channel, account, existing, plainPhone, text)
		if resultErr == nil && result == domain.AttemptStatusDelivered {
			pref := &domain.ContactChannelPreference{
				ContactID:        contact.ID,
				PreferredChannel: channel,
				UpdatedAt:        time.Now(),
			}
			_ = u.contactRepo.SavePreference(ctx, pref)
			
			u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactCompleted)
			return nil
		}
	}

	u.campaignRepo.UpdateCampaignContactStatus(ctx, campaign.ID, contact.ID, domain.CampaignContactFailed)
	return nil
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

func (u *WaterfallUseCase) safeSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (u *WaterfallUseCase) failAttempt(ctx context.Context, attempt *domain.SendAttempt, code, msg string) {
	attempt.MarkFailed(code, msg, 0)
	_ = u.attemptRepo.Update(ctx, attempt)
}

func (u *WaterfallUseCase) checkTelegramPresence(ctx context.Context, account *domain.Account, phone string) (bool, error) {
	if account.DailyCheckCount >= 100 {
		return false, domain.ErrDailyCheckCap
	}
	
	keyPos := "tg:check:positive:" + phone
	if exists, _ := u.cache.Exists(ctx, keyPos); exists {
		return true, nil
	}
	
	keyNeg := "tg:check:negative:" + phone
	if exists, _ := u.cache.Exists(ctx, keyNeg); exists {
		return false, nil
	}
	
	if err := u.safeSleep(ctx, time.Duration(45+rand.Intn(30))*time.Second); err != nil {
		return false, err
	}
	
	imported, err := u.tgClient.ImportContacts(ctx, account.ID, []string{phone})
	if err != nil {
		return false, err
	}
	
	var userIDs []int64
	for _, imp := range imported {
		userIDs = append(userIDs, imp.UserID)
	}
	
	if len(userIDs) > 0 {
		_ = u.tgClient.DeleteContacts(ctx, account.ID, userIDs)
	}

	found := len(imported) > 0
	
	if found {
		u.cache.Set(ctx, keyPos, "1", 6*time.Hour)
		account.IncrementDailyCheck()
		_ = u.accountRepo.UpdateAccount(ctx, account)
	} else {
		u.cache.Set(ctx, keyNeg, "1", 20*time.Minute)
		account.IncrementDailyCheck()
		_ = u.accountRepo.UpdateAccount(ctx, account)
	}
	
	return found, nil
}

func (u *WaterfallUseCase) sendWithRetry(
	ctx context.Context,
	channel string,
	account *domain.Account,
	attempt *domain.SendAttempt,
	phone string,
	content string,
) (domain.DeliveryStatus, error) {
	
	maxRetries := 2
	backoffs := []time.Duration{30 * time.Second, 90 * time.Second}

	for i := 0; i <= maxRetries; i++ {
		attempt.AttemptNumber = i + 1

		var latency int
		var err error
		
		if err = u.safeSleep(ctx, time.Duration(8+rand.Intn(17))*time.Second); err != nil {
			return domain.AttemptStatusFailed, err
		}
		
		if channel == "telegram" {
			hourlyKey := fmt.Sprintf("cascade:account:%s:hourly", account.ID.String())
			hCount, _ := u.cache.Increment(ctx, hourlyKey)
			if hCount == 1 {
				u.cache.Expire(ctx, hourlyKey, time.Hour)
			}

			if hCount > 8 {
				err = domain.ErrRateLimit
				slog.Warn("hourly burst cap reached", "account_id", account.ID, "contact_id", attempt.ContactID)
				u.failAttempt(ctx, attempt, "RATE_LIMIT", err.Error())
				return domain.AttemptStatusFailed, err
			} else {
				latency, err = u.tgClient.Send(ctx, account.ID, phone, content)
			}
		} else {
			latency, err = u.smsClient.Send(ctx, phone, content)
		}

		if err == nil {
			attempt.MarkDelivered(latency)
			u.attemptRepo.Update(ctx, attempt)
			
			if account != nil {
				acc, _ := u.accountRepo.GetAccountByID(ctx, account.ID)
				if acc != nil {
					acc.IncrementDailySend()
					u.accountRepo.UpdateAccount(ctx, acc)
				}
			}
			return domain.AttemptStatusDelivered, nil
		}

		if channel == "telegram" {
			if errors.Is(err, domain.ErrPeerFlood) {
				acc, _ := u.accountRepo.GetAccountByID(ctx, account.ID)
				if acc != nil {
					acc.SetCooldown(time.Now().Add(24 * time.Hour))
					u.accountRepo.UpdateAccount(ctx, acc)
				}
				u.failAttempt(ctx, attempt, "PEER_FLOOD", err.Error())
				return domain.AttemptStatusFailed, err // Escalate so waterfall loop switches account
			}
			if errors.Is(err, domain.ErrAuthUnregistered) {
				acc, _ := u.accountRepo.GetAccountByID(ctx, account.ID)
				if acc != nil {
					acc.TransitionState(domain.StateBanned)
					u.accountRepo.UpdateAccount(ctx, acc)
				}
				u.failAttempt(ctx, attempt, "BANNED", err.Error())
				return domain.AttemptStatusFailed, err // Escalate
			}
			if errors.Is(err, domain.ErrUserNotFound) {
				u.failAttempt(ctx, attempt, "NOT_FOUND", err.Error())
				return domain.AttemptStatusFailed, nil // Not retryable, fallback to SMS gracefully
			}
		}

		if i < maxRetries {
			if err = u.safeSleep(ctx, backoffs[i]); err != nil {
				return domain.AttemptStatusFailed, err
			}
			continue
		}

		u.failAttempt(ctx, attempt, "EXHAUSTED", err.Error())
		return domain.AttemptStatusFailed, err
	}
	
	return domain.AttemptStatusFailed, nil
}
