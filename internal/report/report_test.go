package report

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

func fixtureResult() *scanengine.ScanResult {
	t0 := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(30 * time.Second)
	return &scanengine.ScanResult{
		ScanID:    42,
		StartedAt: t0,
		Finished:  t1,
		Apps: []scanner.InstalledApp{
			{Name: "Firefox", Version: "100.0"},
			{Name: "Some Random Tool", Version: "1.0", Publisher: "Nobody"},
		},
		Findings: []store.Finding{
			{
				AppName:      "Firefox",
				AppVersion:   "100.0",
				CVEID:        "CVE-2022-26485",
				Severity:     "high",
				FixedVersion: "97.0.2",
				Summary:      "Use-after-free in XSLT parameter processing",
				References:   []string{"https://example.com/advisory"},
			},
		},
		Unmatched: []scanner.InstalledApp{
			{Name: "Some Random Tool", Version: "1.0", Publisher: "Nobody"},
		},
	}
}

func TestWrite_JSON_Golden(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, fixtureResult(), FormatJSON, Options{}); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/result.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if normalize(buf.String()) != normalize(string(want)) {
		t.Errorf("JSON çıktısı golden ile eşleşmiyor.\nGot:\n%s\nWant:\n%s", buf.String(), string(want))
	}
}

func TestWrite_Markdown_Golden(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, fixtureResult(), FormatMarkdown, Options{}); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/result.golden.md")
	if err != nil {
		t.Fatal(err)
	}
	if normalize(buf.String()) != normalize(string(want)) {
		t.Errorf("Markdown çıktısı golden ile eşleşmiyor.\nGot:\n%s\nWant:\n%s", buf.String(), string(want))
	}
}

// normalize trailing whitespace ve CRLF farklarını tolere eder
func normalize(s string) string {
	out := []byte{}
	for _, b := range []byte(s) {
		if b == '\r' {
			continue
		}
		out = append(out, b)
	}
	for len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return string(out)
}
