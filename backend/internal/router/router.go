package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"github.com/ybapat/screener/backend/internal/handler"
	"github.com/ybapat/screener/backend/internal/middleware"
	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/internal/service"
)

type Handlers struct {
	Auth        *handler.AuthHandler
	Ingestion   *handler.IngestionHandler
	User        *handler.UserHandler
	Dataset     *handler.DatasetHandler
	Marketplace *handler.MarketplaceHandler
	Dashboard   *handler.DashboardHandler
}

func New(h Handlers, authService *service.AuthService, rdb *redis.Client) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS())
	r.Use(chimiddleware.Timeout(30 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public auth routes
	r.Post("/auth/register", h.Auth.Register)
	r.Post("/auth/login", h.Auth.Login)
	r.Post("/auth/refresh", h.Auth.Refresh)

	// Public marketplace browsing
	r.Get("/api/v1/marketplace/datasets", h.Dataset.ListPublic)
	r.Get("/api/v1/marketplace/datasets/{id}", h.Dataset.GetPublic)
	r.Get("/api/v1/marketplace/datasets/{id}/samples", h.Dataset.GetSamples)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService))

		// Common (any authenticated user)
		r.Get("/api/v1/users/me", h.User.GetProfile)
		r.Patch("/api/v1/users/me", h.User.UpdateProfile)
		r.Post("/auth/logout", h.Auth.Logout)
		r.Get("/api/v1/credits/history", h.Dashboard.CreditHistory)

		// Seller routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(models.RoleSeller, models.RoleAdmin))
			r.Use(middleware.RateLimit(rdb, 10, time.Minute))

			r.Post("/api/v1/data/upload", h.Ingestion.Upload)
			r.Get("/api/v1/data/batches", h.Ingestion.ListBatches)
			r.Get("/api/v1/data/batches/{id}", h.Ingestion.GetBatch)
			r.Delete("/api/v1/data/batches/{id}", h.Ingestion.WithdrawBatch)
			r.Get("/api/v1/privacy/budget", h.User.GetPrivacyBudget)
			r.Get("/api/v1/privacy/ledger", h.User.GetEpsilonLedger)
			r.Get("/api/v1/dashboard/seller", h.Dashboard.SellerOverview)
		})

		// Buyer routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(models.RoleBuyer, models.RoleAdmin))

			r.Post("/api/v1/marketplace/datasets/{id}/purchase", h.Marketplace.Purchase)
			r.Get("/api/v1/buyer/purchases", h.Marketplace.GetPurchases)
			r.Post("/api/v1/marketplace/segments", h.Marketplace.CreateSegment)
			r.Get("/api/v1/marketplace/segments", h.Marketplace.ListSegments)
			r.Post("/api/v1/marketplace/segments/{id}/bids", h.Marketplace.PlaceBid)
			r.Get("/api/v1/marketplace/bids", h.Marketplace.ListBids)
			r.Delete("/api/v1/marketplace/bids/{id}", h.Marketplace.CancelBid)
			r.Post("/api/v1/credits/topup", h.Marketplace.TopupCredits)
			r.Get("/api/v1/dashboard/buyer", h.Dashboard.BuyerOverview)
		})

		// Admin routes: assemble datasets
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(models.RoleAdmin))
			r.Post("/api/v1/admin/datasets/assemble", h.Dataset.Assemble)
		})
	})

	return r
}
