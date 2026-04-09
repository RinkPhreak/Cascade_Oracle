package http

import (
	"github.com/gofiber/fiber/v2"

	"cascade/internal/application/usecase"
	"cascade/internal/delivery/http/dto"
)

type AuthHandler struct {
	uc *usecase.AuthUseCase
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// Login godoc
// @Summary Operator Login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.AuthLoginRequest true "Credentials"
// @Success 200 {object} dto.AuthTokenResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.AuthLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid json"})
	}

	token, err := h.uc.Login(c.Context(), req.Login, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Code: "UNAUTHORIZED", Message: err.Error()})
	}

	return c.JSON(dto.AuthTokenResponse{AccessToken: token})
}
