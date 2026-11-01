package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

var flagPruneOlderThan string

// NewPruneCmd `mr-ducksafe prune` komutu.
func NewPruneCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "prune",
		Short: i18n.T("prune.short"),
		RunE:  runPrune,
	}
	c.Flags().StringVar(&flagPruneOlderThan, "older-than", "30d", i18n.T("flag.olderthan"))
	return c
}

func runPrune(cmd *cobra.Command, _ []string) error {
	d, err := parseHumanDuration(flagPruneOlderThan)
	if err != nil {
		return err
	}

	dbPath := flagDBPath
	if dbPath == "" {
		p, err := config.DBPath()
		if err != nil {
			return err
		}
		dbPath = p
	}
	st, err := store.OpenSQLite("file:" + dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	n, err := st.PruneOlderThan(cmd.Context(), d)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), i18n.Tf("prune.done", n))
	return nil
}

// parseHumanDuration "30d" gibi gün eklerini destekler; gerisini
// Go'nun standart time.ParseDuration'ına bırakır (24h, 90m, vb.).
func parseHumanDuration(s string) (time.Duration, error) {
	if len(s) >= 2 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil {
			return 0, fmt.Errorf(i18n.T("prune.invalid"), s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
