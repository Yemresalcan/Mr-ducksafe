package osv

import "testing"

func TestSeverity_Values(t *testing.T) {
	all := []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityUnknown}
	seen := map[Severity]bool{}
	for _, s := range all {
		if s == "" {
			t.Errorf("severity boş olamaz")
		}
		if seen[s] {
			t.Errorf("duplicate severity: %s", s)
		}
		seen[s] = true
	}
}
