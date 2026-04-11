package http

import (
	"cascade/internal/application/usecase"
	"cascade/internal/delivery/http/dto"
	"github.com/gofiber/fiber/v2"
)

type ProxyHandler struct {
	accountUC *usecase.AccountUseCase
}

func NewProxyHandler(auc *usecase.AccountUseCase) *ProxyHandler {
	return &ProxyHandler{accountUC: auc}
}

// List godoc
// @Summary List all network proxies
// @Tags proxies
// @Success 200 {array} dto.ProxyResponse
// @Router /api/v1/proxies [get]
// @Security Bearer
func (h *ProxyHandler) List(c *fiber.Ctx) error {
	proxies, err := h.accountUC.ListProxies(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	var res []dto.ProxyResponse
	for _, p := range proxies {
		res = append(res, dto.ProxyResponse{
			ID:        p.ID.String(),
			Host:      p.Host,
			Port:      p.Port,
			Username:  p.Username,
			Status:    string(p.Status),
			CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(res)
}

// Create godoc
// @Summary Add a new proxy
// @Tags proxies
// @Accept json
// @Produce json
// @Param request body dto.CreateProxyRequest true "Proxy Address"
// @Success 201 {object} dto.ProxyResponse
// @Router /api/v1/proxies [post]
// @Security Bearer
func (h *ProxyHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateProxyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid schema"})
	}

	proxy, err := h.accountUC.AddProxy(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.ProxyResponse{
		ID:        proxy.ID.String(),
		Host:      proxy.Host,
		Port:      proxy.Port,
		Username:  proxy.Username,
		Status:    string(proxy.Status),
		CreatedAt: proxy.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}
