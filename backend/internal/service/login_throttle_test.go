package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

// fakeClock is a time source that tests move forward by hand.
type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)}
}

func TestLoginThrottle_LocksAfterMaxFailuresUntilTheWindowEnds(t *testing.T) {
	clock := newFakeClock()
	throttle := NewLoginThrottle(clock.Now)

	for i := 1; i < maxLoginFailures; i++ {
		throttle.Fail("a@example.com")
		if wait := throttle.RetryAfter("a@example.com"); wait != 0 {
			t.Fatalf("after %d failures RetryAfter = %v, want 0", i, wait)
		}
	}

	clock.Advance(5 * time.Minute)
	throttle.Fail("a@example.com") // the maxLoginFailures-th failure locks
	if wait := throttle.RetryAfter("a@example.com"); wait != 10*time.Minute {
		t.Fatalf("RetryAfter = %v, want 10m (window started 5m ago)", wait)
	}
	if wait := throttle.RetryAfter("b@example.com"); wait != 0 {
		t.Errorf("another email RetryAfter = %v, want 0", wait)
	}

	clock.Advance(10 * time.Minute)
	if wait := throttle.RetryAfter("a@example.com"); wait != 0 {
		t.Errorf("after the window RetryAfter = %v, want 0", wait)
	}
}

func TestLoginThrottle_ResetClearsFailures(t *testing.T) {
	throttle := NewLoginThrottle(newFakeClock().Now)
	for i := 0; i < maxLoginFailures; i++ {
		throttle.Fail("a@example.com")
	}
	throttle.Reset("a@example.com")

	if wait := throttle.RetryAfter("a@example.com"); wait != 0 {
		t.Errorf("RetryAfter after Reset = %v, want 0", wait)
	}
}

func TestLoginThrottle_PrunesExpiredWindows(t *testing.T) {
	clock := newFakeClock()
	throttle := NewLoginThrottle(clock.Now)
	throttle.Fail("old@example.com")

	clock.Advance(loginFailureWindow)
	throttle.Fail("new@example.com")

	if _, kept := throttle.windows["old@example.com"]; kept {
		t.Error("expired window was not pruned")
	}
	if len(throttle.windows) != 1 {
		t.Errorf("windows = %d, want 1", len(throttle.windows))
	}
}

func TestLogin_LocksEmailAfterRepeatedFailures(t *testing.T) {
	repo := newFakeUserRepository()
	clock := newFakeClock()
	svc := NewAuthService(repo, NewTokenManager("test-secret-with-at-least-32-chars!", time.Hour), NewLoginThrottle(clock.Now))
	ctx := context.Background()
	if _, err := svc.Register(ctx, RegisterInput{Name: "Arya", Email: "arya@example.com", Password: "rahasia123"}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	for i := 0; i < maxLoginFailures; i++ {
		// Mixed case and spaces must count for the same email.
		if _, _, err := svc.Login(ctx, " ARYA@example.com ", "salah-salah"); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("failed attempt %d: error = %v, want ErrInvalidCredentials", i+1, err)
		}
	}

	// Locked: even the right password is refused.
	_, _, err := svc.Login(ctx, "arya@example.com", "rahasia123")
	var tooMany *domain.TooManyAttemptsError
	if !errors.As(err, &tooMany) || tooMany.RetryAfter != loginFailureWindow {
		t.Fatalf("login while locked: error = %v, want TooManyAttemptsError with %v", err, loginFailureWindow)
	}

	clock.Advance(loginFailureWindow)
	if _, _, err := svc.Login(ctx, "arya@example.com", "rahasia123"); err != nil {
		t.Fatalf("login after the lock: error = %v", err)
	}
}

func TestLogin_UnknownEmailsAreLimitedTheSameWay(t *testing.T) {
	svc := NewAuthService(newFakeUserRepository(), NewTokenManager("test-secret-with-at-least-32-chars!", time.Hour), NewLoginThrottle(newFakeClock().Now))

	for i := 0; i < maxLoginFailures; i++ {
		if _, _, err := svc.Login(context.Background(), "ghost@example.com", "x"); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("attempt %d: error = %v, want ErrInvalidCredentials", i+1, err)
		}
	}
	var tooMany *domain.TooManyAttemptsError
	if _, _, err := svc.Login(context.Background(), "ghost@example.com", "x"); !errors.As(err, &tooMany) {
		t.Fatalf("error = %v, want TooManyAttemptsError (same as for a real account)", err)
	}
}

func TestLogin_SuccessResetsFailures(t *testing.T) {
	svc, _, _ := newTestAuthService()
	ctx := context.Background()
	if _, err := svc.Register(ctx, RegisterInput{Name: "Arya", Email: "arya@example.com", Password: "rahasia123"}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	for i := 0; i < maxLoginFailures-1; i++ {
		_, _, _ = svc.Login(ctx, "arya@example.com", "salah-salah")
	}
	if _, _, err := svc.Login(ctx, "arya@example.com", "rahasia123"); err != nil {
		t.Fatalf("correct login: error = %v", err)
	}
	// The counter started over, so another round of failures does not lock yet.
	for i := 0; i < maxLoginFailures-1; i++ {
		if _, _, err := svc.Login(ctx, "arya@example.com", "salah-salah"); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("failure %d after reset: error = %v, want ErrInvalidCredentials", i+1, err)
		}
	}
}
