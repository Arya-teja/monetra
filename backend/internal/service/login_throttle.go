package service

import (
	"sync"
	"time"
)

const (
	maxLoginFailures   = 5
	loginFailureWindow = 15 * time.Minute
)

// LoginThrottle limits failed logins per email: after maxLoginFailures failures
// within loginFailureWindow, the email is locked until that window ends.
//
// Counting is per email, not per IP: logins reach the API from the frontend server
// (NextAuth runs on Vercel), so every user shares the same IP.
//
// The counts live in memory, which is enough for one API instance; they reset when
// the server restarts.
type LoginThrottle struct {
	mu      sync.Mutex
	now     func() time.Time
	windows map[string]failureWindow // keyed by normalized email
}

type failureWindow struct {
	start    time.Time
	failures int
}

func NewLoginThrottle(now func() time.Time) *LoginThrottle {
	return &LoginThrottle{now: now, windows: map[string]failureWindow{}}
}

// RetryAfter returns how long the email is still locked, or 0 when it may try to log in.
func (t *LoginThrottle) RetryAfter(email string) time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	w, ok := t.windows[email]
	if !ok {
		return 0
	}
	remaining := w.start.Add(loginFailureWindow).Sub(t.now())
	if remaining <= 0 {
		delete(t.windows, email)
		return 0
	}
	if w.failures < maxLoginFailures {
		return 0
	}
	return remaining
}

// Fail records a failed login for the email.
func (t *LoginThrottle) Fail(email string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	t.pruneExpired(now)

	w, ok := t.windows[email]
	if !ok {
		w = failureWindow{start: now}
	}
	w.failures++
	t.windows[email] = w
}

// Reset forgets the failures of the email, after a successful login.
func (t *LoginThrottle) Reset(email string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.windows, email)
}

// pruneExpired drops finished windows so memory does not grow with every email
// ever tried. The caller holds the lock.
func (t *LoginThrottle) pruneExpired(now time.Time) {
	for email, w := range t.windows {
		if !now.Before(w.start.Add(loginFailureWindow)) {
			delete(t.windows, email)
		}
	}
}
