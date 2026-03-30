package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ybapat/screener/backend/internal/config"
	"github.com/ybapat/screener/backend/internal/db"
	"github.com/ybapat/screener/backend/internal/handler"
	"github.com/ybapat/screener/backend/internal/privacy"
	"github.com/ybapat/screener/backend/internal/repository"
	"github.com/ybapat/screener/backend/internal/router"
	"github.com/ybapat/screener/backend/internal/service"
)

func main() {
	cfg := config.Load()

	// Logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Database
	pool, err := db.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	// Redis
	rdb, err := db.NewRedisClient(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	slog.Info("connected to redis")

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(pool)
	screenTimeRepo := repository.NewScreenTimeRepository(pool)
	datasetRepo := repository.NewDatasetRepository(pool)
	purchaseRepo := repository.NewPurchaseRepository(pool)
	marketplaceRepo := repository.NewMarketplaceRepository(pool)

	// Privacy
	budgetTracker := privacy.NewBudgetTracker(pool)

	// Services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg.JWTSecret)
	ingestionService := service.NewIngestionService(screenTimeRepo)
	creditService := service.NewCreditService(userRepo, marketplaceRepo)
	anonymizationService := service.NewAnonymizationService(screenTimeRepo, userRepo, datasetRepo, budgetTracker)
	marketplaceService := service.NewMarketplaceService(datasetRepo, purchaseRepo, marketplaceRepo, creditService)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	ingestionHandler := handler.NewIngestionHandler(ingestionService)
	userHandler := handler.NewUserHandler(userRepo, budgetTracker)
	datasetHandler := handler.NewDatasetHandler(marketplaceService, anonymizationService)
	marketplaceHandler := handler.NewMarketplaceHandler(marketplaceService, creditService)
	dashboardHandler := handler.NewDashboardHandler(userRepo, creditService, ingestionService, marketplaceService, budgetTracker)

	// Router
	r := router.New(router.Handlers{
		Auth:        authHandler,
		Ingestion:   ingestionHandler,
		User:        userHandler,
		Dataset:     datasetHandler,
		Marketplace: marketplaceHandler,
		Dashboard:   dashboardHandler,
	}, authService, rdb)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("starting server", "addr", addr, "env", cfg.Env)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
