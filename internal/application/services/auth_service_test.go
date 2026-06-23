package services

import (
	"backend-of/internal/infrastructure/db/jsondb"
	"backend-of/internal/infrastructure/security"
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestLoginWithGoogleCreatesAndReusesUser(t *testing.T) {
	store, err := jsondb.NewStore(filepath.Join(t.TempDir(), "data.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	svc := NewAuthService(store, security.NewTokenManager("test-secret"), 15*time.Minute, time.Hour)

	acc1, ref1, err := svc.LoginWithGoogle(context.Background(), "google-123", "sally@email.com", "Sally Paula Ortiz", "https://example.com/avatar.png")
	if err != nil {
		t.Fatalf("first google login: %v", err)
	}
	if acc1 == "" || ref1 == "" {
		t.Fatalf("expected session tokens")
	}

	user, ok, err := store.GetByGoogleID(context.Background(), "google-123")
	if err != nil || !ok {
		t.Fatalf("expected user stored by google id")
	}
	if user.Role != "user" || user.Username == "" || user.Email != "sally@email.com" {
		t.Fatalf("unexpected stored user: %+v", user)
	}

	acc2, ref2, err := svc.LoginWithGoogle(context.Background(), "google-123", "sally@email.com", "Sally Paula Ortiz", "https://example.com/avatar.png")
	if err != nil {
		t.Fatalf("second google login: %v", err)
	}
	if acc2 == "" || ref2 == "" {
		t.Fatalf("expected session tokens on reuse")
	}
}
