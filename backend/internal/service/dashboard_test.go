package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

// fakeDashboardRepository returns fixed data and records the requested date range.
type fakeDashboardRepository struct {
	balance  int64
	monthly  []domain.MonthTotal
	top      []domain.CategoryTotal
	gotFrom  time.Time
	gotTo    time.Time
	gotLimit int
}

func (r *fakeDashboardRepository) TotalActiveBalance(context.Context, string) (int64, error) {
	return r.balance, nil
}

func (r *fakeDashboardRepository) MonthlyTotals(_ context.Context, _ string, from, to time.Time) ([]domain.MonthTotal, error) {
	r.gotFrom, r.gotTo = from, to
	return r.monthly, nil
}

func (r *fakeDashboardRepository) TopExpenseCategories(_ context.Context, _ string, _, _ time.Time, limit int) ([]domain.CategoryTotal, error) {
	r.gotLimit = limit
	return r.top, nil
}

func month(s string) time.Time {
	m, err := time.Parse(monthLayout, s)
	if err != nil {
		panic(err)
	}
	return m
}

func newDashboardFixture(now time.Time) (*DashboardService, *fakeDashboardRepository, *fakeBudgetRepository) {
	dash := &fakeDashboardRepository{}
	budgets := newFakeBudgetRepository()
	transactions := &fakeTransactionRepository{transactions: map[string]domain.Transaction{}}
	svc := NewDashboardService(dash, transactions, budgets, func() time.Time { return now })
	return svc, dash, budgets
}

func TestUsedPercentage(t *testing.T) {
	if got := usedPercentage(0, 500); got != nil {
		t.Errorf("usedPercentage(0, 500) = %d, want nil", *got)
	}
	if got := usedPercentage(1000, 755); got == nil || *got != 75 {
		t.Errorf("usedPercentage(1000, 755) = %v, want 75", got)
	}
	if got := usedPercentage(1000, 1500); got == nil || *got != 150 {
		t.Errorf("usedPercentage(1000, 1500) = %v, want 150", got)
	}
}

func TestDashboardSummary_DefaultsToCurrentMonthInJakarta(t *testing.T) {
	// 30 Nov 20:00 UTC is already 1 Dec 03:00 in Jakarta.
	svc, dash, _ := newDashboardFixture(time.Date(2026, 11, 30, 20, 0, 0, 0, time.UTC))

	summary, err := svc.Summary(context.Background(), "user-1", "")
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if got := summary.Month.Format(dateLayout); got != "2026-12-01" {
		t.Errorf("month = %s, want 2026-12-01", got)
	}
	if dash.gotFrom.Format(dateLayout) != "2026-12-01" || dash.gotTo.Format(dateLayout) != "2027-01-01" {
		t.Errorf("range = %v..%v, want 2026-12-01..2027-01-01", dash.gotFrom, dash.gotTo)
	}
	if dash.gotLimit != topCategoryCount {
		t.Errorf("top categories limit = %d, want %d", dash.gotLimit, topCategoryCount)
	}
}

func TestDashboardSummary_TotalsAndBudgetAlerts(t *testing.T) {
	svc, dash, budgets := newDashboardFixture(time.Now())
	dash.balance = 2_500_000
	dash.monthly = []domain.MonthTotal{{Month: month("2026-11"), Income: 4_000_000, Expense: 3_000_000}}
	for i, spent := range []int64{100, 850, 1200} {
		budgets.budgets[testUUID('b', i)] = domain.Budget{
			ID: testUUID('b', i), UserID: "user-1", Month: month("2026-11"), Amount: 1000, Spent: spent,
		}
	}

	summary, err := svc.Summary(context.Background(), "user-1", "2026-11")
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.TotalBalance != 2_500_000 || summary.Income != 4_000_000 || summary.Expense != 3_000_000 {
		t.Errorf("balance/income/expense = %d/%d/%d", summary.TotalBalance, summary.Income, summary.Expense)
	}
	if summary.UsedPercentage == nil || *summary.UsedPercentage != 75 {
		t.Errorf("used percentage = %v, want 75", summary.UsedPercentage)
	}
	if len(summary.BudgetAlerts) != 2 {
		t.Errorf("budget alerts = %d, want 2 (warning and over, not safe)", len(summary.BudgetAlerts))
	}
	for _, a := range summary.BudgetAlerts {
		if a.Status == domain.BudgetSafe {
			t.Errorf("safe budget %s listed as alert", a.ID)
		}
	}
}

func TestDashboardSummary_NoIncomeHasNoPercentage(t *testing.T) {
	svc, _, _ := newDashboardFixture(time.Now())

	summary, err := svc.Summary(context.Background(), "user-1", "2026-11")
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.UsedPercentage != nil {
		t.Errorf("used percentage = %d, want nil when income is 0", *summary.UsedPercentage)
	}
}

func TestDashboardSummary_RejectsInvalidMonth(t *testing.T) {
	svc, _, _ := newDashboardFixture(time.Now())

	var validationErr *domain.ValidationError
	if _, err := svc.Summary(context.Background(), "user-1", "2026-13"); !errors.As(err, &validationErr) {
		t.Fatalf("Summary() error = %v, want *domain.ValidationError", err)
	}
}

func TestDashboardTrend_FillsMissingMonthsOldestFirst(t *testing.T) {
	svc, dash, _ := newDashboardFixture(time.Date(2027, 2, 10, 12, 0, 0, 0, time.UTC))
	dash.monthly = []domain.MonthTotal{
		{Month: month("2026-10"), Income: 5_000_000, Expense: 1_000_000},
		{Month: month("2027-01"), Expense: 300_000},
	}

	trend, err := svc.Trend(context.Background(), "user-1", 0)
	if err != nil {
		t.Fatalf("Trend() error = %v", err)
	}

	want := []struct {
		month           string
		income, expense int64
	}{
		{"2026-09", 0, 0},
		{"2026-10", 5_000_000, 1_000_000},
		{"2026-11", 0, 0},
		{"2026-12", 0, 0},
		{"2027-01", 0, 300_000},
		{"2027-02", 0, 0},
	}
	if len(trend) != len(want) {
		t.Fatalf("trend has %d months, want %d", len(trend), len(want))
	}
	for i, w := range want {
		got := trend[i]
		if got.Month.Format(monthLayout) != w.month || got.Income != w.income || got.Expense != w.expense {
			t.Errorf("trend[%d] = %s %d/%d, want %s %d/%d",
				i, got.Month.Format(monthLayout), got.Income, got.Expense, w.month, w.income, w.expense)
		}
	}
	if dash.gotFrom.Format(dateLayout) != "2026-09-01" || dash.gotTo.Format(dateLayout) != "2027-03-01" {
		t.Errorf("range = %v..%v, want 2026-09-01..2027-03-01", dash.gotFrom, dash.gotTo)
	}
}

func TestDashboardTrend_RejectsInvalidMonths(t *testing.T) {
	svc, _, _ := newDashboardFixture(time.Now())

	for _, months := range []int{-1, maxTrendMonths + 1} {
		var validationErr *domain.ValidationError
		if _, err := svc.Trend(context.Background(), "user-1", months); !errors.As(err, &validationErr) {
			t.Errorf("Trend(%d) error = %v, want *domain.ValidationError", months, err)
		}
	}
}
