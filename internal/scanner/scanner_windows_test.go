//go:build windows

package scanner

import (
	"context"
	"testing"
	"time"
)

func TestWindowsScanner_Smoke(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apps, err := s.Scan(ctx)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(apps) == 0 {
		t.Skip("Bu makinede Registry'de hiç entry bulunamadı, smoke test atlanıyor")
	}

	for _, a := range apps {
		if a.Name == "" {
			t.Errorf("Name boş bir app döndü: %+v", a)
		}
		if a.Source != sourceWindowsRegistry {
			t.Errorf("beklenmedik Source: %s", a.Source)
		}
	}
	t.Logf("Registry'de %d kurulu yazılım bulundu", len(apps))
}
