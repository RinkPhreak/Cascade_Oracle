package http

import (
	"cascade/internal/application/usecase"
	"cascade/internal/delivery/http/dto"
	"github.com/gofiber/fiber/v2"
	"io"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
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

// Register godoc
// @Summary Register a new TG account
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body dto.RegisterAccountRequest true "Phone"
// @Success 201 {object} dto.AccountResponse
// @Router /api/v1/accounts/register [post]
// @Security Bearer
func (h *AccountHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid schema"})
	}

	acc, err := h.accountUC.RegisterAccount(c.Context(), req.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.AccountResponse{
		ID:        acc.ID.String(),
		Phone:     acc.Phone,
		Status:    string(acc.Status),
		CreatedAt: acc.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (h *AccountHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid uuid"})
	}

	if err := h.accountUC.DeleteAccount(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AccountHandler) Import(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		slog.Error("failed to parse multipart form", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "failed to parse multipart form"})
	}

	filesMap := make(map[string][]byte)
	// We specifically look for the "file" key as per contract
	formFiles := form.File["file"]
	if len(formFiles) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "missing 'file' field"})
	}

	for _, file := range formFiles {
		f, err := file.Open()
		if err != nil {
			continue
		}
		data, _ := io.ReadAll(f)
		filesMap[file.Filename] = data
		f.Close()
	}

	proxyPort, _ := strconv.Atoi(c.FormValue("proxy_port"))
	proxyReq := dto.CreateProxyRequest{
		Host:     c.FormValue("proxy_host"),
		Port:     proxyPort,
		Username: c.FormValue("proxy_username"),
		Password: c.FormValue("proxy_password"),
	}

	comment := c.FormValue("comment")

	acc, err := h.accountUC.ImportAccount(c.Context(), filesMap, proxyReq, comment)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Code:    "IMPORT_FAILED",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.AccountResponse{
		ID:        acc.ID.String(),
		Phone:     acc.Phone,
		Status:    string(acc.Status),
		CreatedAt: acc.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}
