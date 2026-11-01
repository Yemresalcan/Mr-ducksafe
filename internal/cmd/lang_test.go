package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
)

// withLangEnv points the saved-language file at a temp dir and controls
// whether the session looks interactive.
func withLangEnv(t *testing.T, interactive bool) string {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "lang")
	oldFile, oldInter := langFile, isInteractive
	langFile = func() (string, error) { return file, nil }
	isInteractive = func() bool { return interactive }
	t.Setenv(envLang, "")
	t.Cleanup(func() { langFile, isInteractive = oldFile, oldInter; i18n.Set(i18n.EN) })
	return file
}

func TestResolveLanguage_NonInteractiveDefaultsToEnglishWithoutPrompt(t *testing.T) {
	file := withLangEnv(t, false)
	var out bytes.Buffer
	if got := resolveLanguage([]string{"scan"}, strings.NewReader("2\n"), &out); got != i18n.EN {
		t.Errorf("got %q, want en", got)
	}
	if out.Len() != 0 {
		t.Errorf("must not prompt in CI/pipes, printed %q", out.String())
	}
	if _, err := os.Stat(file); err == nil {
		t.Error("must not save a choice when nobody was asked")
	}
}

func TestResolveLanguage_FirstRunPromptsAndRemembers(t *testing.T) {
	file := withLangEnv(t, true)
	var out bytes.Buffer
	if got := resolveLanguage(nil, strings.NewReader("2\n"), &out); got != i18n.TR {
		t.Fatalf("got %q, want tr", got)
	}
	if !strings.Contains(out.String(), "English") || !strings.Contains(out.String(), "Türkçe") {
		t.Errorf("prompt should offer both languages, got %q", out.String())
	}
	b, err := os.ReadFile(file)
	if err != nil || strings.TrimSpace(string(b)) != "tr" {
		t.Fatalf("choice not saved: %q, %v", b, err)
	}
	// Second run: no prompt, saved choice wins.
	out.Reset()
	if got := resolveLanguage(nil, strings.NewReader(""), &out); got != i18n.TR || out.Len() != 0 {
		t.Errorf("second run got %q and printed %q", got, out.String())
	}
}

func TestResolveLanguage_EmptyAnswerMeansEnglish(t *testing.T) {
	withLangEnv(t, true)
	if got := resolveLanguage(nil, strings.NewReader("\n"), &bytes.Buffer{}); got != i18n.EN {
		t.Errorf("got %q, want en", got)
	}
}

func TestResolveLanguage_FlagAndEnvOverrideSavedChoice(t *testing.T) {
	file := withLangEnv(t, true)
	if err := os.WriteFile(file, []byte("tr\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := resolveLanguage([]string{"scan", "--lang", "en"}, strings.NewReader(""), &bytes.Buffer{}); got != i18n.EN {
		t.Errorf("--lang en: got %q", got)
	}
	if got := resolveLanguage([]string{"--lang=en", "history"}, strings.NewReader(""), &bytes.Buffer{}); got != i18n.EN {
		t.Errorf("--lang=en: got %q", got)
	}
	t.Setenv(envLang, "en")
	if got := resolveLanguage(nil, strings.NewReader(""), &bytes.Buffer{}); got != i18n.EN {
		t.Errorf("%s=en: got %q", envLang, got)
	}
}

func TestResolveLanguage_LangCommandIsNotInterrupted(t *testing.T) {
	withLangEnv(t, true)
	var out bytes.Buffer
	resolveLanguage([]string{"lang", "tr"}, strings.NewReader("1\n"), &out)
	if out.Len() != 0 {
		t.Errorf("`lang` must not trigger the prompt, printed %q", out.String())
	}
}

func TestLangCommandSavesChoice(t *testing.T) {
	file := withLangEnv(t, true)
	c := NewLangCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetArgs([]string{"tr"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(file); strings.TrimSpace(string(b)) != "tr" {
		t.Errorf("saved %q, want tr", b)
	}
	c.SetArgs([]string{"de"})
	if err := c.Execute(); err == nil {
		t.Error("unknown language should fail")
	}
}
