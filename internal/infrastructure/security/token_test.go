package security

import (
	"testing"
	"time"
)

func TestTokenSignVerify(t *testing.T) {
	tm := NewTokenManager("secret")
	tok, err := tm.Sign("admin", "access", time.Now().Add(1*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tm.Verify(tok, "access")
	if err != nil {
		t.Fatal(err)
	}
	if claims.Sub != "admin" {
		t.Fatalf("unexpected sub: %s", claims.Sub)
	}
}
