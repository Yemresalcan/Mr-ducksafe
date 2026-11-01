// Package cmd mr-ducksafe CLI komutlarını tanımlar.
package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
)

var (
	flagVerbose bool
	flagDBPath  string
	flagNoColor bool
)

// NewRootCmd mr-ducksafe root komutunu oluşturur.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "mr-ducksafe",
		Short:         i18n.T("root.short"),
		Long:          i18n.T("root.long"),
		SilenceUsage:  true,
		SilenceErrors: false,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if flagNoColor {
				color.NoColor = true
			}
			if flagLang != "" {
				if _, ok := i18n.Parse(flagLang); !ok {
					return fmt.Errorf(i18n.T("lang.invalid"), flagLang)
				}
			}
			return nil
		},
	}
	root.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, i18n.T("flag.verbose"))
	root.PersistentFlags().StringVar(&flagDBPath, "db", "", i18n.T("flag.db"))
	root.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, i18n.T("flag.nocolor"))
	root.PersistentFlags().StringVar(&flagLang, "lang", "", i18n.T("flag.lang"))

	root.AddCommand(NewVersionCmd())
	root.AddCommand(NewScanCmd())
	root.AddCommand(NewPruneCmd())
	root.AddCommand(NewHistoryCmd())
	root.AddCommand(NewLangCmd())
	return root
}

// Execute mr-ducksafe CLI'sini çalıştırır.
func Execute() error {
	i18n.Set(resolveLanguage(os.Args[1:], os.Stdin, os.Stderr))
	root := NewRootCmd()
	done := prepareDoubleClick(root)
	err := root.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("err.prefix"), err)
	}
	done()
	return err
}
