package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

const (
	minNameLength     = 2
	minPasswordLength = 8
	// bcrypt only uses the first 72 bytes of a password.
	maxPasswordBytes = 72
)

// defaultCategories are created for every new user (PRD Appendix A).
var defaultCategories = []domain.Category{
	{Name: "Gaji", Type: domain.CashflowIncome, IsDefault: true},
	{Name: "Bonus", Type: domain.CashflowIncome, IsDefault: true},
	{Name: "Freelance", Type: domain.CashflowIncome, IsDefault: true},
	{Name: "Hadiah", Type: domain.CashflowIncome, IsDefault: true},
	{Name: "Lainnya", Type: domain.CashflowIncome, IsDefault: true},
	{Name: "Makanan & Minuman", Type: domain.CashflowExpense, IsDefault: true},
	{Name: "Transportasi", Type: domain.CashflowExpense, IsDefault: true},
	{Name: "Belanja", Type: domain.CashflowExpense, IsDefault: true},
	{Name: "Tagihan & Utilitas", Type: domain.CashflowExpense, IsDefault: true},
	{Name: "Kesehatan", Type: domain.CashflowExpense, IsDefault: true},
	{Name: "Hiburan", Type: domain.CashflowExpense, IsDefault: true},
	{Name: "Pendidikan", Type: domain.CashflowExpense, IsDefault: true},
	{Name: "Liburan", Type: domain.CashflowExpense, IsDefault: true},
	{Name: "Lainnya", Type: domain.CashflowExpense, IsDefault: true},
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type AuthService struct {
	users  domain.UserRepository
	tokens *TokenManager
}

func NewAuthService(users domain.UserRepository, tokens *TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

// Register creates a user together with the default categories.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (domain.User, error) {
	name := strings.TrimSpace(in.Name)
	email := normalizeEmail(in.Email)

	fields := map[string]string{}
	if utf8.RuneCountInString(name) < minNameLength {
		fields["name"] = fmt.Sprintf("must be at least %d characters", minNameLength)
	}
	if !isValidEmail(email) {
		fields["email"] = "must be a valid email address"
	}
	if utf8.RuneCountInString(in.Password) < minPasswordLength {
		fields["password"] = fmt.Sprintf("must be at least %d characters", minPasswordLength)
	} else if len(in.Password) > maxPasswordBytes {
		fields["password"] = fmt.Sprintf("must be at most %d bytes", maxPasswordBytes)
	}
	if len(fields) > 0 {
		return domain.User{}, &domain.ValidationError{Fields: fields}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.CreateWithCategories(ctx, domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}, defaultCategories)
	if err != nil {
		return domain.User{}, fmt.Errorf("register user: %w", err)
	}
	return user, nil
}

// Login checks the credentials and returns the user with a new access token.
func (s *AuthService) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	email = normalizeEmail(email)

	fields := map[string]string{}
	if email == "" {
		fields["email"] = "is required"
	}
	if password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		return domain.User{}, "", &domain.ValidationError{Fields: fields}
	}

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, "", fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, "", domain.ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(user.ID)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("issue token: %w", err)
	}
	return user, token, nil
}

// Me returns the profile of the authenticated user.
func (s *AuthService) Me(ctx context.Context, userID string) (domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// isValidEmail accepts a bare address only, rejecting forms like "Name <a@b.c>".
func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
