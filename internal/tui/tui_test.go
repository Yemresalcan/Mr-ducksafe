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

func TestWordmark(t *testing.T) {
	lines := Wordmark("MR.DUCKSAFE")
	if len(lines) != 3 {
		t.Fatalf("want 3 rows, got %d", len(lines))
	}
	for _, l := range lines {
		if w := len([]rune(l)); w > 72 {
			t.Errorf("wordmark row is %d cols wide, must fit an 80-col console", w)
		}
		if strings.ContainsAny(l, "#.") {
			t.Errorf("raw bitmap characters leaked: %q", l)
		}
	}
}

func TestQuestionMarksDefault(t *testing.T) {
	q := Question("Pick one", []string{"First", "Second"}, "(default)")
	for _, want := range []string{"Pick one", "1)", "First", "(default)", "2)", "Second"} {
		if !strings.Contains(q, want) {
			t.Errorf("question missing %q: %q", want, q)
		}
	}
}
