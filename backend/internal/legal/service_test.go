package legal

import (
	"testing"

	legaltext "github.com/altair/usbi-anon-backend/legal"
)

func TestCurrentNotice_MapsEmbeddedContent(t *testing.T) {
	svc := NewService(nil, []byte("test-secret"))

	resp := svc.CurrentNotice()

	if resp.Version != legaltext.CurrentVersion {
		t.Errorf("Version = %q, want %q", resp.Version, legaltext.CurrentVersion)
	}
	if resp.EffectiveDate != legaltext.EffectiveDate {
		t.Errorf("EffectiveDate = %q, want %q", resp.EffectiveDate, legaltext.EffectiveDate)
	}
	if len(resp.Simplified) == 0 {
		t.Error("expected at least one simplified section")
	}
	if len(resp.Full) == 0 {
		t.Error("expected at least one full section")
	}
	if resp.Checksum == "" {
		t.Error("expected a non-empty checksum")
	}
	if resp.Checksum != legaltext.Current().Checksum {
		t.Error("checksum should match legaltext.Current()'s checksum exactly")
	}
}
