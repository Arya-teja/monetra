package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Arya-teja/monetra/backend/internal/config"
	"github.com/Arya-teja/monetra/backend/internal/http/handler"
	"github.com/Arya-teja/monetra/backend/internal/http/middleware"
	"github.com/Arya-teja/monetra/backend/internal/repository/postgres"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

// accessTokenTTL matches the NextAuth session maxAge on the frontend.
const accessTokenTTL = 7 * 24 * time.Hour

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()

	tokens := service.NewTokenManager(cfg.JWTSecret, accessTokenTTL)
	authService := service.NewAuthService(postgres.NewUserRepository(pool), tokens)
	accountRepo := postgres.NewAccountRepository(pool)
	categoryRepo := postgres.NewCategoryRepository(pool)
	transactionRepo := postgres.NewTransactionRepository(pool)
	budgetRepo := postgres.NewBudgetRepository(pool)

	accountService := service.NewAccountService(accountRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	transactionService := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo)
	budgetService := service.NewBudgetService(budgetRepo, categoryRepo)
	dashboardService := service.NewDashboardService(postgres.NewDashboardRepository(pool), transactionRepo, budgetRepo, time.Now)

	health := handler.NewHealthHandler(pool)
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	requireAuth := middleware.Auth(tokens)

	mux := http.NewServeMux()
	protected := func(pattern string, h http.HandlerFunc) {
		mux.Handle(pattern, requireAuth(h))
	}

	mux.HandleFunc("GET /api/v1/health", health.Check)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	protected("GET /api/v1/auth/me", authHandler.Me)

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

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           middleware.Logging(middleware.Recover(middleware.CORS(cfg.CORSOrigin)(mux))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
