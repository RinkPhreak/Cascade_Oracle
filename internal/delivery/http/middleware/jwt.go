package middleware

import (
	"crypto/rsa"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"cascade/internal/delivery/http/dto"
)

func RequireAuth(pubKey *rsa.PublicKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "missing or invalid authorization header",
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return pubKey, nil
		}, jwt.WithLeeway(5*time.Minute))

		if err != nil || !token.Valid {
			log.Printf("🔥 JWT PARSE ERROR: %v\n", err)

			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Code:    "UNAUTHORIZED",
				Message: "token expired or invalid",
			})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("user_id", claims["sub"])
		}

		return c.Next()
	}
}
