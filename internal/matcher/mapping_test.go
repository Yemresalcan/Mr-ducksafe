package matcher

import (
	"context"
	"strings"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
)

func TestNewDefaultMatcher_Loads(t *testing.T) {
	m, err := NewDefaultMatcher()
	if err != nil {
		t.Fatalf("NewDefaultMatcher: %v", err)
	}
	if len(m.entries) < 10 {
		t.Errorf("en az 10 mapping bekleniyor, alındı: %d", len(m.entries))
	}
}

func TestMappingMatcher_Match(t *testing.T) {
	m, err := NewDefaultMatcher()
	if err != nil {
		t.Fatalf("NewDefaultMatcher: %v", err)
	}

	cases := []struct {
		name      string
		app       scanner.InstalledApp
		wantType  IdentifierType
		wantConf  MatchConfidence
		wantSubst string // wantValue'nun içerebileceği alt string
	}{
		{
			name:      "Chrome exact",
			app:       scanner.InstalledApp{Name: "Google Chrome", Version: "120.0.6099.130", Publisher: "Google LLC"},
			wantType:  IdentifierTypeCPE,
			wantConf:  ConfidenceHigh,
			wantSubst: "google:chrome",
		},
		{
			name:      "Firefox pattern",
			app:       scanner.InstalledApp{Name: "Mozilla Firefox 121.0 (x64 en-US)", Version: "121.0", Publisher: "Mozilla"},
			wantType:  IdentifierTypeCPE,
			wantConf:  ConfidenceHigh,
			wantSubst: "mozilla:firefox",
		},
		{
			name:     "Bilinmeyen app",
			app:      scanner.InstalledApp{Name: "Some Random Tool", Version: "1.0", Publisher: "Nobody"},
			wantConf: ConfidenceNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, conf, err := m.Match(context.Background(), tc.app)
			if err != nil {
				t.Fatalf("Match: %v", err)
			}
			if conf != tc.wantConf {
				t.Errorf("confidence = %v, want %v", conf, tc.wantConf)
			}
			if tc.wantConf == ConfidenceNone {
				return
			}
			if id.Type != tc.wantType {
				t.Errorf("type = %v, want %v", id.Type, tc.wantType)
			}
			if tc.wantSubst != "" && !strings.Contains(id.Value, tc.wantSubst) {
				t.Errorf("value %q içermiyor %q", id.Value, tc.wantSubst)
			}
		})
	}
}
