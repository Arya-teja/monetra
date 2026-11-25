package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Arya-teja/monetra/backend/internal/app"
	"github.com/Arya-teja/monetra/backend/internal/testdb"
)

// now is fixed so "the current month" in the dashboard is always November 2026 (WIB).
var now = time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)

type apiError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details"`
}

type response struct {
	Status int
	Data   json.RawMessage
	Meta   *struct {
		Page     int   `json:"page"`
		PageSize int   `json:"pageSize"`
		Total    int64 `json:"total"`
	}
	Error *apiError
}

type client struct {
	t     *testing.T
	url   string
	token string
}

func newServer(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	t.Helper()
	pool := testdb.Open(t)
	srv := httptest.NewServer(app.NewHandler(pool, app.Config{
		JWTSecret:  strings.Repeat("s", 32),
		CORSOrigin: "http://localhost:3000",
		Now:        func() time.Time { return now },
	}))
	t.Cleanup(srv.Close)
	return srv, pool
}

func (c *client) do(method, path string, body any) response {
	c.t.Helper()
	var reader *bytes.Reader
	switch b := body.(type) {
	case nil:
		reader = bytes.NewReader(nil)
	case string:
		reader = bytes.NewReader([]byte(b))
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			c.t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, c.url+path, reader)
	if err != nil {
		c.t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()

	out := response{Status: res.StatusCode}
	if res.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
			c.t.Fatalf("%s %s: decode body: %v", method, path, err)
		}
	}
	return out
}

// expect fails the test unless the response has the wanted status.
func (c *client) expect(want int, method, path string, body any) response {
	c.t.Helper()
	res := c.do(method, path, body)
	if res.Status != want {
		c.t.Fatalf("%s %s: status %d, want %d (error: %+v)", method, path, res.Status, want, res.Error)
	}
	return res
}

func decode[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return v
}

type idOnly struct {
	ID string `json:"id"`
}

func signUp(t *testing.T, srv *httptest.Server, email string) *client {
	t.Helper()
	anon := &client{t: t, url: srv.URL}
	anon.expect(201, "POST", "/api/v1/auth/register", map[string]string{"name": "Test", "email": email, "password": "rahasia123"})
	res := anon.expect(200, "POST", "/api/v1/auth/login", map[string]string{"email": email, "password": "rahasia123"})
	return &client{t: t, url: srv.URL, token: decode[struct {
		AccessToken string `json:"accessToken"`
	}](t, res.Data).AccessToken}
}

func (c *client) createAccount(name string, balance int64) string {
	c.t.Helper()
	res := c.expect(201, "POST", "/api/v1/accounts", map[string]any{"name": name, "type": "bank", "balance": balance})
	return decode[idOnly](c.t, res.Data).ID
}

func (c *client) categoryID(name, typ string) string {
	c.t.Helper()
	res := c.expect(200, "GET", "/api/v1/categories?type="+typ, nil)
	for _, cat := range decode[[]struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}](c.t, res.Data) {
		if cat.Name == name {
			return cat.ID
		}
	}
	c.t.Fatalf("category %q (%s) not found", name, typ)
	return ""
}

func (c *client) createTransaction(accountID, categoryID, typ string, amount int64, date string) string {
	c.t.Helper()
	res := c.expect(201, "POST", "/api/v1/transactions", map[string]any{
		"accountId": accountID, "categoryId": categoryID, "type": typ, "amount": amount, "occurredAt": date,
	})
	return decode[idOnly](c.t, res.Data).ID
}

func (c *client) balance(accountID string) int64 {
	c.t.Helper()
	res := c.expect(200, "GET", "/api/v1/accounts?includeArchived=true", nil)
	for _, a := range decode[[]struct {
		ID      string `json:"id"`
		Balance int64  `json:"balance"`
	}](c.t, res.Data) {
		if a.ID == accountID {
			return a.Balance
		}
	}
	c.t.Fatalf("account %s not found", accountID)
	return 0
}

// assertLedgerConsistent checks that every account balance equals its starting
// balance plus the signed sum of its transactions.
func assertLedgerConsistent(t *testing.T, pool *pgxpool.Pool, starting map[string]int64) {
	t.Helper()
	for id, start := range starting {
		var balance, sum int64
		err := pool.QueryRow(context.Background(), `
			SELECT a.balance,
			       COALESCE(SUM(CASE t.type WHEN 'income' THEN t.amount ELSE -t.amount END), 0)::bigint
			FROM accounts a LEFT JOIN transactions t ON t.account_id = a.id
			WHERE a.id = $1 GROUP BY a.id`, id).Scan(&balance, &sum)
		if err != nil {
			t.Fatalf("ledger query: %v", err)
		}
		if balance != start+sum {
			t.Errorf("account %s: balance %d, want start %d + transactions %d = %d", id, balance, start, sum, start+sum)
		}
	}
}

func TestHealthAndUnknownRoute(t *testing.T) {
	srv, _ := newServer(t)
	c := &client{t: t, url: srv.URL}

	if got := c.expect(200, "GET", "/api/v1/health", nil); string(got.Data) != `{"status":"ok"}` {
		t.Errorf("health data = %s", got.Data)
	}
	if got := c.expect(404, "GET", "/api/v1/nope", nil); got.Error == nil || got.Error.Code != "NOT_FOUND" {
		t.Errorf("unknown route error = %+v, want NOT_FOUND", got.Error)
	}
}

func TestAuth(t *testing.T) {
	srv, _ := newServer(t)
	anon := &client{t: t, url: srv.URL}

	anon.expect(201, "POST", "/api/v1/auth/register", map[string]string{"name": "Arya", "email": " Arya@Example.com ", "password": "rahasia123"})
	if got := anon.expect(409, "POST", "/api/v1/auth/register", map[string]string{"name": "Arya", "email": "arya@example.com", "password": "rahasia123"}); got.Error.Code != "CONFLICT" {
		t.Errorf("duplicate register code = %s", got.Error.Code)
	}
	if got := anon.expect(400, "POST", "/api/v1/auth/register", map[string]string{"name": "A", "email": "x", "password": "1"}); len(got.Error.Details) != 3 {
		t.Errorf("validation details = %v, want name, email, password", got.Error.Details)
	}
	anon.expect(401, "POST", "/api/v1/auth/login", map[string]string{"email": "arya@example.com", "password": "salahsalah"})

	c := signUp(t, srv, "second@example.com")
	me := decode[struct {
		Email string `json:"email"`
	}](t, c.expect(200, "GET", "/api/v1/auth/me", nil).Data)
	if me.Email != "second@example.com" {
		t.Errorf("me email = %q", me.Email)
	}

	(&client{t: t, url: srv.URL, token: "not.a.jwt"}).expect(401, "GET", "/api/v1/auth/me", nil)
}

func TestEveryProtectedRouteRequiresToken(t *testing.T) {
	srv, _ := newServer(t)
	anon := &client{t: t, url: srv.URL}
	id := "00000000-0000-0000-0000-000000000000"

	routes := [][2]string{
		{"GET", "/api/v1/auth/me"}, {"POST", "/api/v1/auth/logout"},
		{"GET", "/api/v1/accounts"}, {"POST", "/api/v1/accounts"},
		{"PATCH", "/api/v1/accounts/" + id}, {"DELETE", "/api/v1/accounts/" + id},
		{"GET", "/api/v1/categories"}, {"POST", "/api/v1/categories"},
		{"PATCH", "/api/v1/categories/" + id}, {"DELETE", "/api/v1/categories/" + id},
		{"GET", "/api/v1/transactions"}, {"POST", "/api/v1/transactions"}, {"GET", "/api/v1/transactions/" + id},
		{"PATCH", "/api/v1/transactions/" + id}, {"DELETE", "/api/v1/transactions/" + id},
		{"GET", "/api/v1/budgets?month=2026-11"}, {"POST", "/api/v1/budgets"}, {"POST", "/api/v1/budgets/copy"},
		{"PATCH", "/api/v1/budgets/" + id}, {"DELETE", "/api/v1/budgets/" + id},
		{"GET", "/api/v1/dashboard/summary"}, {"GET", "/api/v1/dashboard/trend"},
	}
	for _, r := range routes {
		if got := anon.do(r[0], r[1], "{}"); got.Status != 401 {
			t.Errorf("%s %s without token: status %d, want 401", r[0], r[1], got.Status)
		}
	}
}

func TestUsersCannotReachEachOthersData(t *testing.T) {
	srv, _ := newServer(t)
	a := signUp(t, srv, "a@example.com")
	b := signUp(t, srv, "b@example.com")

	account := a.createAccount("BSI", 1_000_000)
	food := a.categoryID("Makanan & Minuman", "expense")
	tx := a.createTransaction(account, food, "expense", 50_000, "2026-11-10")
	budget := decode[idOnly](t, a.expect(201, "POST", "/api/v1/budgets", map[string]any{"categoryId": food, "month": "2026-11", "amount": 500_000}).Data).ID

	// B sees none of A's data...
	for _, path := range []string{"/api/v1/accounts", "/api/v1/transactions", "/api/v1/budgets?month=2026-11"} {
		if got := b.expect(200, "GET", path, nil); string(got.Data) != "[]" {
			t.Errorf("B GET %s = %s, want []", path, got.Data)
		}
	}
	// ...cannot read, change, or delete it (404, not 403, so existence is not revealed)...
	checks := []struct {
		method, path string
		body         any
	}{
		{"PATCH", "/api/v1/accounts/" + account, map[string]string{"name": "x"}},
		{"DELETE", "/api/v1/accounts/" + account, nil},
		{"PATCH", "/api/v1/categories/" + food, map[string]string{"name": "x"}},
		{"DELETE", "/api/v1/categories/" + food, nil},
		{"GET", "/api/v1/transactions/" + tx, nil},
		{"PATCH", "/api/v1/transactions/" + tx, map[string]int64{"amount": 1}},
		{"DELETE", "/api/v1/transactions/" + tx, nil},
		{"PATCH", "/api/v1/budgets/" + budget, map[string]int64{"amount": 1}},
		{"DELETE", "/api/v1/budgets/" + budget, nil},
	}
	for _, ch := range checks {
		if got := b.do(ch.method, ch.path, ch.body); got.Status != 404 {
			t.Errorf("B %s %s: status %d, want 404", ch.method, ch.path, got.Status)
		}
	}
	// ...and cannot book a transaction or budget on A's account or category.
	if got := b.do("POST", "/api/v1/transactions", map[string]any{"accountId": account, "categoryId": food, "type": "expense", "amount": 1, "occurredAt": "2026-11-10"}); got.Status != 400 {
		t.Errorf("B transaction on A's account: status %d, want 400", got.Status)
	}
	if got := b.do("POST", "/api/v1/budgets", map[string]any{"categoryId": food, "month": "2026-11", "amount": 1}); got.Status != 400 {
		t.Errorf("B budget on A's category: status %d, want 400", got.Status)
	}

	// A's data is unchanged.
	if got := a.balance(account); got != 950_000 {
		t.Errorf("A balance = %d, want 950000", got)
	}
}

func TestTransactionsKeepBalancesConsistent(t *testing.T) {
	srv, pool := newServer(t)
	c := signUp(t, srv, "ledger@example.com")
	wallet := c.createAccount("BSI", 1_000_000)
	cash := c.createAccount("Dompet", 100_000)
	food := c.categoryID("Makanan & Minuman", "expense")
	salary := c.categoryID("Gaji", "income")
	starting := map[string]int64{wallet: 1_000_000, cash: 100_000}

	tx := c.createTransaction(wallet, food, "expense", 50_000, "2026-11-10")
	c.createTransaction(wallet, salary, "income", 2_000_000, "2026-11-01")
	assertLedgerConsistent(t, pool, starting)

	c.expect(200, "PATCH", "/api/v1/transactions/"+tx, map[string]any{"amount": 80_000})
	c.expect(200, "PATCH", "/api/v1/transactions/"+tx, map[string]any{"accountId": cash})
	c.expect(200, "PATCH", "/api/v1/transactions/"+tx, map[string]any{"type": "income", "categoryId": salary})
	assertLedgerConsistent(t, pool, starting)
	if got := c.balance(cash); got != 180_000 {
		t.Errorf("cash after moving and flipping the transaction = %d, want 180000", got)
	}

	c.expect(204, "DELETE", "/api/v1/transactions/"+tx, nil)
	assertLedgerConsistent(t, pool, starting)
	if got := c.balance(cash); got != 100_000 {
		t.Errorf("cash after delete = %d, want 100000", got)
	}

	// Expenses may push a balance below zero.
	c.createTransaction(cash, food, "expense", 150_000, "2026-11-12")
	if got := c.balance(cash); got != -50_000 {
		t.Errorf("cash = %d, want -50000", got)
	}
	assertLedgerConsistent(t, pool, starting)
}

func TestFailedBalanceUpdateRollsBackTheTransaction(t *testing.T) {
	srv, pool := newServer(t)
	c := signUp(t, srv, "rollback@example.com")
	wallet := c.createAccount("BSI", 100_000)
	food := c.categoryID("Makanan & Minuman", "expense")

	// Make the balance update fail after the transaction row was inserted.
	if _, err := pool.Exec(context.Background(), "ALTER TABLE accounts ADD CONSTRAINT test_no_negative CHECK (balance >= 0)"); err != nil {
		t.Fatal(err)
	}

	got := c.do("POST", "/api/v1/transactions", map[string]any{"accountId": wallet, "categoryId": food, "type": "expense", "amount": 500_000, "occurredAt": "2026-11-10"})
	if got.Status != 500 {
		t.Fatalf("status %d, want 500", got.Status)
	}

	var count int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM transactions").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("transactions after failed balance update = %d, want 0 (rolled back)", count)
	}
	if b := c.balance(wallet); b != 100_000 {
		t.Errorf("balance = %d, want 100000 (unchanged)", b)
	}
}

func TestDeleteRules(t *testing.T) {
	srv, _ := newServer(t)
	c := signUp(t, srv, "delete@example.com")
	wallet := c.createAccount("BSI", 0)
	food := c.categoryID("Makanan & Minuman", "expense")
	c.createTransaction(wallet, food, "expense", 1_000, "2026-11-10")

	if got := c.do("DELETE", "/api/v1/accounts/"+wallet, nil); got.Status != 409 {
		t.Errorf("delete account in use: status %d, want 409", got.Status)
	}
	if got := c.do("DELETE", "/api/v1/categories/"+food, nil); got.Status != 409 {
		t.Errorf("delete category in use: status %d, want 409", got.Status)
	}

	// A category with only a budget (no transactions) can be deleted; its budget goes too.
	fun := c.categoryID("Hiburan", "expense")
	c.expect(201, "POST", "/api/v1/budgets", map[string]any{"categoryId": fun, "month": "2026-11", "amount": 100_000})
	c.expect(204, "DELETE", "/api/v1/categories/"+fun, nil)
	if got := c.expect(200, "GET", "/api/v1/budgets?month=2026-11", nil); string(got.Data) != "[]" {
		t.Errorf("budgets after deleting their category = %s, want []", got.Data)
	}

	// New transactions on an archived account are rejected.
	c.expect(200, "PATCH", "/api/v1/accounts/"+wallet, map[string]any{"isArchived": true})
	if got := c.do("POST", "/api/v1/transactions", map[string]any{"accountId": wallet, "categoryId": food, "type": "expense", "amount": 1, "occurredAt": "2026-11-10"}); got.Status != 400 || got.Error.Details["accountId"] == "" {
		t.Errorf("transaction on archived account: status %d, error %+v; want 400 on accountId", got.Status, got.Error)
	}
}

func TestBudgetSpentUsesCalendarMonth(t *testing.T) {
	srv, _ := newServer(t)
	c := signUp(t, srv, "budget@example.com")
	wallet := c.createAccount("BSI", 5_000_000)
	food := c.categoryID("Makanan & Minuman", "expense")

	c.expect(201, "POST", "/api/v1/budgets", map[string]any{"categoryId": food, "month": "2026-10", "amount": 1_000_000})
	c.createTransaction(wallet, food, "expense", 500_000, "2026-10-01")
	c.createTransaction(wallet, food, "expense", 350_000, "2026-10-31")
	c.createTransaction(wallet, food, "expense", 999_000, "2026-11-01") // next month: not counted
	c.createTransaction(wallet, food, "expense", 999_000, "2026-09-30") // previous month: not counted

	budgets := decode[[]struct {
		Spent      int64  `json:"spent"`
		Percentage int64  `json:"percentage"`
		Status     string `json:"status"`
	}](t, c.expect(200, "GET", "/api/v1/budgets?month=2026-10", nil).Data)
	if len(budgets) != 1 || budgets[0].Spent != 850_000 || budgets[0].Percentage != 85 || budgets[0].Status != "warning" {
		t.Errorf("October budget = %+v, want spent 850000, 85%%, warning", budgets)
	}

	// Copying into November keeps existing budgets and copies once.
	copied := func() int64 {
		return decode[struct {
			Copied int64 `json:"copied"`
		}](t, c.expect(200, "POST", "/api/v1/budgets/copy", map[string]string{"month": "2026-11"}).Data).Copied
	}
	if first, second := copied(), copied(); first != 1 || second != 0 {
		t.Errorf("copy results = %d then %d, want 1 then 0", first, second)
	}
	if got := c.do("POST", "/api/v1/budgets", map[string]any{"categoryId": food, "month": "2026-11", "amount": 1}); got.Status != 409 {
		t.Errorf("second budget for the same category and month: status %d, want 409", got.Status)
	}
}

func TestDashboardSummaryAndTrend(t *testing.T) {
	srv, _ := newServer(t)
	c := signUp(t, srv, "dash@example.com")
	wallet := c.createAccount("BSI", 1_000_000)
	old := c.createAccount("Lama", 777)
	c.expect(200, "PATCH", "/api/v1/accounts/"+old, map[string]any{"isArchived": true})
	food := c.categoryID("Makanan & Minuman", "expense")
	salary := c.categoryID("Gaji", "income")

	c.createTransaction(wallet, salary, "income", 4_000_000, "2026-11-01")
	c.createTransaction(wallet, food, "expense", 900_000, "2026-11-05")
	c.createTransaction(wallet, food, "expense", 100_000, "2026-09-05")

	summary := decode[struct {
		Month          string `json:"month"`
		TotalBalance   int64  `json:"totalBalance"`
		Income         int64  `json:"income"`
		Expense        int64  `json:"expense"`
		UsedPercentage *int64 `json:"usedPercentage"`
		TopCategories  []struct {
			Total int64 `json:"total"`
		} `json:"topCategories"`
		RecentTransactions []json.RawMessage `json:"recentTransactions"`
	}](t, c.expect(200, "GET", "/api/v1/dashboard/summary", nil).Data)

	if summary.Month != "2026-11" {
		t.Errorf("default month = %s, want 2026-11", summary.Month)
	}
	if summary.TotalBalance != 4_000_000 { // archived account excluded
		t.Errorf("total balance = %d, want 4000000", summary.TotalBalance)
	}
	if summary.Income != 4_000_000 || summary.Expense != 900_000 || summary.UsedPercentage == nil || *summary.UsedPercentage != 22 {
		t.Errorf("income/expense/used = %d/%d/%v, want 4000000/900000/22", summary.Income, summary.Expense, summary.UsedPercentage)
	}
	if len(summary.TopCategories) != 1 || summary.TopCategories[0].Total != 900_000 {
		t.Errorf("top categories = %+v", summary.TopCategories)
	}
	if len(summary.RecentTransactions) != 3 {
		t.Errorf("recent transactions = %d, want 3", len(summary.RecentTransactions))
	}

	trend := decode[[]struct {
		Month   string `json:"month"`
		Expense int64  `json:"expense"`
	}](t, c.expect(200, "GET", "/api/v1/dashboard/trend", nil).Data)
	if len(trend) != 6 || trend[0].Month != "2026-06" || trend[5].Month != "2026-11" || trend[3].Expense != 100_000 {
		t.Errorf("trend = %+v, want Jun..Nov with September expense 100000", trend)
	}
}

func TestTransactionListFiltersAndPages(t *testing.T) {
	srv, _ := newServer(t)
	c := signUp(t, srv, "list@example.com")
	wallet := c.createAccount("BSI", 0)
	food := c.categoryID("Makanan & Minuman", "expense")
	for day := 1; day <= 5; day++ {
		c.createTransaction(wallet, food, "expense", int64(day*1000), fmt.Sprintf("2026-11-%02d", day))
	}
	c.createTransaction(wallet, food, "expense", 9_000, "2026-10-15")

	got := c.expect(200, "GET", "/api/v1/transactions?month=2026-11&pageSize=2&page=2", nil)
	items := decode[[]struct {
		OccurredAt string `json:"occurredAt"`
	}](t, got.Data)
	if got.Meta == nil || got.Meta.Total != 5 || got.Meta.Page != 2 || len(items) != 2 || items[0].OccurredAt != "2026-11-03" {
		t.Errorf("page 2 = %+v meta %+v, want 2 items starting 2026-11-03 of 5", items, got.Meta)
	}

	if bad := c.do("GET", "/api/v1/transactions?month=2026-13&accountId=abc", nil); bad.Status != 400 || len(bad.Error.Details) != 2 {
		t.Errorf("invalid filters: status %d details %v, want 400 with month and accountId", bad.Status, bad.Error.Details)
	}
}

func TestRequestBodyValidation(t *testing.T) {
	srv, _ := newServer(t)
	c := signUp(t, srv, "body@example.com")

	if got := c.do("POST", "/api/v1/accounts", "{not json"); got.Status != 400 || got.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("malformed JSON: %d %+v", got.Status, got.Error)
	}
	// Unknown fields are rejected, so the balance cannot be edited through PATCH.
	account := c.createAccount("BSI", 1_000)
	if got := c.do("PATCH", "/api/v1/accounts/"+account, map[string]any{"balance": 999_999}); got.Status != 400 {
		t.Errorf("PATCH balance: status %d, want 400", got.Status)
	}
	if got := c.balance(account); got != 1_000 {
		t.Errorf("balance = %d, want 1000", got)
	}
}

func TestLogoutRevokesTokensOnEveryDevice(t *testing.T) {
	srv, _ := newServer(t)
	phone := signUp(t, srv, "logout@example.com")

	anon := &client{t: t, url: srv.URL}
	res := anon.expect(200, "POST", "/api/v1/auth/login", map[string]string{"email": "logout@example.com", "password": "rahasia123"})
	laptop := &client{t: t, url: srv.URL, token: decode[struct {
		AccessToken string `json:"accessToken"`
	}](t, res.Data).AccessToken}

	laptop.expect(200, "GET", "/api/v1/auth/me", nil)
	phone.expect(204, "POST", "/api/v1/auth/logout", nil)

	for name, c := range map[string]*client{"phone": phone, "laptop": laptop} {
		if got := c.do("GET", "/api/v1/accounts", nil); got.Status != 401 {
			t.Errorf("%s after logout: status %d, want 401", name, got.Status)
		}
	}
	// Logging in again works.
	signedIn := anon.expect(200, "POST", "/api/v1/auth/login", map[string]string{"email": "logout@example.com", "password": "rahasia123"})
	fresh := &client{t: t, url: srv.URL, token: decode[struct {
		AccessToken string `json:"accessToken"`
	}](t, signedIn.Data).AccessToken}
	fresh.expect(200, "GET", "/api/v1/auth/me", nil)
}
