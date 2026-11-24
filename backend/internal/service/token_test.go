package service

import (
	"testing"
	"time"
)

const testSecret = "test-secret-with-at-least-32-chars!"

func TestTokenManager_IssueAndVerify(t *testing.T) {
	tokens := NewTokenManager(testSecret, time.Hour)

	token, err := tokens.Issue("user-1")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	userID, err := tokens.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if userID != "user-1" {
		t.Errorf("Verify() = %q, want %q", userID, "user-1")
	}
}

func TestTokenManager_RejectsExpiredToken(t *testing.T) {
	tokens := NewTokenManager(testSecret, -time.Minute)

	token, err := tokens.Issue("user-1")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := tokens.Verify(token); err == nil {
		t.Fatal("Verify() accepted an expired token")
	}
}

func TestTokenManager_RejectsTokenSignedWithAnotherSecret(t *testing.T) {
	other := NewTokenManager("another-secret-with-at-least-32-chars", time.Hour)
	token, err := other.Issue("user-1")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := NewTokenManager(testSecret, time.Hour).Verify(token); err == nil {
		t.Fatal("Verify() accepted a token signed with another secret")
	}
}
