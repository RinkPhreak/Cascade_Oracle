package usecase

import (
	"context"
	"encoding/csv"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type CampaignUseCase struct {
	campaignRepo port.CampaignRepository
	contactRepo  port.ContactRepository
	attemptRepo  port.AttemptRepository
	enqueuer     port.TaskEnqueuer
	uow          port.UnitOfWork
	crypto       port.CryptoService
}

func NewCampaignUseCase(cr port.CampaignRepository, contactRepo port.ContactRepository, ar port.AttemptRepository, eq port.TaskEnqueuer, uow port.UnitOfWork, crypto port.CryptoService) *CampaignUseCase {
	return &CampaignUseCase{
		campaignRepo: cr,
		contactRepo:  contactRepo,
		attemptRepo:  ar,
		enqueuer:     eq,
		uow:          uow,
		crypto:       crypto,
	}
}

func (u *CampaignUseCase) CreateCampaign(ctx context.Context, name string, scheduledAt *time.Time, templates map[string]string) (*domain.Campaign, error) {
	campID, _ := uuid.NewV7()
	campaign := &domain.Campaign{
		ID:          campID,
		Name:        name,
		Status:      domain.CampaignStatusDraft,
		ScheduledAt: scheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := u.uow.Execute(ctx, func(txCtx context.Context) error {
		if exErr := u.campaignRepo.Create(txCtx, campaign); exErr != nil {
			return exErr
		}
		for channel, content := range templates {
			tmplID, _ := uuid.NewV7()
			tmpl := &domain.MessageTemplate{
				ID:         tmplID,
				CampaignID: campaign.ID,
				Channel:    channel,
				Content:    content,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err := u.campaignRepo.CreateTemplate(txCtx, tmpl); err != nil {
				return err
			}
		}
		return nil
	})

	return campaign, err
}

func (u *CampaignUseCase) ImportCSV(ctx context.Context, campaignID uuid.UUID, reader io.Reader) (int, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return 0, err
	}

	var parsedContacts []*domain.Contact
	for idx, row := range records {
		if idx == 0 {
			continue // Skip headers
		}
		if len(row) < 2 {
			continue
		}
		phoneRaw := strings.TrimSpace(row[0])
		nameRaw := strings.TrimSpace(row[1])

		if phoneRaw == "" {
			continue
		}

		phoneEnc, _ := u.crypto.Encrypt(phoneRaw)
		nameEnc, _ := u.crypto.Encrypt(nameRaw)
		hash := u.crypto.HashPhone(phoneRaw)

		cID, _ := uuid.NewV7()
		contact := &domain.Contact{
			ID:        cID,
			PhoneHash: hash,
			Phone:     phoneEnc,
			Name:      nameEnc,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		parsedContacts = append(parsedContacts, contact)
	}

	batchSize := 500
	for i := 0; i < len(parsedContacts); i += batchSize {
		end := i + batchSize
		if end > len(parsedContacts) {
			end = len(parsedContacts)
		}
		if err := u.contactRepo.CreateBatch(ctx, parsedContacts[i:end]); err != nil {
			return 0, err
		}
	}

	// Assign contacts to the campaign in batch
	var cIDs []uuid.UUID
	for _, c := range parsedContacts {
		cIDs = append(cIDs, c.ID)
	}
	if err := u.campaignRepo.AssignContactsToCampaign(ctx, campaignID, cIDs, domain.CampaignContactPending); err != nil {
		return 0, err
	}

	return len(parsedContacts), nil
}

func (u *CampaignUseCase) LaunchCampaign(ctx context.Context, campaignID uuid.UUID) error {
	campaign, err := u.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return err
	}
	if campaign.Status == domain.CampaignStatusActive {
		return nil
	}
	if !campaign.IsExecutable(time.Now()) {
		return domain.ErrSystemSuspended
	}

	campaign.TransitionStatus(domain.CampaignStatusActive)
	if err := u.campaignRepo.Update(ctx, campaign); err != nil {
		return err
	}

	contacts, err := u.campaignRepo.ListCampaignContacts(ctx, campaign.ID)
	if err != nil {
		return err
	}

	for _, cc := range contacts {
		if cc.Status != domain.CampaignContactPending {
			continue
		}
		payload := domain.WaterfallPayload{
			CampaignID: campaign.ID,
			ContactID:  cc.ContactID,
		}

		// Enqueue logic with schedule binding
		if err := u.enqueuer.EnqueueWaterfall(ctx, payload, campaign.ScheduledAt); err != nil {
			return err
		}
	}
	return nil
}

func (u *CampaignUseCase) PauseCampaign(ctx context.Context, campaignID uuid.UUID) error {
	campaign, err := u.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return err
	}

	campaign.TransitionStatus(domain.CampaignStatusPaused)
	if err := u.campaignRepo.Update(ctx, campaign); err != nil {
		return err
	}

	return u.enqueuer.CancelCampaignTasks(ctx, campaign.ID)
}

func (u *CampaignUseCase) ListCampaigns(ctx context.Context) ([]*domain.Campaign, error) {
	return u.campaignRepo.ListCampaigns(ctx)
}

func (u *CampaignUseCase) DeleteCampaign(ctx context.Context, id uuid.UUID) error {
	// Cancel any active tasks first
	_ = u.enqueuer.CancelCampaignTasks(ctx, id)
	return u.campaignRepo.Delete(ctx, id)
}

func (u *CampaignUseCase) GetCampaignStats(ctx context.Context, campaignID uuid.UUID, start, end *time.Time) (*domain.CampaignStats, error) {
	return u.campaignRepo.GetStats(ctx, campaignID, start, end)
}

func (u *CampaignUseCase) GetCampaignTasks(ctx context.Context, campaignID uuid.UUID) ([]*domain.SendAttempt, error) {
	return u.attemptRepo.GetStuckAttempts(ctx, campaignID)
}
