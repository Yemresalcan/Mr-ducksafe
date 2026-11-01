package i18n

import "testing"

func TestEveryMessageHasBothLanguages(t *testing.T) {
	for key, m := range messages {
		if m[EN] == "" {
			t.Errorf("%q has no English text", key)
		}
		if m[TR] == "" {
			t.Errorf("%q has no Turkish text", key)
		}
	}
}

func TestTAndFallbacks(t *testing.T) {
	defer Set(EN)

	Set(EN)
	if got := T("sev.critical"); got != "CRIT" {
		t.Errorf("EN sev.critical = %q", got)
	}
	Set(TR)
	if got := T("sev.critical"); got != "KRIT" {
		t.Errorf("TR sev.critical = %q", got)
	}
	if got := Tf("count.high", 3); got != "3 yüksek" {
		t.Errorf("TR count.high = %q", got)
	}
	if got := T("no.such.key"); got != "no.such.key" {
		t.Errorf("unknown key should echo itself, got %q", got)
	}
	Set("de")
	if Current() != EN {
		t.Errorf("unknown language should fall back to EN, got %q", Current())
	}
}

func TestParse(t *testing.T) {
	for in, want := range map[string]Lang{"1": EN, "en": EN, " English ": EN, "2": TR, "tr": TR, "Türkçe": TR} {
		got, ok := Parse(in)
		if !ok || got != want {
			t.Errorf("Parse(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	if _, ok := Parse("de"); ok {
		t.Error("Parse(de) should not be ok")
	}
}
