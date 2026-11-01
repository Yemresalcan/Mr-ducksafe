package report

import (
	"fmt"
	"io"
	"sort"

	"github.com/fatih/color"
	"github.com/yemresalcan/mr-ducksafe/internal/advice"
	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

// appGroup bir uygulamanın tüm bulgularının özeti.
type appGroup struct {
	name      string
	version   string
	counts    map[string]int
	worst     string
	summaries []string
}

var severityRank = map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1, "unknown": 0}

func writeSummary(w io.Writer, r *scanengine.ScanResult, opts Options) error {
	if opts.NoColor {
		color.NoColor = true
	}

	groups := groupByApp(r.Findings)

	if len(groups) == 0 {
		fmt.Fprintln(w, i18n.Tf("sum.clean", color.GreenString("✓"), len(r.Apps)))
		return nil
	}

	fmt.Fprintf(w, "%s\n\n", i18n.Tf("sum.title",
		color.New(color.Bold).Sprintf("%d", len(groups)), len(r.Findings)))

	for _, g := range groups {
		fmt.Fprintf(w, "  %s  %s  %s\n",
			colorSeverity(g.worst),
			color.New(color.Bold).Sprint(g.name),
			color.New(color.Faint).Sprint(g.version),
		)
		fmt.Fprintf(w, "          %s\n", countsLine(g.counts))
		fmt.Fprintf(w, "          %s %s\n", color.New(color.Faint).Sprint(i18n.T("sum.urgency")), advice.Urgency(g.worst))
		fmt.Fprintf(w, "          %s %s\n", color.New(color.Faint).Sprint(i18n.T("sum.risk")), advice.RiskLine(g.worst, g.summaries))
		fmt.Fprintf(w, "          %s %s\n\n", color.New(color.Faint).Sprint(i18n.T("sum.fix")), remediationLine(r, g.name))
	}

	fmt.Fprintln(w, color.New(color.Faint).Sprint(i18n.T("sum.details")))

	if opts.ShowUnmatched && len(r.Unmatched) > 0 {
		fmt.Fprintln(w, i18n.Tf("sum.unmatched", len(r.Unmatched)))
		for _, u := range r.Unmatched {
			fmt.Fprintf(w, "  - %s %s\n", u.Name, u.Version)
		}
	}
	return nil
}

func groupByApp(findings []store.Finding) []appGroup {
	byName := map[string]*appGroup{}
	for _, f := range findings {
		g, ok := byName[f.AppName]
		if !ok {
			g = &appGroup{name: f.AppName, version: f.AppVersion, counts: map[string]int{}}
			byName[f.AppName] = g
		}
		g.counts[f.Severity]++
		if f.Summary != "" {
			g.summaries = append(g.summaries, f.Summary)
		}
		if severityRank[f.Severity] > severityRank[g.worst] {
			g.worst = f.Severity
		}
	}

	out := make([]appGroup, 0, len(byName))
	for _, g := range byName {
		out = append(out, *g)
	}
	// En tehlikeli üstte; eşitlikte bulgu sayısı çok olan üstte.
	sort.Slice(out, func(i, j int) bool {
		if severityRank[out[i].worst] != severityRank[out[j].worst] {
			return severityRank[out[i].worst] > severityRank[out[j].worst]
		}
		return total(out[i].counts) > total(out[j].counts)
	})
	return out
}

func total(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}

func countsLine(c map[string]int) string {
	parts := []string{}
	if c["critical"] > 0 {
		parts = append(parts, color.RedString("%s", i18n.Tf("count.critical", c["critical"])))
	}
	if c["high"] > 0 {
		parts = append(parts, color.HiRedString("%s", i18n.Tf("count.high", c["high"])))
	}
	if c["medium"] > 0 {
		parts = append(parts, color.YellowString("%s", i18n.Tf("count.medium", c["medium"])))
	}
	if c["low"] > 0 {
		parts = append(parts, color.BlueString("%s", i18n.Tf("count.low", c["low"])))
	}
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " · "
		}
		out += p
	}
	return out
}

func remediationLine(r *scanengine.ScanResult, appName string) string {
	rem, ok := r.Remediation[appName]
	if ok && rem.WingetID != "" {
		return color.CyanString("winget upgrade --id %s", rem.WingetID)
	}
	if ok && rem.UpdateURL != "" {
		return i18n.T("sum.update") + color.CyanString(rem.UpdateURL)
	}
	return i18n.T("sum.updateapp")
}
