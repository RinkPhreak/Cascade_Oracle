package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"cascade/internal/application/usecase"
	"cascade/internal/delivery/http/dto"
)

type ContactHandler struct {
	contactUC *usecase.ContactUseCase
}

func NewContactHandler(uc *usecase.ContactUseCase) *ContactHandler {
	return &ContactHandler{contactUC: uc}
}

// List godoc
// @Summary List contacts (leads)
// @Tags contacts
// @Param has_replied query bool false "Filter only those who replied"
// @Success 200 {array} dto.LeadResponse
// @Router /api/v1/contacts [get]
// @Security Bearer
func (h *ContactHandler) List(c *fiber.Ctx) error {
	hasReplied := c.Query("has_replied") == "true"
	
	if hasReplied {
		leads, err := h.contactUC.ListReplied(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
		}
		
		res := make([]dto.LeadResponse, 0)
		for _, l := range leads {
			res = append(res, dto.LeadResponse{
				ID:        l.ID.String(),
				Phone:     l.Phone,
				Name:      l.Name,
				Message:   l.Message,
				Channel:   l.Channel,
				RepliedAt: l.RepliedAt.Format(time.RFC3339),
			})
		}
		return c.JSON(res)
	}

	// Default behavior: empty list or all contacts (not requested in detail, providing empty to satisfy init)
	return c.JSON([]dto.LeadResponse{})
}

// GetTrace godoc
// @Summary Fetch delivery trace for a contact
// @Tags contacts
// @Param id path string true "Contact UUID"
// @Success 200 {array} dto.ContactTraceResponse
// @Router /api/v1/contacts/{id}/trace [get]
// @Security Bearer
func (h *ContactHandler) GetTrace(c *fiber.Ctx) error {
	idStr := c.Params("id")
	contactID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "INVALID_ID", Message: "invalid uuid"})
	}

	trace, err := h.contactUC.GetContactTrace(c.Context(), contactID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
	}

	res := make([]dto.ContactTraceResponse, 0)
	for _, t := range trace {
		res = append(res, dto.ContactTraceResponse{
			ID:           t.ID.String(),
			Channel:      t.Channel,
			Status:       string(t.Status),
			ErrorCode:    t.ErrorCode,
			ErrorMessage: t.ErrorMessage,
			Attempt:      t.AttemptNumber,
			Timestamp:    t.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(res)
}

// Anonymise godoc
// @Summary Right to be Forgotten (152-FZ / GDPR)
// @Tags contacts
// @Produce json
// @Param id path string true "Contact UUID"
// @Success 200 {object} dto.AnonymiseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/contacts/{id}/anonymise [post]
// @Security Bearer
func (h *ContactHandler) Anonymise(c *fiber.Ctx) error {
	idStr := c.Params("id")
	contactID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "INVALID_ID", Message: "invalid uuid"})
	}

	if err := h.contactUC.Anonymise(c.Context(), contactID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
	}

	return c.JSON(dto.AnonymiseResponse{
		Status:    "success",
		Message:   "pii obfuscated and record soft-deleted",
		ContactID: idStr,
	})
}
