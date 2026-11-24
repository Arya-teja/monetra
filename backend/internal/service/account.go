package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

const maxAccountTextLength = 50

type CreateAccountInput struct {
	Name            string
	Type            domain.AccountType
	InstitutionName string
	AccountNumber   string
	Balance         int64
}

// UpdateAccountInput changes only the fields that are not nil.
// An empty InstitutionName or AccountNumber clears that field.
type UpdateAccountInput struct {
	Name            *string
	Type            *domain.AccountType
	InstitutionName *string
	AccountNumber   *string
	IsArchived      *bool
}

type AccountService struct {
	accounts domain.AccountRepository
}

func NewAccountService(accounts domain.AccountRepository) *AccountService {
	return &AccountService{accounts: accounts}
}

func (s *AccountService) List(ctx context.Context, userID string, includeArchived bool) ([]domain.Account, error) {
	accounts, err := s.accounts.List(ctx, userID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return accounts, nil
}

// Create stores a new account. The initial balance can only be set here;
// afterwards the balance changes through transactions only.
func (s *AccountService) Create(ctx context.Context, userID string, in CreateAccountInput) (domain.Account, error) {
	account := domain.Account{
		UserID:          userID,
		Name:            strings.TrimSpace(in.Name),
		Type:            in.Type,
		InstitutionName: optionalText(in.InstitutionName),
		AccountNumber:   optionalText(in.AccountNumber),
		Balance:         in.Balance,
	}

	fields := validateAccount(account)
	if in.Balance < 0 {
		fields["balance"] = "must not be negative"
	}
	if len(fields) > 0 {
		return domain.Account{}, &domain.ValidationError{Fields: fields}
	}

	created, err := s.accounts.Create(ctx, account)
	if err != nil {
		return domain.Account{}, fmt.Errorf("create account: %w", err)
	}
	return created, nil
}

func (s *AccountService) Update(ctx context.Context, userID, id string, in UpdateAccountInput) (domain.Account, error) {
	account, err := s.accounts.Get(ctx, userID, id)
	if err != nil {
		return domain.Account{}, fmt.Errorf("get account: %w", err)
	}

	if in.Name != nil {
		account.Name = strings.TrimSpace(*in.Name)
	}
	if in.Type != nil {
		account.Type = *in.Type
	}
	if in.InstitutionName != nil {
		account.InstitutionName = optionalText(*in.InstitutionName)
	}
	if in.AccountNumber != nil {
		account.AccountNumber = optionalText(*in.AccountNumber)
	}
	if in.IsArchived != nil {
		account.IsArchived = *in.IsArchived
	}

	if fields := validateAccount(account); len(fields) > 0 {
		return domain.Account{}, &domain.ValidationError{Fields: fields}
	}

	updated, err := s.accounts.Update(ctx, account)
	if err != nil {
		return domain.Account{}, fmt.Errorf("update account: %w", err)
	}
	return updated, nil
}

func (s *AccountService) Delete(ctx context.Context, userID, id string) error {
	if err := s.accounts.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

func validateAccount(a domain.Account) map[string]string {
	fields := map[string]string{}
	if n := utf8.RuneCountInString(a.Name); n == 0 || n > maxAccountTextLength {
		fields["name"] = fmt.Sprintf("must be 1-%d characters", maxAccountTextLength)
	}
	if !a.Type.Valid() {
		fields["type"] = "must be one of bank, ewallet, cash"
	}
	if a.InstitutionName != nil && utf8.RuneCountInString(*a.InstitutionName) > maxAccountTextLength {
		fields["institutionName"] = fmt.Sprintf("must be at most %d characters", maxAccountTextLength)
	}
	if a.AccountNumber != nil && utf8.RuneCountInString(*a.AccountNumber) > maxAccountTextLength {
		fields["accountNumber"] = fmt.Sprintf("must be at most %d characters", maxAccountTextLength)
	}
	return fields
}

// optionalText trims s and turns an empty result into nil (stored as NULL).
func optionalText(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
