package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

var flagHistoryLimit int

// NewHistoryCmd `mr-ducksafe history` komutu.
func NewHistoryCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "history",
		Short: "Geçmiş tarama kayıtlarını listele",
		RunE:  runHistory,
	}
	c.Flags().IntVar(&flagHistoryLimit, "limit", 20, "kaç kayıt gösterilsin")
	return c
}

func runHistory(cmd *cobra.Command, _ []string) error {
	dbPath := flagDBPath
	if dbPath == "" {
		p, err := config.DBPath()
		if err != nil {
			return err
		}
		dbPath = p
	}
	st, err := store.OpenSQLite("file:" + dbPath + "?_journal=WAL&_fk=1")
	if err != nil {
		return fmt.Errorf("DB aç: %w", err)
	}
	defer st.Close()

	ctx := cmd.Context()
	scans, err := st.ListScans(ctx, flagHistoryLimit)
	if err != nil {
		return err
	}
	if len(scans) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Henüz tarama kaydı yok. `mr-ducksafe scan` ile başla.")
		return nil
	}

	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTarih\tDurum\tSüre\tBulgu")
	for _, s := range scans {
		duration := "-"
		if s.FinishedAt != nil {
			duration = s.FinishedAt.Sub(s.StartedAt).Round(1e9).String()
		}
		findings := "-"
		if n, err := st.CountFindings(ctx, s.ID); err == nil {
			findings = fmt.Sprintf("%d", n)
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			s.ID,
			s.StartedAt.Local().Format("2006-01-02 15:04:05"),
			s.Status,
			duration,
			findings,
		)
	}
	return tw.Flush()
}
