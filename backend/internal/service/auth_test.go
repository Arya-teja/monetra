package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

// fakeUserRepository is an in-memory domain.UserRepository.
type fakeUserRepository struct {
	users      map[string]domain.User // keyed by email
	categories []domain.Category
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{users: map[string]domain.User{}}
}

func (r *fakeUserRepository) CreateWithCategories(_ context.Context, user domain.User, categories []domain.Category) (domain.User, error) {
	if _, ok := r.users[user.Email]; ok {
		return domain.User{}, domain.ErrEmailTaken
	}
	user.ID = "user-" + user.Email
	r.users[user.Email] = user
	r.categories = append(r.categories, categories...)
	return user, nil
}

func (r *fakeUserRepository) FindByEmail(_ context.Context, email string) (domain.User, error) {
	user, ok := r.users[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, id string) (domain.User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

func (r *fakeUserRepository) TokenVersion(ctx context.Context, id string) (int, error) {
	user, err := r.FindByID(ctx, id)
	return user.TokenVersion, err
}

func (r *fakeUserRepository) IncrementTokenVersion(ctx context.Context, id string) error {
	user, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	user.TokenVersion++
	r.users[user.Email] = user
	return nil
}

func newTestAuthService() (*AuthService, *fakeUserRepository, *TokenManager) {
	repo := newFakeUserRepository()
	tokens := NewTokenManager("test-secret-with-at-least-32-chars!", time.Hour)
	return NewAuthService(repo, tokens), repo, tokens
}

func TestRegister_CreatesUserWithHashedPasswordAndDefaultCategories(t *testing.T) {
	svc, repo, _ := newTestAuthService()

	user, err := svc.Register(context.Background(), RegisterInput{
		Name:     "  Arya  ",
		Email:    "  Arya@Example.COM ",
		Password: "rahasia123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if user.Name != "Arya" {
		t.Errorf("name = %q, want %q", user.Name, "Arya")
	}
	if user.Email != "arya@example.com" {
		t.Errorf("email = %q, want %q", user.Email, "arya@example.com")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("rahasia123")); err != nil {
		t.Errorf("stored password hash does not match the password: %v", err)
	}

	var income, expense int
	for _, c := range repo.categories {
		if !c.IsDefault {
			t.Errorf("category %q is not marked as default", c.Name)
		}
		switch c.Type {
		case domain.CashflowIncome:
			income++
		case domain.CashflowExpense:
			expense++
		}
	}
	if income != 5 || expense != 9 {
		t.Errorf("default categories = %d income, %d expense; want 5 income, 9 expense", income, expense)
	}
}

func TestRegister_RejectsInvalidInput(t *testing.T) {
	svc, _, _ := newTestAuthService()

	_, err := svc.Register(context.Background(), RegisterInput{
		Name:     "A",
		Email:    "bukan-email",
		Password: "pendek",
	})

	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Register() error = %v, want *domain.ValidationError", err)
	}
	for _, field := range []string{"name", "email", "password"} {
		if _, ok := validationErr.Fields[field]; !ok {
			t.Errorf("missing validation error for %q", field)
		}
	}
}

func TestRegister_RejectsPasswordLongerThanBcryptLimit(t *testing.T) {
	svc, _, _ := newTestAuthService()

	long := make([]byte, maxPasswordBytes+1)
	for i := range long {
		long[i] = 'a'
	}

	_, err := svc.Register(context.Background(), RegisterInput{
		Name:     "Arya",
		Email:    "arya@example.com",
		Password: string(long),
	})

	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) || validationErr.Fields["password"] == "" {
		t.Fatalf("Register() error = %v, want password validation error", err)
	}
}

func TestRegister_RejectsDuplicateEmail(t *testing.T) {
	svc, _, _ := newTestAuthService()
	in := RegisterInput{Name: "Arya", Email: "arya@example.com", Password: "rahasia123"}

	if _, err := svc.Register(context.Background(), in); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	in.Email = "ARYA@example.com"
	if _, err := svc.Register(context.Background(), in); !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("second Register() error = %v, want ErrEmailTaken", err)
	}
}

func TestLogin_ReturnsTokenForValidCredentials(t *testing.T) {
	svc, _, _ := newTestAuthService()
	registered, err := svc.Register(context.Background(), RegisterInput{
		Name: "Arya", Email: "arya@example.com", Password: "rahasia123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, token, err := svc.Login(context.Background(), " ARYA@example.com ", "rahasia123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user.ID != registered.ID {
		t.Errorf("user ID = %q, want %q", user.ID, registered.ID)
	}

	userID, err := svc.Authenticate(context.Background(), token.Value)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if userID != registered.ID {
		t.Errorf("token subject = %q, want %q", userID, registered.ID)
	}
}

func TestLogin_RejectsWrongCredentials(t *testing.T) {
	svc, _, _ := newTestAuthService()
	if _, err := svc.Register(context.Background(), RegisterInput{
		Name: "Arya", Email: "arya@example.com", Password: "rahasia123",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{name: "wrong password", email: "arya@example.com", password: "salah12345"},
		{name: "unknown email", email: "orang@example.com", password: "rahasia123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := svc.Login(context.Background(), tt.email, tt.password)
			if !errors.Is(err, domain.ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
			}
		})
	}
}

func TestLogin_RequiresEmailAndPassword(t *testing.T) {
	svc, _, _ := newTestAuthService()

	_, _, err := svc.Login(context.Background(), " ", "")

	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Login() error = %v, want *domain.ValidationError", err)
	}
	if validationErr.Fields["email"] == "" || validationErr.Fields["password"] == "" {
		t.Errorf("fields = %v, want email and password errors", validationErr.Fields)
	}
}

func TestLogout_RevokesEarlierTokensOnAllDevices(t *testing.T) {
	svc, _, _ := newTestAuthService()
	ctx := context.Background()
	if _, err := svc.Register(ctx, RegisterInput{Name: "Arya", Email: "arya@example.com", Password: "rahasia123"}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	login := func() string {
		t.Helper()
		_, token, err := svc.Login(ctx, "arya@example.com", "rahasia123")
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}
		return token.Value
	}
	phone, laptop := login(), login()

	userID, err := svc.Authenticate(ctx, phone)
	if err != nil {
		t.Fatalf("Authenticate() before logout error = %v", err)
	}
	if err := svc.Logout(ctx, userID); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	for name, token := range map[string]string{"phone": phone, "laptop": laptop} {
		if _, err := svc.Authenticate(ctx, token); !errors.Is(err, domain.ErrInvalidToken) {
			t.Errorf("%s token after logout: error = %v, want ErrInvalidToken", name, err)
		}
	}
	if _, err := svc.Authenticate(ctx, login()); err != nil {
		t.Errorf("token from a new login: error = %v, want nil", err)
	}
}

func TestAuthenticate_RejectsInvalidTokens(t *testing.T) {
	svc, _, tokens := newTestAuthService()
	ghost, err := tokens.Issue("no-such-user", 0)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	for name, token := range map[string]string{"garbage": "not.a.jwt", "unknown user": ghost.Value} {
		if _, err := svc.Authenticate(context.Background(), token); !errors.Is(err, domain.ErrInvalidToken) {
			t.Errorf("%s: error = %v, want ErrInvalidToken", name, err)
		}
	}
}
