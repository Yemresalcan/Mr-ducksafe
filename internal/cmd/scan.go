package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/nvd"
	"github.com/yemresalcan/mr-ducksafe/internal/report"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

var (
	flagScanFormat        string
	flagScanFailOn        string
	flagScanShowUnmatched bool
)

// NewScanCmd `mr-ducksafe scan` komutu.
func NewScanCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan",
		Short: "Kurulu yazılımları tarayıp CVE raporu üret",
		RunE:  runScan,
	}
	c.Flags().StringVar(&flagScanFormat, "format", "table", "çıktı formatı: table | json | markdown")
	c.Flags().StringVar(&flagScanFailOn, "fail-on", "", "eşik aşıldığında exit 1: critical | high | medium | low")
	c.Flags().BoolVar(&flagScanShowUnmatched, "show-unmatched", false, "eşleştirilemeyen yazılımları da göster")
	return c
}

func runScan(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

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

	sc, err := scanner.New()
	if err != nil {
		return err
	}
	m, err := matcher.NewDefaultMatcher()
	if err != nil {
		return err
	}
	nvdClient := nvd.NewClient()
	if !nvdClient.HasAPIKey() {
		fmt.Fprintln(cmd.ErrOrStderr(),
			"ipucu: NVD_API_KEY ortam değişkeni ayarlanmamış — tarama yavaş olabilir (5 istek/30sn).")
		fmt.Fprintln(cmd.ErrOrStderr(),
			"       Ücretsiz key: https://nvd.nist.gov/developers/request-an-api-key")
	}
	vulnClient := vuln.NewCachedClient(nvdClient)

	eng := scanengine.New(sc, m, vulnClient, st)
	result, err := eng.Run(ctx, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan failed:", err)
		os.Exit(2)
	}

	if err := report.Write(cmd.OutOrStdout(), result, report.Format(flagScanFormat), report.Options{
		ShowUnmatched: flagScanShowUnmatched,
	}); err != nil {
		return err
	}

	if shouldFail(result, flagScanFailOn) {
		os.Exit(1)
	}
	return nil
}

func shouldFail(r *scanengine.ScanResult, threshold string) bool {
	if threshold == "" {
		return false
	}
	rank := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
	thr, ok := rank[threshold]
	if !ok {
		return false
	}
	for _, f := range r.Findings {
		if rank[f.Severity] >= thr {
			return true
		}
	}
	return false
}
