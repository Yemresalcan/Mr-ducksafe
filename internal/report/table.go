package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

// filterByApp sonucu sadece adı filtreyi (case-insensitive) içeren
// uygulamanın bulgularına daraltır.
func filterByApp(r *scanengine.ScanResult, filter string) *scanengine.ScanResult {
	f := strings.ToLower(filter)
	out := *r // sığ kopya; slice'ları aşağıda değiştiriyoruz
	var findings []store.Finding
	for _, fn := range r.Findings {
		if strings.Contains(strings.ToLower(fn.AppName), f) {
			findings = append(findings, fn)
		}
	}
	out.Findings = findings
	return &out
}

func writeTable(w io.Writer, r *scanengine.ScanResult, opts Options) error {
	if opts.NoColor {
		color.NoColor = true
	}

	if opts.AppFilter != "" {
		r = filterByApp(r, opts.AppFilter)
	}

	counts := r.CountBySeverity()
	fmt.Fprintf(w, "mr-ducksafe scan #%d — %d kurulu yazılım tarandı (%d eşleştirilemedi)\n",
		r.ScanID, len(r.Apps), len(r.Unmatched))
	fmt.Fprintf(w, "Bulgular: %s %d  %s %d  %s %d  %s %d\n\n",
		color.RedString("KRITIK"), counts["critical"],
		color.HiRedString("YÜKSEK"), counts["high"],
		color.YellowString("ORTA"), counts["medium"],
		color.BlueString("DÜŞÜK"), counts["low"],
	)

	if len(r.Findings) == 0 {
		fmt.Fprintln(w, color.GreenString("✓ Açık güvenlik bulgusu yok."))
	} else {
		t := tablewriter.NewWriter(w)
		t.Header([]string{"Sev", "Uygulama", "Sürüm", "CVE", "Düzeltme", "Özet"})

		for _, f := range r.Findings {
			summary := f.Summary
			if len(summary) > 70 {
				summary = summary[:67] + "..."
			}
			_ = t.Append([]string{
				colorSeverity(f.Severity),
				f.AppName, f.AppVersion, f.CVEID, f.FixedVersion, summary,
			})
		}
		_ = t.Render()
	}

	if opts.ShowUnmatched && len(r.Unmatched) > 0 {
		fmt.Fprintf(w, "\nEşleştirilemeyen yazılımlar (%d):\n", len(r.Unmatched))
		for _, u := range r.Unmatched {
			fmt.Fprintf(w, "  - %s %s (%s)\n", u.Name, u.Version, u.Publisher)
		}
	}

	if len(r.Warnings) > 0 {
		fmt.Fprintf(w, "\n%d uyarı (use --verbose for detail)\n", len(r.Warnings))
	}
	return nil
}

func colorSeverity(s string) string {
	switch s {
	case "critical":
		return color.RedString("KRIT")
	case "high":
		return color.HiRedString("YÜKS")
	case "medium":
		return color.YellowString("ORTA")
	case "low":
		return color.BlueString("DÜŞK")
	default:
		return color.WhiteString("UNK ")
	}
}
