package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/nvd"
	"github.com/yemresalcan/mr-ducksafe/internal/report"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
	"github.com/yemresalcan/mr-ducksafe/internal/tui"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

var (
	flagScanFormat        string
	flagScanFailOn        string
	flagScanShowUnmatched bool
	flagScanDetails       bool
	flagScanApp           string
	flagScanOutput        string
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
	c.Flags().BoolVar(&flagScanDetails, "details", false, "özet yerine tüm CVE'leri tablo halinde göster")
	c.Flags().StringVar(&flagScanApp, "app", "", "sadece adı bu metni içeren uygulamayı göster")
	c.Flags().StringVarP(&flagScanOutput, "output", "o", "", "raporu bu dosyaya yaz (varsayılan: stdout)")
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
	stderr := cmd.ErrOrStderr()
	tty := tui.IsTTY(os.Stderr)
	if tty {
		fmt.Fprint(stderr, tui.Banner(Version))
	}

	nvdClient := nvd.NewClient()
	if !nvdClient.HasAPIKey() {
		fmt.Fprintln(stderr,
			"ipucu: NVD_API_KEY ortam değişkeni ayarlanmamış — tarama yavaş olabilir (5 istek/30sn).")
		fmt.Fprintln(stderr,
			"       Ücretsiz key: https://nvd.nist.gov/developers/request-an-api-key")
	}
	vulnClient := vuln.NewCachedClient(nvdClient)

	eng := scanengine.New(sc, m, vulnClient, st)
	result, err := runWithProgress(ctx, eng, stderr, tty)
	if err != nil {
		fmt.Fprintln(stderr, "scan failed:", err)
		exitFunc(2)
	}

	out := cmd.OutOrStdout()
	if flagScanOutput != "" {
		f, err := os.Create(flagScanOutput) //nolint:gosec // G304: the path is chosen by the user via --output
		if err != nil {
			return fmt.Errorf("output dosyası aç: %w", err)
		}
		defer f.Close()
		out = f
	}
	if err := report.Write(out, result, report.Format(flagScanFormat), report.Options{
		ShowUnmatched: flagScanShowUnmatched,
		Details:       flagScanDetails,
		AppFilter:     flagScanApp,
	}); err != nil {
		return err
	}
	if flagScanOutput != "" {
		fmt.Fprintf(stderr, "Rapor yazıldı: %s\n", flagScanOutput)
	}

	if shouldFail(result, flagScanFailOn) {
		exitFunc(1)
	}
	return nil
}

// runWithProgress engine.Run'ı bir progress channel ile çalıştırır ve
// aşama olaylarını canlı spinner'a çevirir.
func runWithProgress(ctx context.Context, eng *scanengine.Engine, w io.Writer, tty bool) (*scanengine.ScanResult, error) {
	prog := tui.NewProgress(w, tty)
	prog.Update("Kurulu yazılımlar taranıyor...")
	prog.Start()

	ch := make(chan scanengine.Progress, 16)
	var (
		result *scanengine.ScanResult
		runErr error
		wg     sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, runErr = eng.Run(ctx, ch)
		close(ch)
	}()

	for p := range ch {
		switch p.Stage {
		case scanengine.StageScan:
			prog.Complete(fmt.Sprintf("Tarama tamamlandı — %d program", p.Total))
			prog.Update("Eşleştiriliyor...")
		case scanengine.StageMatch:
			prog.Complete(fmt.Sprintf("Eşleştirme tamamlandı — %d eşleşme", p.Done))
			prog.Update("NVD sorgusu başlıyor...")
		case scanengine.StageQuery:
			prog.Update(fmt.Sprintf("NVD sorgusu... %d/%d", p.Done, p.Total))
		}
	}
	wg.Wait()

	if runErr == nil && result != nil {
		prog.Complete(fmt.Sprintf("Tamamlandı — %d bulgu", len(result.Findings)))
	}
	prog.Stop()
	return result, runErr
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
