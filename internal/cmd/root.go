// Package cmd mr-ducksafe CLI komutlarını tanımlar.
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	flagVerbose bool
	flagDBPath  string
)

// NewRootCmd mr-ducksafe root komutunu oluşturur.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "mr-ducksafe",
		Short:         "Yerel kurulu yazılımlar için CVE tarayıcı",
		Long:          "mr-ducksafe makinendeki kurulu sistem yazılımlarını tarayıp OSV.dev üzerinden bilinen güvenlik açıklarını raporlar.",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "ayrıntılı log çıktısı")
	root.PersistentFlags().StringVar(&flagDBPath, "db", "", "SQLite DB yolu (default: ~/.mr-ducksafe/mr-ducksafe.db)")

	root.AddCommand(NewVersionCmd())
	root.AddCommand(NewScanCmd())
	root.AddCommand(NewPruneCmd())
	root.AddCommand(NewHistoryCmd())
	return root
}

// Execute mr-ducksafe CLI'sini çalıştırır.
func Execute() error {
	return NewRootCmd().Execute()
}
