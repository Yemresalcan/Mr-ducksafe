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
	if len(m.entries) < 50 {
		t.Errorf("en az 50 mapping bekleniyor, alındı: %d", len(m.entries))
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

func TestMappingMatcher_VersionFromName(t *testing.T) {
	m, err := NewDefaultMatcher()
	if err != nil {
		t.Fatalf("NewDefaultMatcher: %v", err)
	}

	// IntelliJ DisplayName'inde marketing versiyonu var; Registry version'u
	// build numarası (252.x). version_from_name isimden 2025.2.2 çıkarmalı.
	app := scanner.InstalledApp{
		Name:      "IntelliJ IDEA 2025.2.2",
		Version:   "252.26199.169",
		Publisher: "JetBrains s.r.o.",
	}
	id, conf, err := m.Match(context.Background(), app)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if conf != ConfidenceHigh {
		t.Fatalf("confidence = %v, want high", conf)
	}
	if id.Version != "2025.2.2" {
		t.Errorf("resolved version = %q, want 2025.2.2 (isimden)", id.Version)
	}
}

func TestMappingMatcher_RemediationFields(t *testing.T) {
	m, _ := NewDefaultMatcher()
	app := scanner.InstalledApp{Name: "Google Chrome", Version: "120.0.6099.130", Publisher: "Google LLC"}
	id, _, _ := m.Match(context.Background(), app)
	if id.WingetID != "Google.Chrome" {
		t.Errorf("WingetID = %q, want Google.Chrome", id.WingetID)
	}
	if id.UpdateURL == "" {
		t.Error("UpdateURL boş olmamalı")
	}
}

func TestMappingMatcher_VersionFallsBackToAppVersion(t *testing.T) {
	m, _ := NewDefaultMatcher()
	// Chrome'da version_from_name yok → app.Version kullanılmalı.
	app := scanner.InstalledApp{Name: "Google Chrome", Version: "120.0.6099.130", Publisher: "Google LLC"}
	id, _, _ := m.Match(context.Background(), app)
	if id.Version != "120.0.6099.130" {
		t.Errorf("resolved version = %q, want 120.0.6099.130 (app.Version)", id.Version)
	}
}
