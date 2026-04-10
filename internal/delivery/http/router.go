//go:generate go run github.com/swaggo/swag/cmd/swag@latest init -g router.go -d . --parseDependency -o ../../../api -ot yaml
package http

import (
	"crypto/rsa"

	"github.com/gofiber/fiber/v2"
	"cascade/internal/delivery/http/middleware"
)

func SetupRoutes(app *fiber.App, pubKey *rsa.PublicKey, authH *AuthHandler, campH *CampaignHandler, sysH *SystemHandler, conH *ContactHandler, accH *AccountHandler, proxyH *ProxyHandler) {
	api := app.Group("/api/v1")

	// Public Routes
	auth := api.Group("/auth")
	auth.Post("/login", authH.Login)

	// Protected Routes
	protected := api.Group("/", middleware.RequireAuth(pubKey))

	// System metrics & control
	sys := protected.Group("/system")
	sys.Post("/halt", sysH.HaltSystem)
	sys.Post("/resume", sysH.ResumeSystem)
	sys.Get("/metrics", sysH.GetMetrics)

	// Accounts & Pool
	accounts := protected.Group("/accounts")
	accounts.Get("/", accH.List)
	accounts.Post("/register", accH.Register)

	proxies := protected.Group("/proxies")
	proxies.Get("/", proxyH.List)
	proxies.Post("/", proxyH.Create)

	// Campaigns Monitoring
	camps := protected.Group("/campaigns")
	camps.Get("/", campH.List)
	camps.Post("/", campH.Create)
	camps.Post("/:id/import", campH.ImportCSV)
	camps.Post("/:id/start", campH.Start)
	camps.Post("/:id/pause", campH.Pause)
	camps.Get("/:id/stats", campH.GetStats)
	camps.Get("/:id/tasks", campH.GetTasks)

	// Leads & Trace
	contacts := protected.Group("/contacts")
	contacts.Get("/", conH.List)
	contacts.Get("/:id/trace", conH.GetTrace)
	contacts.Post("/:id/anonymise", conH.Anonymise)
}
