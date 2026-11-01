package report

import (
	"fmt"
	"io"

	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

func writeMarkdown(w io.Writer, r *scanengine.ScanResult, _ Options) error {
	counts := r.CountBySeverity()
	fmt.Fprintf(w, "# mr-ducksafe scan #%d\n\n", r.ScanID)
	fmt.Fprintln(w, i18n.Tf("md.scanned", len(r.Apps)))
	fmt.Fprintln(w, i18n.Tf("md.unmatched", len(r.Unmatched)))
	fmt.Fprintf(w, "%s\n\n", i18n.Tf("md.findings",
		counts["critical"], counts["high"], counts["medium"], counts["low"]))

	if len(r.Findings) == 0 {
		fmt.Fprintln(w, i18n.T("md.none"))
		return nil
	}

	fmt.Fprintln(w, i18n.T("md.header"))
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
