package report

import (
	"fmt"
	"io"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

func writeMarkdown(w io.Writer, r *scanengine.ScanResult, _ Options) error {
	counts := r.CountBySeverity()
	fmt.Fprintf(w, "# mr-ducksafe scan #%d\n\n", r.ScanID)
	fmt.Fprintf(w, "- Tarandı: **%d** kurulu yazılım\n", len(r.Apps))
	fmt.Fprintf(w, "- Eşleştirilemedi: **%d**\n", len(r.Unmatched))
	fmt.Fprintf(w, "- Bulgular: 🔴 %d kritik · 🟠 %d yüksek · 🟡 %d orta · 🔵 %d düşük\n\n",
		counts["critical"], counts["high"], counts["medium"], counts["low"])

	if len(r.Findings) == 0 {
		fmt.Fprintln(w, "✅ Açık güvenlik bulgusu yok.")
		return nil
	}

	fmt.Fprintln(w, "| Severity | Uygulama | Sürüm | CVE | Düzeltme | Özet |")
	fmt.Fprintln(w, "|---|---|---|---|---|---|")
	for _, f := range r.Findings {
		summary := f.Summary
		if len(summary) > 100 {
			summary = summary[:97] + "..."
		}
		fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s |\n",
			f.Severity, f.AppName, f.AppVersion, f.CVEID, f.FixedVersion, summary)
	}
	return nil
}
