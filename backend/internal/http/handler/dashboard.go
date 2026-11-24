package handler

import (
	"net/http"

	"github.com/Arya-teja/monetra/backend/internal/http/middleware"
	"github.com/Arya-teja/monetra/backend/internal/http/response"
	"github.com/Arya-teja/monetra/backend/internal/service"
)

type DashboardHandler struct {
	dashboard *service.DashboardService
}

func NewDashboardHandler(dashboard *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

type categoryTotalResponse struct {
	CategoryID   string `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Total        int64  `json:"total"`
}

type monthTotalResponse struct {
	Month   string `json:"month"`
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	s, err := h.dashboard.Summary(r.Context(), middleware.UserID(r.Context()), r.URL.Query().Get("month"))
	if err != nil {
		writeError(w, err)
		return
	}

	top := make([]categoryTotalResponse, 0, len(s.TopCategories))
	for _, c := range s.TopCategories {
		top = append(top, categoryTotalResponse{CategoryID: c.CategoryID, CategoryName: c.CategoryName, Total: c.Total})
	}
	alerts := make([]budgetResponse, 0, len(s.BudgetAlerts))
	for _, b := range s.BudgetAlerts {
		alerts = append(alerts, toBudgetResponse(b))
	}
	recent := make([]transactionResponse, 0, len(s.RecentTransactions))
	for _, t := range s.RecentTransactions {
		recent = append(recent, toTransactionResponse(t))
	}

	response.JSON(w, http.StatusOK, struct {
		Month              string                  `json:"month"`
		TotalBalance       int64                   `json:"totalBalance"`
		Income             int64                   `json:"income"`
		Expense            int64                   `json:"expense"`
		UsedPercentage     *int64                  `json:"usedPercentage"`
		TopCategories      []categoryTotalResponse `json:"topCategories"`
		BudgetAlerts       []budgetResponse        `json:"budgetAlerts"`
		RecentTransactions []transactionResponse   `json:"recentTransactions"`
	}{
		Month:              s.Month.Format(monthLayout),
		TotalBalance:       s.TotalBalance,
		Income:             s.Income,
		Expense:            s.Expense,
		UsedPercentage:     s.UsedPercentage,
		TopCategories:      top,
		BudgetAlerts:       alerts,
		RecentTransactions: recent,
	})
}

func (h *DashboardHandler) Trend(w http.ResponseWriter, r *http.Request) {
	months, invalid := optionalInt(r.URL.Query().Get("months"))
	if invalid {
		response.ValidationError(w, map[string]string{"months": "must be a number"})
		return
	}

	trend, err := h.dashboard.Trend(r.Context(), middleware.UserID(r.Context()), months)
	if err != nil {
		writeError(w, err)
		return
	}

	body := make([]monthTotalResponse, 0, len(trend))
	for _, t := range trend {
		body = append(body, monthTotalResponse{Month: t.Month.Format(monthLayout), Income: t.Income, Expense: t.Expense})
	}
	response.JSON(w, http.StatusOK, body)
}
