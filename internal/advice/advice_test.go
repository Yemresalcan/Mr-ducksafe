package advice

import (
	"strings"
	"testing"
)

func TestUrgency(t *testing.T) {
	cases := map[string]string{
		"critical": "Acilen",
		"high":     "Yakında",
		"medium":   "Fırsat",
		"low":      "Düşük",
		"bilinmez": "önerilir",
	}
	for sev, want := range cases {
		if got := Urgency(sev); !strings.Contains(got, want) {
			t.Errorf("Urgency(%q) = %q, %q içermeli", sev, got, want)
		}
	}
}

func TestRiskLine_KeywordMatch(t *testing.T) {
	cases := []struct {
		name      string
		summaries []string
		wantSub   string
	}{
		{"rce", []string{"Allows remote code execution via crafted input"}, "Uzaktan"},
		{"xss", []string{"Cross-site scripting in the editor"}, "script"},
		{"memory", []string{"Use-after-free in XSLT parameter processing"}, "Bellek"},
		{"priv-esc", []string{"Local privilege escalation"}, "yetkilerini"},
		{"dos", []string{"Denial of service via malformed packet"}, "çökertilerek"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RiskLine("high", tc.summaries)
			if !strings.Contains(got, tc.wantSub) {
				t.Errorf("RiskLine = %q, %q içermeli", got, tc.wantSub)
			}
		})
	}
}

func TestRiskLine_FallbackBySeverity(t *testing.T) {
	got := RiskLine("critical", []string{"Some obscure issue with no keyword"})
	if !strings.Contains(got, "ciddi") {
		t.Errorf("fallback critical = %q, 'ciddi' içermeli", got)
	}
}
