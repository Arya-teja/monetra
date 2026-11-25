package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

const (
	maxNoteLength      = 200
	defaultPageSize    = 20
	maxPageSize        = 100
	msgNotFound        = "not found"
	msgArchivedAccount = "account is archived"
)

type CreateTransactionInput struct {
	AccountID  string
	CategoryID string
	Type       domain.CashflowType
	Amount     int64
	OccurredAt string // YYYY-MM-DD
	Note       string
}

// UpdateTransactionInput changes only the fields that are not nil. An empty Note clears it.
type UpdateTransactionInput struct {
	AccountID  *string
	CategoryID *string
	Type       *domain.CashflowType
	Amount     *int64
	OccurredAt *string
	Note       *string
}

// ListTransactionsInput holds the raw query parameters; empty strings mean "no filter".
type ListTransactionsInput struct {
	Month      string // YYYY-MM
	AccountID  string
	CategoryID string
	Type       string
	Page       int // 0 means the first page
	PageSize   int // 0 means the default page size
}

type TransactionPage struct {
	Items    []domain.Transaction
	Page     int
	PageSize int
	Total    int64
}

type TransactionService struct {
	transactions domain.TransactionRepository
	accounts     domain.AccountRepository
	categories   domain.CategoryRepository
}

func NewTransactionService(
	transactions domain.TransactionRepository,
	accounts domain.AccountRepository,
	categories domain.CategoryRepository,
) *TransactionService {
	return &TransactionService{transactions: transactions, accounts: accounts, categories: categories}
}

func (s *TransactionService) List(ctx context.Context, userID string, in ListTransactionsInput) (TransactionPage, error) {
	filter := domain.TransactionFilter{UserID: userID}
	fields := map[string]string{}

	if in.Month != "" {
		if from, ok := parseMonth(in.Month, fields); ok {
			to := from.AddDate(0, 1, 0)
			filter.From, filter.To = &from, &to
		}
	}
	if in.AccountID != "" {
		if !isUUID(in.AccountID) {
			fields["accountId"] = "must be a valid id"
		}
		filter.AccountID = &in.AccountID
	}
	if in.CategoryID != "" {
		if !isUUID(in.CategoryID) {
			fields["categoryId"] = "must be a valid id"
		}
		filter.CategoryID = &in.CategoryID
	}
	if in.Type != "" {
		t := domain.CashflowType(in.Type)
		if !t.Valid() {
			fields["type"] = "must be income or expense"
		}
		filter.Type = &t
	}

	page, pageSize := in.Page, in.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	if page < 1 {
		fields["page"] = "must be at least 1"
	}
	if pageSize < 1 || pageSize > maxPageSize {
		fields["pageSize"] = fmt.Sprintf("must be 1-%d", maxPageSize)
	}
	if len(fields) > 0 {
		return TransactionPage{}, &domain.ValidationError{Fields: fields}
	}

	filter.Limit = pageSize
	filter.Offset = (page - 1) * pageSize

	items, total, err := s.transactions.List(ctx, filter)
	if err != nil {
		return TransactionPage{}, fmt.Errorf("list transactions: %w", err)
	}
	return TransactionPage{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *TransactionService) Get(ctx context.Context, userID, id string) (domain.Transaction, error) {
	t, err := s.transactions.Get(ctx, userID, id)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("get transaction: %w", err)
	}
	return t, nil
}

// Create stores a transaction and applies its amount to the account balance atomically.
func (s *TransactionService) Create(ctx context.Context, userID string, in CreateTransactionInput) (domain.Transaction, error) {
	t := domain.Transaction{
		UserID:     userID,
		AccountID:  in.AccountID,
		CategoryID: in.CategoryID,
		Type:       in.Type,
		Amount:     in.Amount,
		Note:       optionalText(in.Note),
	}

	fields := map[string]string{}
	t.OccurredAt = parseDate(in.OccurredAt, fields)
	validateTransaction(t, fields)
	if len(fields) == 0 {
		if err := s.checkReferences(ctx, t, nil, fields); err != nil {
			return domain.Transaction{}, err
		}
	}
	if len(fields) > 0 {
		return domain.Transaction{}, &domain.ValidationError{Fields: fields}
	}

	created, err := s.transactions.Create(ctx, t, balanceChanges(nil, &t))
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("create transaction: %w", err)
	}
	return created, nil
}

// Update changes a transaction and corrects the affected account balances atomically.
func (s *TransactionService) Update(ctx context.Context, userID, id string, in UpdateTransactionInput) (domain.Transaction, error) {
	old, err := s.transactions.Get(ctx, userID, id)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("get transaction: %w", err)
	}

	t := old
	fields := map[string]string{}
	if in.AccountID != nil {
		t.AccountID = *in.AccountID
	}
	if in.CategoryID != nil {
		t.CategoryID = *in.CategoryID
	}
	if in.Type != nil {
		t.Type = *in.Type
	}
	if in.Amount != nil {
		t.Amount = *in.Amount
	}
	if in.OccurredAt != nil {
		t.OccurredAt = parseDate(*in.OccurredAt, fields)
	}
	if in.Note != nil {
		t.Note = optionalText(*in.Note)
	}

	validateTransaction(t, fields)
	if len(fields) == 0 {
		if err := s.checkReferences(ctx, t, &old, fields); err != nil {
			return domain.Transaction{}, err
		}
	}
	if len(fields) > 0 {
		return domain.Transaction{}, &domain.ValidationError{Fields: fields}
	}

	updated, err := s.transactions.Update(ctx, t, old.UpdatedAt, balanceChanges(&old, &t))
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("update transaction: %w", err)
	}
	return updated, nil
}

// Delete removes a transaction and reverts its effect on the account balance atomically.
func (s *TransactionService) Delete(ctx context.Context, userID, id string) error {
	old, err := s.transactions.Get(ctx, userID, id)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}
	if err := s.transactions.Delete(ctx, old, balanceChanges(&old, nil)); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	return nil
}

// checkReferences verifies that the account and category belong to the user, that the
// category type matches the transaction type, and that a newly chosen account is not
// archived. old is nil when creating. Problems are added to fields.
func (s *TransactionService) checkReferences(ctx context.Context, t domain.Transaction, old *domain.Transaction, fields map[string]string) error {
	account, err := s.accounts.Get(ctx, t.UserID, t.AccountID)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		fields["accountId"] = msgNotFound
	case err != nil:
		return fmt.Errorf("get account: %w", err)
	case account.IsArchived && (old == nil || old.AccountID != t.AccountID):
		fields["accountId"] = msgArchivedAccount
	}

	category, err := s.categories.Get(ctx, t.UserID, t.CategoryID)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		fields["categoryId"] = msgNotFound
	case err != nil:
		return fmt.Errorf("get category: %w", err)
	case category.Type != t.Type:
		fields["categoryId"] = "category type must match the transaction type"
	}
	return nil
}

// balanceChanges returns what must be added to account balances when a transaction
// goes from before to after. before is nil for a create, after is nil for a delete.
func balanceChanges(before, after *domain.Transaction) []domain.BalanceChange {
	deltas := map[string]int64{}
	var order []string
	add := func(accountID string, delta int64) {
		if _, seen := deltas[accountID]; !seen {
			order = append(order, accountID)
		}
		deltas[accountID] += delta
	}

	if before != nil {
		add(before.AccountID, -signedAmount(*before))
	}
	if after != nil {
		add(after.AccountID, signedAmount(*after))
	}

	var changes []domain.BalanceChange
	for _, accountID := range order {
		if deltas[accountID] != 0 {
			changes = append(changes, domain.BalanceChange{AccountID: accountID, Delta: deltas[accountID]})
		}
	}
	return changes
}

// signedAmount is the effect of a transaction on its account balance.
func signedAmount(t domain.Transaction) int64 {
	if t.Type == domain.CashflowExpense {
		return -t.Amount
	}
	return t.Amount
}

func validateTransaction(t domain.Transaction, fields map[string]string) {
	if !isUUID(t.AccountID) {
		fields["accountId"] = "must be a valid id"
	}
	if !isUUID(t.CategoryID) {
		fields["categoryId"] = "must be a valid id"
	}
	if !t.Type.Valid() {
		fields["type"] = "must be income or expense"
	}
	if t.Amount <= 0 {
		fields["amount"] = "must be greater than 0"
	}
	if t.Note != nil && utf8.RuneCountInString(*t.Note) > maxNoteLength {
		fields["note"] = fmt.Sprintf("must be at most %d characters", maxNoteLength)
	}
}

// isUUID reports whether s has the canonical 8-4-4-4-12 hexadecimal UUID form.
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
				return false
			}
		}
	}
	return true
}
