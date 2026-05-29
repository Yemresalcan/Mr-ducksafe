package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestBanner_ContainsName(t *testing.T) {
	color.NoColor = true
	b := Banner("v0.1.0")
	if !strings.Contains(b, "mr-ducksafe") {
		t.Error("banner 'mr-ducksafe' içermeli")
	}
	if !strings.Contains(b, "v0.1.0") {
		t.Error("banner sürümü içermeli")
	}
}

func TestProgress_CompleteWritesCheckmark(t *testing.T) {
	color.NoColor = true
	var buf bytes.Buffer
	p := NewProgress(&buf, true)
	p.Complete("Tarama tamamlandı")
	out := buf.String()
	if !strings.Contains(out, "✓") {
		t.Errorf("✓ bekleniyor, alındı: %q", out)
	}
	if !strings.Contains(out, "Tarama tamamlandı") {
		t.Errorf("mesaj bekleniyor, alındı: %q", out)
	}
}

func TestProgress_DisabledIsNoOp(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress(&buf, false)
	p.Start()
	p.Update("bir şey")
	p.Complete("tamam")
	p.Stop()
	if buf.Len() != 0 {
		t.Errorf("disabled modda çıktı olmamalı, alındı: %q", buf.String())
	}
}
