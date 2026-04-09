package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"cascade/internal/application/usecase"
	"cascade/internal/delivery/http/dto"
	"cascade/internal/domain"
)

type CampaignHandler struct {
	campaignUC *usecase.CampaignUseCase
	authUC     *usecase.AuthUseCase
}

func NewCampaignHandler(uc *usecase.CampaignUseCase, authUC *usecase.AuthUseCase) *CampaignHandler {
	return &CampaignHandler{campaignUC: uc, authUC: authUC}
}

func (h *CampaignHandler) MountRoutes(router fiber.Router) {
	group := router.Group("/api/v1/campaigns")
	group.Post("/", h.Create)
	group.Post("/:id/import", h.ImportCSV)
	group.Post("/:id/start", h.Start)
}

// Create godoc
// @Summary Create a new campaign
// @Tags campaigns
// @Accept json
// @Produce json
// @Param request body dto.CreateCampaignRequest true "Campaign params"
// @Success 201 {object} dto.CampaignResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/campaigns [post]
func (h *CampaignHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateCampaignRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid json"})
	}

	// Skipping explicit go-validator setup for brevity, assuming standard structurally valid JSON
	if req.Name == "" || len(req.Templates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "VALIDATION_ERROR", Message: "missing required fields"})
	}

	camp, err := h.campaignUC.CreateCampaign(c.Context(), req.Name, req.ScheduledAt, req.Templates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL", Message: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.CampaignResponse{
		ID:          camp.ID.String(),
		Name:        camp.Name,
		Status:      string(camp.Status),
		ScheduledAt: camp.ScheduledAt,
		CreatedAt:   camp.CreatedAt,
	})
}

// ImportCSV godoc
// @Summary Import contacts via CSV
// @Tags campaigns
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Campaign UUID"
// @Param file formData file true "CSV File"
// @Success 200 {object} map[string]int
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/campaigns/{id}/import [post]
func (h *CampaignHandler) ImportCSV(c *fiber.Ctx) error {
	idStr := c.Params("id")
	campID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "INVALID_ID", Message: "invalid uuid"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "NO_FILE", Message: "multipart file required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "FILE_ERROR", Message: "cannot open file"})
	}
	defer file.Close()

	count, err := h.campaignUC.ImportCSV(c.Context(), campID, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "IMPORT_FAILED", Message: err.Error()})
	}

	return c.JSON(fiber.Map{"imported_count": count})
}

// Start godoc
// @Summary Enqueue campaign for processing
// @Tags campaigns
// @Param id path string true "Campaign UUID"
// @Success 202 "Accepted for processing"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /api/v1/campaigns/{id}/start [post]
func (h *CampaignHandler) Start(c *fiber.Ctx) error {
	idStr := c.Params("id")
	campID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "INVALID_ID", Message: "invalid uuid"})
	}

	if err := h.campaignUC.LaunchCampaign(c.Context(), campID); err != nil {
		if errors.Is(err, domain.ErrSystemSuspended) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(dto.ErrorResponse{
				Code:    "SUSPENDED",
				Message: "system is currently suspended",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL", Message: err.Error()})
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// Pause godoc
// @Summary Emergency Campaign Pause
// @Tags campaigns
// @Accept json
// @Produce json
// @Param id path string true "Campaign UUID"
// @Param request body dto.BreakGlassRequest true "Re-auth required"
// @Success 200 "Campaign Paused"
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/campaigns/{id}/pause [post]
// @Security Bearer
func (h *CampaignHandler) Pause(c *fiber.Ctx) error {
	var req dto.BreakGlassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid schema"})
	}

	if err := h.authUC.VerifyPassword(req.Password); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Code: "FORBIDDEN", Message: "invalid re-authentication password"})
	}

	idStr := c.Params("id")
	campID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "INVALID_ID", Message: "invalid uuid"})
	}

	if err := h.campaignUC.PauseCampaign(c.Context(), campID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
	}
	
	return c.SendStatus(fiber.StatusOK)
}
