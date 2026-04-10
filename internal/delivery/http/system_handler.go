package http

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"cascade/internal/application/port"
	"cascade/internal/application/usecase"
	"cascade/internal/delivery/http/dto"
)

type SystemHandler struct {
	authUC *usecase.AuthUseCase
	cache  port.Cache
}

func NewSystemHandler(authUC *usecase.AuthUseCase, cache port.Cache) *SystemHandler {
	return &SystemHandler{authUC: authUC, cache: cache}
}

// HaltSystem godoc
// @Summary Emergency Break-Glass Halt
// @Tags system
// @Accept json
// @Produce json
// @Param request body dto.BreakGlassRequest true "Reason and Re-auth"
// @Success 200 "System Halted"
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/system/halt [post]
// @Security Bearer
func (h *SystemHandler) HaltSystem(c *fiber.Ctx) error {
	var req dto.BreakGlassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid schema"})
	}

	if err := h.authUC.VerifyPassword(req.Password); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Code: "FORBIDDEN", Message: "invalid re-authentication password"})
	}

	// Идемпотентность: проверяем, не остановлена ли уже система
	existingReason, err := h.cache.Get(c.Context(), "cascade:system:halted")
	if err == nil && existingReason != "" {
		return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Code: "CONFLICT", Message: "System is already halted"})
	}

	h.cache.Set(c.Context(), "cascade:system:halted", req.Reason, 8760*time.Hour)
	return c.SendStatus(fiber.StatusOK)
}

// ResumeSystem godoc
// @Summary Emergency Break-Glass Resume
// @Tags system
// @Accept json
// @Produce json
// @Param request body dto.BreakGlassRequest true "Re-auth required"
// @Success 200 "System Resumed"
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/system/resume [post]
// @Security Bearer
func (h *SystemHandler) ResumeSystem(c *fiber.Ctx) error {
	var req dto.BreakGlassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid schema"})
	}

	if err := h.authUC.VerifyPassword(req.Password); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Code: "FORBIDDEN", Message: "invalid re-authentication password"})
	}

	// Идемпотентность: проверяем, остановлена ли система вообще
	existingReason, err := h.cache.Get(c.Context(), "cascade:system:halted")
	if err != nil || existingReason == "" {
		return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Code: "CONFLICT", Message: "System is already operational"})
	}

	h.cache.Del(c.Context(), "cascade:system:halted")
	return c.SendStatus(fiber.StatusOK)
}

// GetMetrics godoc
// @Summary Fetch system health and halt state
// @Tags system
// @Success 200 {object} dto.SystemMetricsResponse
// @Router /api/v1/system/metrics [get]
// @Security Bearer
func (h *SystemHandler) GetMetrics(c *fiber.Ctx) error {
	reason, err := h.cache.Get(c.Context(), "cascade:system:halted")
	isHalted := err == nil && reason != ""

	status := "OPERATIONAL"
	if isHalted {
		status = "HALTED"
	}

	// Возвращаем точную структуру, которую ждет фронтенд
	return c.JSON(dto.SystemMetricsResponse{
		CascadeMemoryUsageRatio: 0.1, // Заглушка до подключения memory_monitor
		ActiveTgAccounts:        0,   // Заглушка до интеграции счетчика БД
		TotalTgAccounts:         0,   // Заглушка до интеграции счетчика БД
		QueueDepth:              0,   // Заглушка до Asynq Inspector
		SystemStatus:            status,
	})
}
