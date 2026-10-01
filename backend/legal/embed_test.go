package legaltext

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCurrent_ParsesEmbeddedFiles(t *testing.T) {
	n := Current()
	if n.Version != CurrentVersion {
		t.Errorf("Version = %q, want %q", n.Version, CurrentVersion)
	}
	if len(n.Simplified) == 0 {
		t.Error("expected at least one simplified section")
	}
	if len(n.Full) == 0 {
		t.Error("expected at least one full section")
	}
	if n.Checksum == "" {
		t.Error("expected a non-empty checksum")
	}
}

func TestCurrent_ChecksumIsStable(t *testing.T) {
	a := Current()
	b := Current()
	if a.Checksum != b.Checksum {
		t.Errorf("checksum changed between calls: %q vs %q", a.Checksum, b.Checksum)
	}
}

func TestVerifyVersion(t *testing.T) {
	if !VerifyVersion(CurrentVersion) {
		t.Error("expected the current version to verify")
	}
	if VerifyVersion("") {
		t.Error("expected an empty version to fail verification")
	}
	if VerifyVersion("v1.0-preliminar") {
		t.Error("expected a stale/unknown version to fail verification")
	}
}

func TestSealPayload_BindsAccountVersionChecksumAndTime(t *testing.T) {
	accountID := uuid.New()
	acceptedAt := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)

	payload := SealPayload(accountID, acceptedAt)
	got := string(payload)

	n := Current()
	want := accountID.String() + "|" + n.Version + "|" + n.Checksum + "|" + acceptedAt.Format(time.RFC3339Nano)
	if got != want {
		t.Errorf("SealPayload = %q, want %q", got, want)
	}

	// Distintos accountID o acceptedAt deben producir payloads distintos —
	// si no, el HMAC resultante colisionaría entre cuentas o fechas.
	other := SealPayload(uuid.New(), acceptedAt)
	if string(other) == got {
		t.Error("expected different accountID to produce a different payload")
	}
	otherTime := SealPayload(accountID, acceptedAt.Add(time.Second))
	if string(otherTime) == got {
		t.Error("expected different acceptedAt to produce a different payload")
	}
}
