package crypto

import (
	"strings"
	"testing"
)

func TestHashPasswordUsesSingleArgon2Lane(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !strings.Contains(hash, ",p=1$") {
		t.Fatalf("HashPassword() encoded parallelism = %q, want p=1", hash)
	}
}

func TestVerifyPasswordAcceptsGeneratedHash(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	ok, err := VerifyPassword("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !ok {
		t.Fatal("VerifyPassword() = false, want true")
	}
}
