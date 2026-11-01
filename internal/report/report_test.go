package report

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/i18n"

	"github.com/fatih/color"
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
		Remediation: map[string]scanengine.Remediation{
			"Firefox": {WingetID: "Mozilla.Firefox", UpdateURL: "https://mozilla.org/firefox/"},
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

func TestWrite_Summary_GroupsAndAdvises(t *testing.T) {
	color.NoColor = true
	var buf bytes.Buffer
	if err := Write(&buf, fixtureResult(), FormatTable, Options{}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	wants := []string{
		"güncellenmesi gereken 1 yazılım",     // gruplu başlık
		"Firefox",                             // uygulama adı
		"1 yüksek",                            // severity sayımı
		"Yakında güncelle",                    // aciliyet (high)
		"Bellek hatası",                       // risk (use-after-free keyword)
		"winget upgrade --id Mozilla.Firefox", // remediation
		"--details",                           // detay ipucu
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("özet çıktısı %q içermeli.\nÇıktı:\n%s", want, out)
		}
	}
}

func TestWrite_Summary_NoFindings(t *testing.T) {
	color.NoColor = true
	var buf bytes.Buffer
	r := fixtureResult()
	r.Findings = nil
	if err := Write(&buf, r, FormatTable, Options{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "açık bulunamadı") {
		t.Errorf("bulgu yokken 'açık bulunamadı' beklenir, alındı: %q", buf.String())
	}
}

func TestWrite_AppFilter_ShowsTable(t *testing.T) {
	color.NoColor = true
	var buf bytes.Buffer
	if err := Write(&buf, fixtureResult(), FormatTable, Options{AppFilter: "firefox"}); err != nil {
		t.Fatal(err)
	}
	// --app verilince tam tabloya gider; CVE id görünmeli.
	if !strings.Contains(buf.String(), "CVE-2022-26485") {
		t.Errorf("app filtreli tablo CVE içermeli, alındı:\n%s", buf.String())
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

func TestWrite_Summary_EN(t *testing.T) {
	i18n.Set(i18n.EN)
	defer i18n.Set(i18n.TR)
	color.NoColor = true
	var buf bytes.Buffer
	if err := Write(&buf, fixtureResult(), FormatTable, Options{}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"1 applications need updating",
		"HIGH",
		"1 high",
		"Urgency:",
		"Update soon",
		"memory error",
		"Fix:",
		"winget upgrade --id Mozilla.Firefox",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("English summary should contain %q.\nOutput:\n%s", want, out)
		}
	}
	for _, notWant := range []string{"Aciliyet", "Çözüm", "yüksek"} {
		if strings.Contains(out, notWant) {
			t.Errorf("English summary must not contain Turkish %q.\nOutput:\n%s", notWant, out)
		}
	}
}
