package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Arya-teja/monetra/backend/internal/domain"
	"github.com/Arya-teja/monetra/backend/internal/repository/postgres/db"
)

type DashboardRepository struct {
	queries *db.Queries
}

func NewDashboardRepository(pool *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{queries: db.New(pool)}
}

func (r *DashboardRepository) TotalActiveBalance(ctx context.Context, userID string) (int64, error) {
	total, err := r.queries.TotalActiveBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("sum balances: %w", err)
	}
	return total, nil
}

func (r *DashboardRepository) MonthlyTotals(ctx context.Context, userID string, from, to time.Time) ([]domain.MonthTotal, error) {
	rows, err := r.queries.MonthlyTotals(ctx, db.MonthlyTotalsParams{UserID: userID, FromDate: from, ToDate: to})
	if err != nil {
		return nil, fmt.Errorf("sum monthly totals: %w", err)
	}

	totals := make([]domain.MonthTotal, 0, len(rows))
	for _, row := range rows {
		totals = append(totals, domain.MonthTotal{Month: row.Month, Income: row.Income, Expense: row.Expense})
	}
	return totals, nil
}

func (r *DashboardRepository) TopExpenseCategories(ctx context.Context, userID string, from, to time.Time, limit int) ([]domain.CategoryTotal, error) {
	rows, err := r.queries.TopExpenseCategories(ctx, db.TopExpenseCategoriesParams{
		UserID:   userID,
		FromDate: from,
		ToDate:   to,
		RowLimit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("select top categories: %w", err)
	}

	totals := make([]domain.CategoryTotal, 0, len(rows))
	for _, row := range rows {
		totals = append(totals, domain.CategoryTotal{CategoryID: row.ID, CategoryName: row.Name, Total: row.Total})
	}
	return totals, nil
}
