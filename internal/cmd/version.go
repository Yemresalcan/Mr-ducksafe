package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
)

// Version build sırasında ldflags ile set edilir; yoksa "dev".
var Version = "dev"

// NewVersionCmd `mr-ducksafe version` komutu.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: i18n.T("version.short"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "mr-ducksafe %s\n", Version)
			return nil
		},
	}
}
