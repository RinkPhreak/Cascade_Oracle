package http

import (
	"github.com/gofiber/fiber/v2"
	"cascade/internal/application/usecase"
	"cascade/internal/delivery/http/dto"
)

type AccountHandler struct {
	accountUC *usecase.AccountUseCase
}

func NewAccountHandler(auc *usecase.AccountUseCase) *AccountHandler {
	return &AccountHandler{accountUC: auc}
}

// List godoc
// @Summary List all Telegram accounts
// @Tags accounts
// @Success 200 {array} dto.AccountResponse
// @Router /api/v1/accounts [get]
// @Security Bearer
func (h *AccountHandler) List(c *fiber.Ctx) error {
	accounts, err := h.accountUC.ListAccounts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	var res []dto.AccountResponse
	for _, a := range accounts {
		res = append(res, dto.AccountResponse{
			ID:              a.ID.String(),
			Phone:           a.Phone, // Already encrypted, handled by UC if needed otherwise visible as cipher
			Channel:         a.Channel,
			ProxyID:         a.ProxyID.String(),
			Status:          string(a.Status),
			DailyCheckCount: a.DailyCheckCount,
			DailySendCount:  a.DailySendCount,
			CreatedAt:       a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(res)
}
