package main

import (
	"context"
	"encoding/hex"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/golang-migrate/migrate/v4"
	pgDriver "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	asynqAdapter "cascade/internal/adapters/asynq"
	"cascade/internal/adapters/cache"
	"cascade/internal/adapters/crypto"
	"cascade/internal/adapters/db"
	"cascade/internal/adapters/messenger"
	"cascade/internal/application/usecase"
	deliveryhttp "cascade/internal/delivery/http"
	"cascade/internal/delivery/worker"
	"cascade/internal/infrastructure/observability"
)

func main() {
	// Configure logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 1. App Configuration (Envs)
	dsn := os.Getenv("POSTGRES_DSN")
	redisAddr := os.Getenv("REDIS_ADDR")
	appIDStr := os.Getenv("TG_APP_ID")
	appHash := os.Getenv("TG_APP_HASH")
	secretKey := os.Getenv("APP_ENCRYPTION_KEY")
	adminLogin := os.Getenv("ADMIN_LOGIN")
	adminPasswdHash := os.Getenv("ADMIN_PASSWORD_HASH")
	jwtPubKeyStr := os.Getenv("JWT_PUBLIC_KEY")
	jwtPrivKeyStr := os.Getenv("JWT_PRIVATE_KEY")

	if len(secretKey) != 32 {
		log.Fatalf("Fatal: APP_ENCRYPTION_KEY must be exactly 32 bytes (got %d)", len(secretKey))
	}

	appID, _ := strconv.Atoi(appIDStr)

	// Clean up keys from .env if they contain literal string quotes or encoded newlines
	cleanKey := func(k string) string {
		k = strings.Trim(k, `"`)
		k = strings.Trim(k, `'`)
		return strings.ReplaceAll(k, "\\n", "\n")
	}

	// Parse keys using jwt
	jwtPrivKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cleanKey(jwtPrivKeyStr)))
	if err != nil {
		log.Fatalf("Fatal: could not parse JWT_PRIVATE_KEY: %v", err)
	}
	jwtPubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cleanKey(jwtPubKeyStr)))
	if err != nil {
		log.Fatalf("Fatal: could not parse JWT_PUBLIC_KEY: %v", err)
	}

	// 2. Database Connection
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Run Database Migrations
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm: %v", err)
	}

	driver, err := pgDriver.WithInstance(sqlDB, &pgDriver.Config{})
	if err != nil {
		log.Fatalf("failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file:///app/migrations", "postgres", driver)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}
	slog.Info("Database migrations applied successfully")

	// 3. Application Context & Orchestration
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	// 4. Repositories
	accountRepo := db.NewAccountRepository(gormDB)
	campaignRepo := db.NewCampaignRepository(gormDB)
	contactRepo := db.NewContactRepository(gormDB)
	attemptRepo := db.NewAttemptRepository(gormDB)
	proxyRepo := db.NewProxyRepository(gormDB)

	// 5. Infrastructure Adapters
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	redisCache := cache.NewRedisCache(redisClient)

	hexKey := hex.EncodeToString([]byte(secretKey))
	cryptoSvc, err := crypto.NewCryptoService(hexKey, "default_pepper", nil)
	if err != nil {
		log.Fatalf("failed to init crypto service: %v", err)
	}

	enqueuer := asynqAdapter.NewAsynqEnqueuer(redisAddr)
	tgPool := messenger.NewTelegramClientPool(accountRepo, proxyRepo, appID, appHash)

	uow := db.NewUnitOfWork(gormDB)

	// 6. Application UseCases
	authUC := usecase.NewAuthUseCase(adminLogin, adminPasswdHash, jwtPrivKey)
	contactUC := usecase.NewContactUseCase(contactRepo, attemptRepo, uow, cryptoSvc)
	campUC := usecase.NewCampaignUseCase(campaignRepo, contactRepo, attemptRepo, enqueuer, uow, cryptoSvc)
	accountUC := usecase.NewAccountUseCase(accountRepo, proxyRepo, tgPool)
	waterfallUC := usecase.NewWaterfallUseCase(
		campaignRepo, contactRepo, accountRepo, proxyRepo,
		attemptRepo, tgPool, nil, // Replaced explicit SMS with nil
		redisCache, enqueuer, cryptoSvc,
	)

	// 7. Delivery / Transport
	fiberApp := fiber.New()

	authHandler := deliveryhttp.NewAuthHandler(authUC)
	campaignHandler := deliveryhttp.NewCampaignHandler(campUC, authUC)
	systemHandler := deliveryhttp.NewSystemHandler(authUC, redisCache)
	contactHandler := deliveryhttp.NewContactHandler(contactUC)
	accountHandler := deliveryhttp.NewAccountHandler(accountUC)
	proxyHandler := deliveryhttp.NewProxyHandler(accountUC)

	deliveryhttp.SetupRoutes(fiberApp, jwtPubKey, authHandler, campaignHandler, systemHandler, contactHandler, accountHandler, proxyHandler)

	asynqMux := worker.NewServerMux(waterfallUC)
	asynqSrv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 50,
			Queues: map[string]int{
				"cascade:send":     5, // Future explicit high-priority setup
				"cascade:precheck": 2, // Future explicit low-priority setup
				"default":          3,
			},
		},
	)

	memMonitor := observability.NewMemoryMonitor(redisCache)

	// 8. Graceful Shutdown Signal Interception
	g.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-sigChan:
			slog.Info("received shutdown signal", "signal", sig)
			cancel() // Initiate context shutdown
		case <-gCtx.Done():
		}
		return nil
	})

	// 9. Start Services
	g.Go(func() error {
		slog.Info("Starting Fiber Server on :3000")
		if err := fiberApp.Listen(":3000"); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		slog.Info("Gracefully shutting down Fiber")
		return fiberApp.Shutdown()
	})

	g.Go(func() error {
		slog.Info("Starting Asynq Server")
		if err := asynqSrv.Run(asynqMux); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		slog.Info("Gracefully shutting down Asynq multiplexer")
		asynqSrv.Shutdown()
		return nil
	})

	g.Go(func() error {
		slog.Info("Starting Memory Monitor cgroup guard")
		memMonitor.Start(gCtx, 90.0) // 90% threshold
		return nil
	})

	// Wait for termination
	if err := g.Wait(); err != nil {
		slog.Error("server exited with error", "error", err)
	}
	slog.Info("shutdown complete")
}
