package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

type jsonOutput struct {
	SchemaVersion string          `json:"schemaVersion"`
	ScanID        int64           `json:"scanId"`
	StartedAt     time.Time       `json:"startedAt"`
	FinishedAt    time.Time       `json:"finishedAt"`
	AppsTotal     int             `json:"appsTotal"`
	Counts        map[string]int  `json:"countsBySeverity"`
	Findings      []jsonFinding   `json:"findings"`
	Unmatched     []jsonUnmatched `json:"unmatched"`
	Warnings      []string        `json:"warnings,omitempty"`
}

type jsonFinding struct {
	AppName      string   `json:"appName"`
	AppVersion   string   `json:"appVersion"`
	CVEID        string   `json:"cveId"`
	Severity     string   `json:"severity"`
	FixedVersion string   `json:"fixedVersion,omitempty"`
	Summary      string   `json:"summary"`
	References   []string `json:"references,omitempty"`
}

type jsonUnmatched struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Publisher string `json:"publisher"`
}

func writeJSON(w io.Writer, r *scanengine.ScanResult, _ Options) error {
	out := jsonOutput{
		SchemaVersion: "1",
		ScanID:        r.ScanID,
		StartedAt:     r.StartedAt,
		FinishedAt:    r.Finished,
		AppsTotal:     len(r.Apps),
		Counts:        r.CountBySeverity(),
	}
	for _, f := range r.Findings {
		out.Findings = append(out.Findings, jsonFinding{
			AppName: f.AppName, AppVersion: f.AppVersion,
			CVEID: f.CVEID, Severity: f.Severity,
			FixedVersion: f.FixedVersion, Summary: f.Summary,
			References: f.References,
		})
	}
	for _, u := range r.Unmatched {
		out.Unmatched = append(out.Unmatched, jsonUnmatched{
			Name: u.Name, Version: u.Version, Publisher: u.Publisher,
		})
	}
	for _, warn := range r.Warnings {
		out.Warnings = append(out.Warnings, warn.Error())
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
