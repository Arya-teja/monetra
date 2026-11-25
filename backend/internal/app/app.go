// Package app is the composition root: it wires repositories, services, handlers,
// routes, and middleware into one http.Handler. main.go and the integration tests
// both use it, so the tests run exactly the application that is deployed.
package app

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Arya-teja/monetra/backend/internal/http/handler"
	"github.com/Arya-teja/monetra/backend/internal/http/middleware"
	"github.com/Arya-teja/monetra/backend/internal/repository/postgres"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

// accessTokenTTL matches the NextAuth session maxAge on the frontend.
const accessTokenTTL = 7 * 24 * time.Hour

type Config struct {
	JWTSecret  string
	CORSOrigin string
	// Now returns the current time; the dashboard uses it to find the current month.
	Now func() time.Time
}

func NewHandler(pool *pgxpool.Pool, cfg Config) http.Handler {
	tokens := service.NewTokenManager(cfg.JWTSecret, accessTokenTTL)

	accountRepo := postgres.NewAccountRepository(pool)
	categoryRepo := postgres.NewCategoryRepository(pool)
	transactionRepo := postgres.NewTransactionRepository(pool)
	budgetRepo := postgres.NewBudgetRepository(pool)

	authService := service.NewAuthService(postgres.NewUserRepository(pool), tokens)
	accountService := service.NewAccountService(accountRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	transactionService := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo)
	budgetService := service.NewBudgetService(budgetRepo, categoryRepo)
	dashboardService := service.NewDashboardService(postgres.NewDashboardRepository(pool), transactionRepo, budgetRepo, cfg.Now)

	health := handler.NewHealthHandler(pool)
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	requireAuth := middleware.Auth(authService)

	mux := http.NewServeMux()
	protected := func(pattern string, h http.HandlerFunc) {
		mux.Handle(pattern, requireAuth(h))
	}

	mux.HandleFunc("GET /api/v1/health", health.Check)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	protected("GET /api/v1/auth/me", authHandler.Me)
	protected("POST /api/v1/auth/logout", authHandler.Logout)

	protected("GET /api/v1/accounts", accountHandler.List)
	protected("POST /api/v1/accounts", accountHandler.Create)
	protected("PATCH /api/v1/accounts/{id}", accountHandler.Update)
	protected("DELETE /api/v1/accounts/{id}", accountHandler.Delete)

	protected("GET /api/v1/categories", categoryHandler.List)
	protected("POST /api/v1/categories", categoryHandler.Create)
	protected("PATCH /api/v1/categories/{id}", categoryHandler.Rename)
	protected("DELETE /api/v1/categories/{id}", categoryHandler.Delete)

	protected("GET /api/v1/transactions", transactionHandler.List)
	protected("POST /api/v1/transactions", transactionHandler.Create)
	protected("GET /api/v1/transactions/{id}", transactionHandler.Get)
	protected("PATCH /api/v1/transactions/{id}", transactionHandler.Update)
	protected("DELETE /api/v1/transactions/{id}", transactionHandler.Delete)

	protected("GET /api/v1/budgets", budgetHandler.List)
	protected("POST /api/v1/budgets", budgetHandler.Create)
	protected("POST /api/v1/budgets/copy", budgetHandler.Copy)
	protected("PATCH /api/v1/budgets/{id}", budgetHandler.Update)
	protected("DELETE /api/v1/budgets/{id}", budgetHandler.Delete)

	protected("GET /api/v1/dashboard/summary", dashboardHandler.Summary)
	protected("GET /api/v1/dashboard/trend", dashboardHandler.Trend)

	mux.HandleFunc("/", handler.NotFound)

	return middleware.Logging(middleware.Recover(middleware.CORS(cfg.CORSOrigin)(mux)))
}
