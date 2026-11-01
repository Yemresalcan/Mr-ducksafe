package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
	"github.com/yemresalcan/mr-ducksafe/internal/tui"
)

// envLang overrides the saved language for a single run.
const envLang = "MR_DUCKSAFE_LANG"

var flagLang string

// langFile is where the chosen language is remembered (~/.mr-ducksafe/lang).
var langFile = func() (string, error) {
	dir, err := config.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "lang"), nil
}

// welcomed is set when the first-run splash was printed, so the regular
// banner is not shown a second time in the same run.
var welcomed bool

// welcomeText is the first-run splash (overridable in tests).
var welcomeText = func() string {
	return tui.Welcome(Version, "CVE scanner for installed Windows software")
}

// isInteractive reports whether a person can answer the language prompt.
// CI, pipes and redirected input never get a prompt.
var isInteractive = func() bool {
	return tui.IsTTY(os.Stdin) && tui.IsTTY(os.Stderr)
}

// resolveLanguage picks the output language before any command is built:
// --lang flag > MR_DUCKSAFE_LANG > saved choice > first-run prompt > English.
func resolveLanguage(args []string, in io.Reader, out io.Writer) i18n.Lang {
	if v, ok := langFromArgs(args); ok {
		if l, ok := i18n.Parse(v); ok {
			return l
		}
	}
	if v := os.Getenv(envLang); v != "" {
		if l, ok := i18n.Parse(v); ok {
			return l
		}
	}
	if l, ok := loadSavedLang(); ok {
		return l
	}
	if isInteractive() && !isLangCommand(args) {
		l := promptLang(in, out)
		_ = saveLang(l)
		i18n.Set(l)
		fmt.Fprintln(out, i18n.T("lang.saved"))
		fmt.Fprintln(out)
		return l
	}
	return i18n.EN
}

// langFromArgs finds "--lang X" or "--lang=X" without needing cobra.
func langFromArgs(args []string) (string, bool) {
	for i, a := range args {
		if a == "--" {
			break
		}
		if v, ok := strings.CutPrefix(a, "--lang="); ok {
			return v, true
		}
		if a == "--lang" && i+1 < len(args) {
			return args[i+1], true
		}
	}
	return "", false
}

// isLangCommand: `mr-ducksafe lang ...` sets the language itself, so it must
// not be interrupted by the first-run prompt.
func isLangCommand(args []string) bool {
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		return a == "lang"
	}
	return false
}

func promptLang(in io.Reader, out io.Writer) i18n.Lang {
	fmt.Fprint(out, welcomeText())
	welcomed = true
	r := bufio.NewReader(in)
	for tries := 0; tries < 3; tries++ {
		fmt.Fprint(out, tui.Question("Choose your language / Dil seçin",
			[]string{"English", "Türkçe"}, "(default / varsayılan)"))
		line, err := r.ReadString('\n')
		if strings.TrimSpace(line) == "" {
			return i18n.EN
		}
		if l, ok := i18n.Parse(line); ok {
			return l
		}
		if err != nil {
			break
		}
	}
	return i18n.EN
}

func loadSavedLang() (i18n.Lang, bool) {
	p, err := langFile()
	if err != nil {
		return i18n.EN, false
	}
	b, err := os.ReadFile(p) //nolint:gosec // G304: fixed path under the user's data dir
	if err != nil {
		return i18n.EN, false
	}
	return i18n.Parse(string(b))
}

func saveLang(l i18n.Lang) error {
	p, err := langFile()
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte(string(l)+"\n"), 0o600)
}

// NewLangCmd `mr-ducksafe lang [en|tr]` shows or changes the saved language.
func NewLangCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "lang [en|tr]",
		Short:     i18n.T("lang.short"),
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"en", "tr"},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				fmt.Fprintln(out, i18n.Tf("lang.current", i18n.Current()))
				return nil
			}
			l, ok := i18n.Parse(args[0])
			if !ok {
				return fmt.Errorf(i18n.T("lang.invalid"), args[0])
			}
			if err := saveLang(l); err != nil {
				return err
			}
			i18n.Set(l)
			fmt.Fprintln(out, i18n.T("lang.saved"))
			return nil
		},
	}
}
