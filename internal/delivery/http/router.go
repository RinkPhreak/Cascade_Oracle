//go:generate go run github.com/swaggo/swag/cmd/swag@latest init -g router.go -d . --parseDependency -o ../../../api -ot yaml
package http

import (
	"crypto/rsa"

	"github.com/gofiber/fiber/v2"
	"cascade/internal/delivery/http/middleware"
)

func SetupRoutes(app *fiber.App, pubKey *rsa.PublicKey, authH *AuthHandler, campH *CampaignHandler, sysH *SystemHandler, conH *ContactHandler) {
	api := app.Group("/api/v1")

	// Public Routes
	auth := api.Group("/auth")
	auth.Post("/login", authH.Login)

	// Protected Routes
	protected := api.Group("/", middleware.RequireAuth(pubKey))

	sys := protected.Group("/system")
	sys.Post("/halt", sysH.HaltSystem)
	sys.Post("/resume", sysH.ResumeSystem)

	camps := protected.Group("/campaigns")
	camps.Post("/", campH.Create)
	camps.Post("/:id/import", campH.ImportCSV)
	camps.Post("/:id/start", campH.Start)
	camps.Post("/:id/pause", campH.Pause)

	contacts := protected.Group("/contacts")
	contacts.Post("/:id/anonymise", conH.Anonymise)
}
