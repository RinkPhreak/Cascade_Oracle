package http

import (
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
