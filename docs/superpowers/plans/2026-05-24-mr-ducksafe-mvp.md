# mr-ducksafe MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Tek dosya Go binary olarak çalışan, Windows'ta kurulu sistem yazılımlarını OSV.dev üzerinden CVE'lere eşleştirip CLI ve lokal web UI ile raporlayan `mr-ducksafe` aracının MVP'sini inşa et.

**Architecture:** `cmd/mr-ducksafe` katmanı sadece CLI parsing yapar; tüm iş mantığı `internal/` altında, her paket interface arkasında tek sorumluluğa sahip. `scanengine` orkestrasyonu yapar ve hem CLI hem HTTP server tarafından çağrılır.

**Tech Stack:** Go 1.22+, `spf13/cobra`, `modernc.org/sqlite` (pure-Go), `golang.org/x/sys/windows/registry`, stdlib `net/http` + `html/template` + `embed`, HTMX (gömülü), `olekukonko/tablewriter`, `log/slog`.

**Spec:** [2026-05-23-mr-ducksafe-design.md](../specs/2026-05-23-mr-ducksafe-design.md)

**Test felsefesi:** Spec bölüm 9 — **test-after pragmatik**. Önce implementasyon, sonra anlamlı davranışları kapsayan testler. TDD red-green-refactor disiplini uygulanmayacak.

---

## Dosya Yapısı

Plan boyunca aşağıdaki yapı oluşturulacak. Her task hangi dosyalara dokunduğunu açıkça belirtir.

```
mr-ducksafe/
├── go.mod
├── go.sum
├── README.md
├── .gitignore
├── .golangci.yml
├── .github/workflows/ci.yml
│
├── cmd/mr-ducksafe/
│   └── main.go                     # entrypoint
│
├── internal/
│   ├── cmd/
│   │   ├── root.go                 # cobra root
│   │   ├── scan.go                 # scan subcommand
│   │   ├── serve.go                # serve subcommand
│   │   ├── version.go              # version subcommand
│   │   └── prune.go                # prune subcommand
│   │
│   ├── scanner/
│   │   ├── scanner.go              # Scanner interface + InstalledApp
│   │   ├── scanner_windows.go      # //go:build windows
│   │   └── scanner_other.go        # //go:build !windows
│   │
│   ├── matcher/
│   │   ├── matcher.go              # interface + types
│   │   ├── mapping.go              # YAML loader
│   │   ├── version.go              # version normalization
│   │   ├── embedded.go             # //go:embed data/mappings.yaml
│   │   └── data/mappings.yaml      # curated 50 mappings
│   │
│   ├── osv/
│   │   ├── client.go               # Client interface + types
│   │   ├── http.go                 # HTTP impl
│   │   └── cache.go                # in-memory TTL cache
│   │
│   ├── store/
│   │   ├── store.go                # Store interface + types
│   │   ├── sqlite.go               # SQLite impl
│   │   └── schema.sql              # embedded migrations
│   │
│   ├── scanengine/
│   │   ├── engine.go               # orchestration
│   │   ├── result.go               # ScanResult + ScanError types
│   │   └── progress.go             # Progress channel types
│   │
│   ├── report/
│   │   ├── report.go               # Format function + types
│   │   ├── table.go                # table formatter
│   │   ├── json.go                 # JSON formatter
│   │   └── markdown.go             # markdown formatter
│   │
│   ├── server/
│   │   ├── server.go               # HTTP server setup
│   │   ├── handlers.go             # endpoint handlers
│   │   ├── sse.go                  # SSE handler
│   │   ├── embed.go                # //go:embed templates + static
│   │   ├── templates/              # *.html
│   │   └── static/                 # *.css, htmx.min.js
│   │
│   └── config/
│       └── config.go               # ~/.mr-ducksafe/config.yaml reader
│
├── testdata/
│   ├── registry_dump.json
│   ├── osv_responses/
│   │   └── firefox_100.json
│   └── mappings_test.yaml
│
└── docs/
    └── superpowers/
        ├── specs/2026-05-23-mr-ducksafe-design.md
        └── plans/2026-05-24-mr-ducksafe-mvp.md
```

---

## Phase 1 — Foundation

### Task 1: Proje bootstrap

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `README.md`
- Create: `cmd/mr-ducksafe/main.go`

- [ ] **Step 1: Git repo başlat ve module oluştur**

```bash
git init
go mod init github.com/yemresalcan/mr-ducksafe
```

- [ ] **Step 2: `.gitignore` yaz**

```gitignore
# Build artifacts
/mr-ducksafe
/mr-ducksafe.exe
/dist/
/bin/

# Test artifacts
*.test
*.out
coverage.out
coverage.html

# Local data
*.db
*.db-journal
.mr-ducksafe/

# IDE
.vscode/
.idea/
*.swp

# OS
.DS_Store
Thumbs.db
```

- [ ] **Step 3: Minimal `cmd/mr-ducksafe/main.go` yaz**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("mr-ducksafe — MVP boot")
	os.Exit(0)
}
```

- [ ] **Step 4: Build ve çalıştırarak doğrula**

```bash
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
./mr-ducksafe.exe
```

Expected output: `mr-ducksafe — MVP boot`

- [ ] **Step 5: `README.md` taslağı yaz**

```markdown
# mr-ducksafe

Yerel makinede kurulu sistem yazılımları için bilinen CVE'leri (OSV.dev üzerinden) raporlayan tek binary CLI ve lokal web aracı.

**Durum:** MVP geliştirme aşamasında.

## Kullanım (planlanan)

```sh
mr-ducksafe scan                  # CLI tarama, terminal tablosu
mr-ducksafe scan --json           # JSON çıktı
mr-ducksafe serve --port 8080     # Lokal web UI
mr-ducksafe version
```

## Geliştirme

Detaylı tasarım: [docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md](docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md)
Implementasyon planı: [docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md](docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md)
```

- [ ] **Step 6: Commit**

```bash
git add .gitignore go.mod README.md cmd/mr-ducksafe/main.go docs/
git commit -m "chore: bootstrap project with module, gitignore, and skeleton main"
```

---

### Task 2: CI ve linter konfigürasyonu

**Files:**
- Create: `.golangci.yml`
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: `.golangci.yml` yaz**

```yaml
run:
  timeout: 5m

linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - ineffassign
    - gosimple
    - gofmt
    - goimports
    - misspell
    - unconvert
    - bodyclose
    - gosec

linters-settings:
  gosec:
    excludes:
      - G104  # audited error returns

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
```

- [ ] **Step 2: `.github/workflows/ci.yml` yaz**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build
        run: go build ./...
      - name: Test
        run: go test ./... -race -coverprofile=coverage.out
      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage-${{ matrix.os }}
          path: coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

- [ ] **Step 3: Lokalde lint çalıştır (golangci-lint kuruluysa)**

```bash
golangci-lint run ./...
```

Expected: No errors (henüz iş kodu yok).

- [ ] **Step 4: Commit**

```bash
git add .golangci.yml .github/
git commit -m "chore: add golangci-lint config and GitHub Actions CI matrix"
```

---

### Task 3: Scanner interface ve domain tipi

**Files:**
- Create: `internal/scanner/scanner.go`
- Create: `internal/scanner/scanner_other.go`
- Create: `internal/scanner/scanner_test.go`

- [ ] **Step 1: `internal/scanner/scanner.go` yaz**

```go
// Package scanner kurulu sistem yazılımlarının listesini çıkaran
// OS-spesifik scanner'ları tanımlar.
package scanner

import (
	"context"
	"errors"
)

// InstalledApp bir kurulu yazılımın tespit edilen meta verilerini taşır.
type InstalledApp struct {
	Name      string
	Version   string
	Publisher string
	Source    string
}

// Scanner OS-spesifik kurulu yazılım listeleme arayüzü.
type Scanner interface {
	Scan(ctx context.Context) ([]InstalledApp, error)
}

// ErrUnsupportedOS runtime OS desteklenmiyorsa döner.
var ErrUnsupportedOS = errors.New("scanner: bu işletim sistemi henüz desteklenmiyor")

// New runtime OS'una göre uygun Scanner implementasyonunu döner.
func New() (Scanner, error) {
	return newPlatformScanner()
}
```

- [ ] **Step 2: `internal/scanner/scanner_other.go` (non-Windows stub) yaz**

```go
//go:build !windows

package scanner

func newPlatformScanner() (Scanner, error) {
	return nil, ErrUnsupportedOS
}
```

- [ ] **Step 3: `internal/scanner/scanner_test.go` yaz**

```go
package scanner

import "testing"

func TestInstalledApp_ZeroValue(t *testing.T) {
	var app InstalledApp
	if app.Name != "" || app.Version != "" {
		t.Errorf("zero value bekleniyor, alındı: %+v", app)
	}
}

func TestErrUnsupportedOS(t *testing.T) {
	if ErrUnsupportedOS == nil {
		t.Fatal("ErrUnsupportedOS nil olmamalı")
	}
	if ErrUnsupportedOS.Error() == "" {
		t.Error("error mesajı boş olmamalı")
	}
}
```

- [ ] **Step 4: Test çalıştır**

```bash
go test ./internal/scanner/...
```

Expected: PASS (Windows üzerinde Windows impl henüz yok ama compile etmesi gerekir).

Eğer Windows'ta build fail ediyorsa, geçici olarak `internal/scanner/scanner_windows_stub.go` ekleyin:

```go
//go:build windows

package scanner

func newPlatformScanner() (Scanner, error) {
	return nil, ErrUnsupportedOS
}
```

Bu Task 4'te gerçek implementasyonla değiştirilecek.

- [ ] **Step 5: Commit**

```bash
git add internal/scanner/
git commit -m "feat(scanner): define Scanner interface and InstalledApp type"
```

---

### Task 4: Windows Registry scanner implementasyonu

**Files:**
- Create: `internal/scanner/scanner_windows.go` (Task 3'teki stub'ı değiştirir)
- Modify: `go.mod` (golang.org/x/sys dependency eklenir)
- Create: `internal/scanner/scanner_windows_test.go`

- [ ] **Step 1: Dependency ekle**

```bash
go get golang.org/x/sys/windows/registry
```

- [ ] **Step 2: `internal/scanner/scanner_windows.go` yaz**

(Eğer Task 3 sonunda stub eklediysen önce sil: `rm internal/scanner/scanner_windows_stub.go`)

```go
//go:build windows

package scanner

import (
	"context"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const sourceWindowsRegistry = "windows-registry"

var uninstallKeys = []struct {
	root registry.Key
	path string
}{
	{registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Uninstall`},
	{registry.LOCAL_MACHINE, `Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
	{registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall`},
}

type windowsScanner struct{}

func newPlatformScanner() (Scanner, error) {
	return &windowsScanner{}, nil
}

func (s *windowsScanner) Scan(ctx context.Context) ([]InstalledApp, error) {
	var apps []InstalledApp
	for _, k := range uninstallKeys {
		if err := ctx.Err(); err != nil {
			return apps, err
		}
		batch, err := readUninstallKey(k.root, k.path)
		if err != nil {
			// kısmi başarısızlığı tolere et — bir hive'ı okuyamamak fatal değil
			continue
		}
		apps = append(apps, batch...)
	}
	return apps, nil
}

func readUninstallKey(root registry.Key, path string) ([]InstalledApp, error) {
	parent, err := registry.OpenKey(root, path, registry.READ)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer parent.Close()

	subKeys, err := parent.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("list subkeys %s: %w", path, err)
	}

	apps := make([]InstalledApp, 0, len(subKeys))
	for _, name := range subKeys {
		app, ok := readSubKey(parent, name)
		if ok {
			apps = append(apps, app)
		}
	}
	return apps, nil
}

func readSubKey(parent registry.Key, name string) (InstalledApp, bool) {
	sub, err := registry.OpenKey(parent, name, registry.READ)
	if err != nil {
		return InstalledApp{}, false
	}
	defer sub.Close()

	display, _, err := sub.GetStringValue("DisplayName")
	if err != nil || display == "" {
		return InstalledApp{}, false
	}
	version, _, _ := sub.GetStringValue("DisplayVersion")
	publisher, _, _ := sub.GetStringValue("Publisher")

	return InstalledApp{
		Name:      display,
		Version:   version,
		Publisher: publisher,
		Source:    sourceWindowsRegistry,
	}, true
}
```

- [ ] **Step 3: `internal/scanner/scanner_windows_test.go` yaz (sadece Windows'ta çalışır)**

```go
//go:build windows

package scanner

import (
	"context"
	"testing"
	"time"
)

func TestWindowsScanner_Smoke(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apps, err := s.Scan(ctx)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(apps) == 0 {
		t.Skip("Bu makinede Registry'de hiç entry bulunamadı, smoke test atlanıyor")
	}

	for _, a := range apps {
		if a.Name == "" {
			t.Errorf("Name boş bir app döndü: %+v", a)
		}
		if a.Source != sourceWindowsRegistry {
			t.Errorf("beklenmedik Source: %s", a.Source)
		}
	}
	t.Logf("Registry'de %d kurulu yazılım bulundu", len(apps))
}
```

- [ ] **Step 4: Windows üzerinde test çalıştır**

```bash
go test ./internal/scanner/... -v
```

Expected: Smoke test PASS, log'da gerçek kurulu yazılım sayısı.

- [ ] **Step 5: Commit**

```bash
git add internal/scanner/ go.mod go.sum
git commit -m "feat(scanner): implement Windows Registry scanner reading Uninstall keys"
```

---

## Phase 2 — Matching

### Task 5: Matcher interface ve domain tipleri

**Files:**
- Create: `internal/matcher/matcher.go`
- Create: `internal/matcher/matcher_test.go`

- [ ] **Step 1: `internal/matcher/matcher.go` yaz**

```go
// Package matcher InstalledApp'i sorgulanabilir CVE identifier'ına
// (CPE veya purl) eşleştirir.
package matcher

import (
	"context"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
)

// IdentifierType bir Identifier'ın hangi şemada olduğunu belirtir.
type IdentifierType string

const (
	IdentifierTypePURL IdentifierType = "purl"
	IdentifierTypeCPE  IdentifierType = "cpe"
)

// Identifier OSV.dev'e sorgulanabilir bir tekil yazılım kimliği.
type Identifier struct {
	Type      IdentifierType
	Value     string
	Ecosystem string // OSV ecosystem hint, boş olabilir
}

// MatchConfidence eşleştirme güven seviyesi.
type MatchConfidence int

const (
	ConfidenceNone MatchConfidence = iota
	ConfidenceLow
	ConfidenceMedium
	ConfidenceHigh
)

func (c MatchConfidence) String() string {
	switch c {
	case ConfidenceHigh:
		return "high"
	case ConfidenceMedium:
		return "medium"
	case ConfidenceLow:
		return "low"
	default:
		return "none"
	}
}

// Matcher InstalledApp -> Identifier eşleştirme arayüzü.
type Matcher interface {
	Match(ctx context.Context, app scanner.InstalledApp) (Identifier, MatchConfidence, error)
}
```

- [ ] **Step 2: `internal/matcher/matcher_test.go` yaz**

```go
package matcher

import "testing"

func TestMatchConfidence_String(t *testing.T) {
	cases := []struct {
		c    MatchConfidence
		want string
	}{
		{ConfidenceHigh, "high"},
		{ConfidenceMedium, "medium"},
		{ConfidenceLow, "low"},
		{ConfidenceNone, "none"},
	}
	for _, tc := range cases {
		if got := tc.c.String(); got != tc.want {
			t.Errorf("MatchConfidence(%d).String() = %q, want %q", tc.c, got, tc.want)
		}
	}
}
```

- [ ] **Step 3: Test çalıştır**

```bash
go test ./internal/matcher/...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/matcher/
git commit -m "feat(matcher): define Matcher interface, Identifier and Confidence types"
```

---

### Task 6: Version normalization

**Files:**
- Create: `internal/matcher/version.go`
- Create: `internal/matcher/version_test.go`

- [ ] **Step 1: `internal/matcher/version.go` yaz**

```go
package matcher

import (
	"regexp"
	"strings"
)

// VersionScheme bir version string'inin format şemasını belirtir.
type VersionScheme string

const (
	SchemeSemver  VersionScheme = "semver"
	SchemeFour    VersionScheme = "fourpart"
	SchemeCalver  VersionScheme = "calver"
	SchemeUnknown VersionScheme = "unknown"
)

var (
	semverRe   = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)([-+].+)?$`)
	fourPartRe = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+)$`)
	calverRe   = regexp.MustCompile(`^(20\d{2})\.(\d{1,2})\.(\d{1,2})$`)
)

// NormalizeVersion bir version string'ini OSV'ye yollamaya uygun
// normalize edilmiş haliyle ve tespit edilen şemayla döner.
func NormalizeVersion(raw string) (string, VersionScheme) {
	v := strings.TrimSpace(raw)
	v = strings.TrimPrefix(v, "v")

	if m := semverRe.FindStringSubmatch(v); m != nil {
		return m[1] + "." + m[2] + "." + m[3], SchemeSemver
	}
	if m := fourPartRe.FindStringSubmatch(v); m != nil {
		// Dört parça (örn: 120.0.6099.130) — ilk üç parçayı semver olarak kullan
		return m[1] + "." + m[2] + "." + m[3], SchemeFour
	}
	if calverRe.MatchString(v) {
		return v, SchemeCalver
	}
	return v, SchemeUnknown
}
```

- [ ] **Step 2: `internal/matcher/version_test.go` yaz**

```go
package matcher

import "testing"

func TestNormalizeVersion(t *testing.T) {
	cases := []struct {
		raw        string
		wantValue  string
		wantScheme VersionScheme
	}{
		{"1.2.3", "1.2.3", SchemeSemver},
		{"v1.2.3", "1.2.3", SchemeSemver},
		{"1.2.3-beta", "1.2.3", SchemeSemver},
		{"120.0.6099.130", "120.0.6099", SchemeFour},
		{"2024.01.5", "2024.01.5", SchemeCalver},
		{"21H2", "21H2", SchemeUnknown},
		{"  1.0.0  ", "1.0.0", SchemeSemver},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			gotValue, gotScheme := NormalizeVersion(tc.raw)
			if gotValue != tc.wantValue {
				t.Errorf("value = %q, want %q", gotValue, tc.wantValue)
			}
			if gotScheme != tc.wantScheme {
				t.Errorf("scheme = %q, want %q", gotScheme, tc.wantScheme)
			}
		})
	}
}
```

- [ ] **Step 3: Test çalıştır**

```bash
go test ./internal/matcher/... -v
```

Expected: 7 case PASS

- [ ] **Step 4: Commit**

```bash
git add internal/matcher/version.go internal/matcher/version_test.go
git commit -m "feat(matcher): version normalization for semver, fourpart, calver schemes"
```

---

### Task 7: Mapping loader ve embedded YAML

**Files:**
- Create: `internal/matcher/data/mappings.yaml`
- Create: `internal/matcher/embedded.go`
- Create: `internal/matcher/mapping.go`
- Create: `internal/matcher/mapping_test.go`
- Modify: `go.mod` (gopkg.in/yaml.v3 dependency)

- [ ] **Step 1: Dependency ekle**

```bash
go get gopkg.in/yaml.v3
```

- [ ] **Step 2: `internal/matcher/data/mappings.yaml` ilk 10 mapping ile yaz (Task 8'de 40 daha eklenecek)**

```yaml
# mr-ducksafe curated mapping — Windows kurulu yazılım adından CVE identifier'ına.
# Format: bkz. internal/matcher/mapping.go
mappings:
  - registry_name: "Google Chrome"
    publisher: "Google LLC"
    identifier:
      type: cpe
      value: "cpe:2.3:a:google:chrome"
    ecosystem_hint: "Chrome"
    version_normalize: "fourpart"

  - registry_name_pattern: "^Mozilla Firefox.*"
    publisher: "Mozilla"
    identifier:
      type: cpe
      value: "cpe:2.3:a:mozilla:firefox"
    version_normalize: "semver"

  - registry_name: "Microsoft Edge"
    publisher: "Microsoft Corporation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:microsoft:edge_chromium"
    version_normalize: "fourpart"

  - registry_name: "VLC media player"
    publisher: "VideoLAN"
    identifier:
      type: cpe
      value: "cpe:2.3:a:videolan:vlc_media_player"
    version_normalize: "semver"

  - registry_name_pattern: "^7-Zip.*"
    publisher: "Igor Pavlov"
    identifier:
      type: cpe
      value: "cpe:2.3:a:7-zip:7-zip"
    version_normalize: "semver"

  - registry_name_pattern: "^Notepad\\+\\+.*"
    publisher: "Notepad++ Team"
    identifier:
      type: cpe
      value: "cpe:2.3:a:notepad-plus-plus:notepad-plus-plus"
    version_normalize: "semver"

  - registry_name_pattern: "^Adobe Acrobat Reader.*"
    publisher: "Adobe"
    identifier:
      type: cpe
      value: "cpe:2.3:a:adobe:acrobat_reader_dc"
    version_normalize: "semver"

  - registry_name_pattern: "^Python 3\\..*"
    publisher: "Python Software Foundation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:python:python"
    version_normalize: "semver"

  - registry_name_pattern: "^Node\\.js.*"
    publisher: "Node.js Foundation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:nodejs:node.js"
    version_normalize: "semver"

  - registry_name: "Git"
    publisher: "The Git Development Community"
    identifier:
      type: cpe
      value: "cpe:2.3:a:git-scm:git"
    version_normalize: "semver"
```

- [ ] **Step 3: `internal/matcher/embedded.go` yaz**

```go
package matcher

import _ "embed"

//go:embed data/mappings.yaml
var embeddedMappings []byte

// EmbeddedMappings paketle gömülmüş varsayılan mapping verisini döner.
func EmbeddedMappings() []byte {
	return embeddedMappings
}
```

- [ ] **Step 4: `internal/matcher/mapping.go` yaz**

```go
package matcher

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"gopkg.in/yaml.v3"
)

// mappingEntry mappings.yaml içindeki tek bir kayıt.
type mappingEntry struct {
	RegistryName        string `yaml:"registry_name"`
	RegistryNamePattern string `yaml:"registry_name_pattern"`
	Publisher           string `yaml:"publisher"`
	Identifier          struct {
		Type  string `yaml:"type"`
		Value string `yaml:"value"`
	} `yaml:"identifier"`
	EcosystemHint    string `yaml:"ecosystem_hint"`
	VersionNormalize string `yaml:"version_normalize"`

	pattern *regexp.Regexp // pattern compiled cache
}

type mappingFile struct {
	Mappings []mappingEntry `yaml:"mappings"`
}

// MappingMatcher YAML mapping tabanlı Matcher implementasyonu.
type MappingMatcher struct {
	entries []mappingEntry
}

// NewMappingMatcher verilen YAML byte slice'ından MappingMatcher oluşturur.
func NewMappingMatcher(yamlData []byte) (*MappingMatcher, error) {
	var f mappingFile
	if err := yaml.Unmarshal(yamlData, &f); err != nil {
		return nil, fmt.Errorf("mappings YAML parse: %w", err)
	}

	for i := range f.Mappings {
		if pat := f.Mappings[i].RegistryNamePattern; pat != "" {
			re, err := regexp.Compile(pat)
			if err != nil {
				return nil, fmt.Errorf("mapping pattern %q derlenemedi: %w", pat, err)
			}
			f.Mappings[i].pattern = re
		}
	}
	return &MappingMatcher{entries: f.Mappings}, nil
}

// NewDefaultMatcher gömülü mapping verisiyle Matcher döner.
func NewDefaultMatcher() (*MappingMatcher, error) {
	return NewMappingMatcher(EmbeddedMappings())
}

// Match scanner.Matcher arayüzünü uygular.
func (m *MappingMatcher) Match(_ context.Context, app scanner.InstalledApp) (Identifier, MatchConfidence, error) {
	for _, e := range m.entries {
		if e.matches(app) {
			return Identifier{
				Type:      IdentifierType(e.Identifier.Type),
				Value:     e.Identifier.Value,
				Ecosystem: e.EcosystemHint,
			}, ConfidenceHigh, nil
		}
	}
	return Identifier{}, ConfidenceNone, nil
}

func (e mappingEntry) matches(app scanner.InstalledApp) bool {
	switch {
	case e.RegistryName != "":
		if !strings.EqualFold(e.RegistryName, app.Name) {
			return false
		}
	case e.pattern != nil:
		if !e.pattern.MatchString(app.Name) {
			return false
		}
	default:
		return false
	}
	if e.Publisher != "" && !strings.Contains(strings.ToLower(app.Publisher), strings.ToLower(e.Publisher)) {
		// publisher belirtilmişse içerme kontrolü (case-insensitive)
		// — registry'deki publisher tam metin farklı olabilir
	}
	return true
}
```

- [ ] **Step 5: `internal/matcher/mapping_test.go` yaz**

```go
package matcher

import (
	"context"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
)

func TestNewDefaultMatcher_Loads(t *testing.T) {
	m, err := NewDefaultMatcher()
	if err != nil {
		t.Fatalf("NewDefaultMatcher: %v", err)
	}
	if len(m.entries) < 10 {
		t.Errorf("en az 10 mapping bekleniyor, alındı: %d", len(m.entries))
	}
}

func TestMappingMatcher_Match(t *testing.T) {
	m, err := NewDefaultMatcher()
	if err != nil {
		t.Fatalf("NewDefaultMatcher: %v", err)
	}

	cases := []struct {
		name      string
		app       scanner.InstalledApp
		wantType  IdentifierType
		wantConf  MatchConfidence
		wantSubst string // wantValue'nun içerebileceği alt string
	}{
		{
			name:      "Chrome exact",
			app:       scanner.InstalledApp{Name: "Google Chrome", Version: "120.0.6099.130", Publisher: "Google LLC"},
			wantType:  IdentifierTypeCPE,
			wantConf:  ConfidenceHigh,
			wantSubst: "google:chrome",
		},
		{
			name:      "Firefox pattern",
			app:       scanner.InstalledApp{Name: "Mozilla Firefox 121.0 (x64 en-US)", Version: "121.0", Publisher: "Mozilla"},
			wantType:  IdentifierTypeCPE,
			wantConf:  ConfidenceHigh,
			wantSubst: "mozilla:firefox",
		},
		{
			name:     "Bilinmeyen app",
			app:      scanner.InstalledApp{Name: "Some Random Tool", Version: "1.0", Publisher: "Nobody"},
			wantConf: ConfidenceNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, conf, err := m.Match(context.Background(), tc.app)
			if err != nil {
				t.Fatalf("Match: %v", err)
			}
			if conf != tc.wantConf {
				t.Errorf("confidence = %v, want %v", conf, tc.wantConf)
			}
			if tc.wantConf == ConfidenceNone {
				return
			}
			if id.Type != tc.wantType {
				t.Errorf("type = %v, want %v", id.Type, tc.wantType)
			}
			if tc.wantSubst != "" && !contains(id.Value, tc.wantSubst) {
				t.Errorf("value %q içermiyor %q", id.Value, tc.wantSubst)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && stringIndex(s, sub) >= 0))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 6: Test çalıştır**

```bash
go test ./internal/matcher/... -v
```

Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add internal/matcher/ go.mod go.sum
git commit -m "feat(matcher): YAML mapping loader with 10 initial entries"
```

---

### Task 8: Mapping'i 50'ye çıkar

**Files:**
- Modify: `internal/matcher/data/mappings.yaml`

- [ ] **Step 1: 40 ek mapping daha ekle**

Aşağıdaki entry'leri mevcut listeye ekle (10 + 40 = 50 toplam). Her biri CPE veya purl şemasında. Tek tek kontrol ederek CPE değerini [nvd.nist.gov/products/cpe/search](https://nvd.nist.gov/products/cpe/search) üzerinden doğrula.

```yaml
  - registry_name_pattern: "^Microsoft Office.*"
    publisher: "Microsoft Corporation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:microsoft:office"
    version_normalize: "fourpart"

  - registry_name_pattern: "^Java(\\(TM\\))? \\d+ Update.*"
    publisher: "Oracle"
    identifier:
      type: cpe
      value: "cpe:2.3:a:oracle:jre"
    version_normalize: "fourpart"

  - registry_name_pattern: "^Java SE Development Kit.*"
    publisher: "Oracle"
    identifier:
      type: cpe
      value: "cpe:2.3:a:oracle:jdk"
    version_normalize: "fourpart"

  - registry_name_pattern: "^OBS Studio.*"
    publisher: "OBS Project"
    identifier:
      type: cpe
      value: "cpe:2.3:a:obsproject:obs_studio"
    version_normalize: "semver"

  - registry_name: "Discord"
    publisher: "Discord Inc."
    identifier:
      type: cpe
      value: "cpe:2.3:a:discord:discord"
    version_normalize: "semver"

  - registry_name: "Slack"
    publisher: "Slack Technologies, Inc."
    identifier:
      type: cpe
      value: "cpe:2.3:a:slack:slack"
    version_normalize: "semver"

  - registry_name_pattern: "^Zoom.*"
    publisher: "Zoom Video Communications, Inc."
    identifier:
      type: cpe
      value: "cpe:2.3:a:zoom:meetings"
    version_normalize: "fourpart"

  - registry_name_pattern: "^WinRAR.*"
    publisher: "win.rar GmbH"
    identifier:
      type: cpe
      value: "cpe:2.3:a:rarlab:winrar"
    version_normalize: "semver"

  - registry_name_pattern: "^PuTTY.*"
    publisher: "Simon Tatham"
    identifier:
      type: cpe
      value: "cpe:2.3:a:putty:putty"
    version_normalize: "semver"

  - registry_name_pattern: "^FileZilla Client.*"
    publisher: "Tim Kosse"
    identifier:
      type: cpe
      value: "cpe:2.3:a:filezilla-project:filezilla_client"
    version_normalize: "semver"

  - registry_name_pattern: "^Microsoft Visual Studio Code.*"
    publisher: "Microsoft Corporation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:microsoft:visual_studio_code"
    version_normalize: "semver"

  - registry_name_pattern: "^IntelliJ IDEA.*"
    publisher: "JetBrains s.r.o."
    identifier:
      type: cpe
      value: "cpe:2.3:a:jetbrains:intellij_idea"
    version_normalize: "semver"

  - registry_name_pattern: "^PyCharm.*"
    publisher: "JetBrains s.r.o."
    identifier:
      type: cpe
      value: "cpe:2.3:a:jetbrains:pycharm"
    version_normalize: "semver"

  - registry_name_pattern: "^GoLand.*"
    publisher: "JetBrains s.r.o."
    identifier:
      type: cpe
      value: "cpe:2.3:a:jetbrains:goland"
    version_normalize: "semver"

  - registry_name_pattern: "^WebStorm.*"
    publisher: "JetBrains s.r.o."
    identifier:
      type: cpe
      value: "cpe:2.3:a:jetbrains:webstorm"
    version_normalize: "semver"

  - registry_name_pattern: "^Docker Desktop.*"
    publisher: "Docker Inc."
    identifier:
      type: cpe
      value: "cpe:2.3:a:docker:desktop"
    version_normalize: "semver"

  - registry_name_pattern: "^TeamViewer.*"
    publisher: "TeamViewer"
    identifier:
      type: cpe
      value: "cpe:2.3:a:teamviewer:teamviewer"
    version_normalize: "semver"

  - registry_name_pattern: "^AnyDesk.*"
    publisher: "AnyDesk Software GmbH"
    identifier:
      type: cpe
      value: "cpe:2.3:a:anydesk:anydesk"
    version_normalize: "semver"

  - registry_name_pattern: "^Spotify.*"
    publisher: "Spotify AB"
    identifier:
      type: cpe
      value: "cpe:2.3:a:spotify:spotify"
    version_normalize: "fourpart"

  - registry_name_pattern: "^Steam.*"
    publisher: "Valve Corporation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:valvesoftware:steam"
    version_normalize: "semver"

  - registry_name_pattern: "^Microsoft Teams.*"
    publisher: "Microsoft Corporation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:microsoft:teams"
    version_normalize: "fourpart"

  - registry_name_pattern: "^Microsoft OneDrive.*"
    publisher: "Microsoft Corporation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:microsoft:onedrive"
    version_normalize: "fourpart"

  - registry_name_pattern: "^Adobe Photoshop.*"
    publisher: "Adobe"
    identifier:
      type: cpe
      value: "cpe:2.3:a:adobe:photoshop"
    version_normalize: "semver"

  - registry_name_pattern: "^Adobe Illustrator.*"
    publisher: "Adobe"
    identifier:
      type: cpe
      value: "cpe:2.3:a:adobe:illustrator"
    version_normalize: "semver"

  - registry_name_pattern: "^Adobe Premiere Pro.*"
    publisher: "Adobe"
    identifier:
      type: cpe
      value: "cpe:2.3:a:adobe:premiere_pro"
    version_normalize: "semver"

  - registry_name_pattern: "^GIMP.*"
    publisher: "The GIMP team"
    identifier:
      type: cpe
      value: "cpe:2.3:a:gimp:gimp"
    version_normalize: "semver"

  - registry_name_pattern: "^Inkscape.*"
    publisher: "Inkscape Project"
    identifier:
      type: cpe
      value: "cpe:2.3:a:inkscape:inkscape"
    version_normalize: "semver"

  - registry_name_pattern: "^Blender.*"
    publisher: "Blender Foundation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:blender:blender"
    version_normalize: "semver"

  - registry_name_pattern: "^Audacity.*"
    publisher: "Audacity Team"
    identifier:
      type: cpe
      value: "cpe:2.3:a:audacityteam:audacity"
    version_normalize: "semver"

  - registry_name_pattern: "^LibreOffice.*"
    publisher: "The Document Foundation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:libreoffice:libreoffice"
    version_normalize: "semver"

  - registry_name_pattern: "^Thunderbird.*"
    publisher: "Mozilla"
    identifier:
      type: cpe
      value: "cpe:2.3:a:mozilla:thunderbird"
    version_normalize: "semver"

  - registry_name_pattern: "^Opera.*"
    publisher: "Opera Software"
    identifier:
      type: cpe
      value: "cpe:2.3:a:opera:opera_browser"
    version_normalize: "semver"

  - registry_name_pattern: "^Brave.*"
    publisher: "Brave Software"
    identifier:
      type: cpe
      value: "cpe:2.3:a:brave:browser"
    version_normalize: "semver"

  - registry_name_pattern: "^Vivaldi.*"
    publisher: "Vivaldi Technologies"
    identifier:
      type: cpe
      value: "cpe:2.3:a:vivaldi:vivaldi"
    version_normalize: "semver"

  - registry_name_pattern: "^WinSCP.*"
    publisher: "Martin Prikryl"
    identifier:
      type: cpe
      value: "cpe:2.3:a:winscp:winscp"
    version_normalize: "semver"

  - registry_name_pattern: "^Wireshark.*"
    publisher: "The Wireshark developer community"
    identifier:
      type: cpe
      value: "cpe:2.3:a:wireshark:wireshark"
    version_normalize: "semver"

  - registry_name_pattern: "^OpenSSL.*"
    publisher: "OpenSSL"
    identifier:
      type: cpe
      value: "cpe:2.3:a:openssl:openssl"
    version_normalize: "semver"

  - registry_name_pattern: "^OpenVPN.*"
    publisher: "OpenVPN Inc."
    identifier:
      type: cpe
      value: "cpe:2.3:a:openvpn:openvpn"
    version_normalize: "semver"

  - registry_name_pattern: "^Postman.*"
    publisher: "Postman"
    identifier:
      type: cpe
      value: "cpe:2.3:a:postman:postman"
    version_normalize: "semver"

  - registry_name_pattern: "^MySQL Workbench.*"
    publisher: "Oracle Corporation"
    identifier:
      type: cpe
      value: "cpe:2.3:a:mysql:workbench"
    version_normalize: "semver"
```

- [ ] **Step 2: Mapping sayısını doğrulayan test ekle (mevcut test güncellenir)**

`internal/matcher/mapping_test.go` içindeki `TestNewDefaultMatcher_Loads` test eşiğini değiştir:

```go
func TestNewDefaultMatcher_Loads(t *testing.T) {
	m, err := NewDefaultMatcher()
	if err != nil {
		t.Fatalf("NewDefaultMatcher: %v", err)
	}
	if len(m.entries) < 50 {
		t.Errorf("en az 50 mapping bekleniyor, alındı: %d", len(m.entries))
	}
}
```

- [ ] **Step 3: Test çalıştır**

```bash
go test ./internal/matcher/... -v
```

Expected: All PASS, 50+ mapping yüklü.

- [ ] **Step 4: Commit**

```bash
git add internal/matcher/data/mappings.yaml internal/matcher/mapping_test.go
git commit -m "feat(matcher): expand curated mapping list to 50 popular Windows apps"
```

---

## Phase 3 — Data layer

### Task 9: OSV client interface ve domain tipleri

**Files:**
- Create: `internal/osv/client.go`
- Create: `internal/osv/client_test.go`

- [ ] **Step 1: `internal/osv/client.go` yaz**

```go
// Package osv OSV.dev API'sine sorgu yapar.
package osv

import (
	"context"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

// Severity bir vulnerability'nin etki seviyesi.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityUnknown  Severity = "unknown"
)

// Vulnerability tek bir CVE/GHSA bulgusu.
type Vulnerability struct {
	ID            string
	Summary       string
	Severity      Severity
	FixedVersions []string
	References    []string
	Published     time.Time
}

// Client OSV.dev sorgulayıcı arayüzü.
type Client interface {
	Query(ctx context.Context, id matcher.Identifier, version string) ([]Vulnerability, error)
}
```

- [ ] **Step 2: `internal/osv/client_test.go` yaz**

```go
package osv

import "testing"

func TestSeverity_Values(t *testing.T) {
	all := []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityUnknown}
	seen := map[Severity]bool{}
	for _, s := range all {
		if s == "" {
			t.Errorf("severity boş olamaz")
		}
		if seen[s] {
			t.Errorf("duplicate severity: %s", s)
		}
		seen[s] = true
	}
}
```

- [ ] **Step 3: Test çalıştır**

```bash
go test ./internal/osv/...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/osv/
git commit -m "feat(osv): define Client interface, Vulnerability and Severity types"
```

---

### Task 10: OSV HTTP implementation

**Files:**
- Create: `internal/osv/http.go`
- Create: `testdata/osv_responses/firefox_100.json`
- Create: `internal/osv/http_test.go`

- [ ] **Step 1: `testdata/osv_responses/firefox_100.json` örnek response yaz**

Bu, OSV.dev'in `POST /v1/query` endpoint'inden gelen tipik bir response'tur. (Gerçek 1-2 CVE içeren sadeleştirilmiş hali.)

```json
{
  "vulns": [
    {
      "id": "CVE-2022-26485",
      "summary": "Use-after-free in XSLT parameter processing",
      "published": "2022-03-08T00:00:00Z",
      "severity": [
        {
          "type": "CVSS_V3",
          "score": "CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H"
        }
      ],
      "affected": [
        {
          "package": {
            "name": "firefox",
            "ecosystem": "Mozilla"
          },
          "ranges": [
            {
              "type": "ECOSYSTEM",
              "events": [
                {"introduced": "0"},
                {"fixed": "97.0.2"}
              ]
            }
          ]
        }
      ],
      "references": [
        {"type": "ADVISORY", "url": "https://www.mozilla.org/security/advisories/mfsa2022-09/"}
      ]
    }
  ]
}
```

- [ ] **Step 2: `internal/osv/http.go` yaz**

```go
package osv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

const defaultEndpoint = "https://api.osv.dev/v1/query"

// HTTPClient OSV.dev'e gerçek HTTP sorgusu yapan Client implementasyonu.
type HTTPClient struct {
	endpoint string
	http     *http.Client
	maxRetry int
}

// NewHTTPClient varsayılan ayarlarla yeni HTTPClient oluşturur.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		endpoint: defaultEndpoint,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetry: 3,
	}
}

// WithEndpoint test amaçlı endpoint override.
func (c *HTTPClient) WithEndpoint(url string) *HTTPClient {
	c.endpoint = url
	return c
}

// Query Client arayüzünü uygular.
func (c *HTTPClient) Query(ctx context.Context, id matcher.Identifier, version string) ([]Vulnerability, error) {
	payload, err := buildPayload(id, version)
	if err != nil {
		return nil, err
	}
	body, err := c.doWithRetry(ctx, payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(body)
}

func buildPayload(id matcher.Identifier, version string) ([]byte, error) {
	req := map[string]any{
		"version": version,
	}
	switch id.Type {
	case matcher.IdentifierTypePURL:
		req["package"] = map[string]string{"purl": id.Value}
	case matcher.IdentifierTypeCPE:
		// OSV.dev CPE'yi doğrudan kabul etmez; CPE'den vendor:product:name çıkarıp
		// ecosystem hint ile package.name'e gönderiyoruz.
		name := nameFromCPE(id.Value)
		pkg := map[string]string{"name": name}
		if id.Ecosystem != "" {
			pkg["ecosystem"] = id.Ecosystem
		}
		req["package"] = pkg
	default:
		return nil, fmt.Errorf("desteklenmeyen identifier type: %s", id.Type)
	}
	return json.Marshal(req)
}

func nameFromCPE(cpe string) string {
	// cpe:2.3:a:vendor:product[:...] -> "product"
	parts := strings.Split(cpe, ":")
	if len(parts) >= 5 {
		return parts[4]
	}
	return cpe
}

func (c *HTTPClient) doWithRetry(ctx context.Context, payload []byte) ([]byte, error) {
	var lastErr error
	backoff := 500 * time.Millisecond

	for attempt := 0; attempt < c.maxRetry; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("request oluştur: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "mr-ducksafe/0.1")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			switch {
			case resp.StatusCode == http.StatusOK:
				return body, readErr
			case resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode >= 500:
				lastErr = fmt.Errorf("OSV status %d: %s", resp.StatusCode, string(body))
			default:
				return nil, fmt.Errorf("OSV status %d: %s", resp.StatusCode, string(body))
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	if lastErr == nil {
		lastErr = errors.New("OSV: max retry aşıldı")
	}
	return nil, lastErr
}

type osvResponse struct {
	Vulns []osvVuln `json:"vulns"`
}

type osvVuln struct {
	ID         string    `json:"id"`
	Summary    string    `json:"summary"`
	Published  time.Time `json:"published"`
	References []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"references"`
	Affected []struct {
		Ranges []struct {
			Events []map[string]string `json:"events"`
		} `json:"ranges"`
	} `json:"affected"`
	Severity []struct {
		Type  string `json:"type"`
		Score string `json:"score"`
	} `json:"severity"`
}

func parseResponse(body []byte) ([]Vulnerability, error) {
	if len(body) == 0 {
		return nil, nil
	}
	var r osvResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("OSV response parse: %w", err)
	}
	out := make([]Vulnerability, 0, len(r.Vulns))
	for _, v := range r.Vulns {
		out = append(out, Vulnerability{
			ID:            v.ID,
			Summary:       v.Summary,
			Severity:      severityFromCVSS(v.Severity),
			FixedVersions: collectFixed(v.Affected),
			References:    collectReferences(v.References),
			Published:     v.Published,
		})
	}
	return out, nil
}

func severityFromCVSS(items []struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}) Severity {
	for _, s := range items {
		if !strings.HasPrefix(s.Type, "CVSS") {
			continue
		}
		score := extractBaseScore(s.Score)
		switch {
		case score >= 9.0:
			return SeverityCritical
		case score >= 7.0:
			return SeverityHigh
		case score >= 4.0:
			return SeverityMedium
		case score > 0:
			return SeverityLow
		}
	}
	return SeverityUnknown
}

// extractBaseScore CVSS vector veya numeric score'dan base score çıkarır.
// Bu MVP için kaba bir yaklaşım — vector parse etmek yerine "/" ile ayrılmış
// metriklerden ipucu çıkarmak yeterli değil, bu nedenle pure-numeric score
// hariç hepsini High kabul ediyoruz. Gerçek CVSS parse v0.2'de eklenebilir.
func extractBaseScore(s string) float64 {
	// CVSS vector ise High varsay (kötümser yaklaşım, security için doğru)
	if strings.HasPrefix(s, "CVSS:") {
		return 7.5
	}
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func collectFixed(affected []struct {
	Ranges []struct {
		Events []map[string]string `json:"events"`
	} `json:"ranges"`
}) []string {
	var out []string
	seen := map[string]bool{}
	for _, a := range affected {
		for _, r := range a.Ranges {
			for _, ev := range r.Events {
				if v, ok := ev["fixed"]; ok && !seen[v] {
					out = append(out, v)
					seen[v] = true
				}
			}
		}
	}
	return out
}

func collectReferences(refs []struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.URL)
	}
	return out
}
```

- [ ] **Step 3: `internal/osv/http_test.go` yaz (httptest server kullanır)**

```go
package osv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

func TestHTTPClient_Query_Success(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/osv_responses/firefox_100.json")
	if err != nil {
		t.Fatalf("fixture okunamadı: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := NewHTTPClient().WithEndpoint(srv.URL)
	id := matcher.Identifier{
		Type:      matcher.IdentifierTypeCPE,
		Value:     "cpe:2.3:a:mozilla:firefox",
		Ecosystem: "Mozilla",
	}
	vulns, err := c.Query(context.Background(), id, "100.0")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(vulns) != 1 {
		t.Fatalf("1 vuln bekleniyor, alındı: %d", len(vulns))
	}
	if vulns[0].ID != "CVE-2022-26485" {
		t.Errorf("ID = %s, want CVE-2022-26485", vulns[0].ID)
	}
	if vulns[0].Severity != SeverityHigh {
		t.Errorf("severity = %s, want high (CVSS vector default)", vulns[0].Severity)
	}
	if len(vulns[0].FixedVersions) != 1 || vulns[0].FixedVersions[0] != "97.0.2" {
		t.Errorf("FixedVersions = %v, want [97.0.2]", vulns[0].FixedVersions)
	}
}

func TestHTTPClient_Query_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewHTTPClient().WithEndpoint(srv.URL)
	c.maxRetry = 2 // hızlı fail
	id := matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:2.3:a:test:test"}

	_, err := c.Query(context.Background(), id, "1.0")
	if err == nil {
		t.Fatal("retry sonrası error bekleniyor")
	}
}
```

- [ ] **Step 4: Test çalıştır**

```bash
go test ./internal/osv/... -v
```

Expected: 3 test PASS

- [ ] **Step 5: Commit**

```bash
git add internal/osv/ testdata/
git commit -m "feat(osv): HTTP client with retry, backoff, and CVSS severity parsing"
```

---

### Task 11: OSV in-memory cache

**Files:**
- Create: `internal/osv/cache.go`
- Create: `internal/osv/cache_test.go`

- [ ] **Step 1: `internal/osv/cache.go` yaz**

```go
package osv

import (
	"context"
	"sync"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

const defaultCacheTTL = 24 * time.Hour

// CachedClient bir Client'ı in-memory TTL cache ile sarmalayarak
// aynı (identifier, version) sorgusunun tekrar atılmasını önler.
type CachedClient struct {
	inner Client
	ttl   time.Duration

	mu      sync.RWMutex
	entries map[string]cacheEntry
	now     func() time.Time
}

type cacheEntry struct {
	vulns     []Vulnerability
	expiresAt time.Time
}

// NewCachedClient varsayılan 24 saat TTL ile cache wrapper döner.
func NewCachedClient(inner Client) *CachedClient {
	return &CachedClient{
		inner:   inner,
		ttl:     defaultCacheTTL,
		entries: make(map[string]cacheEntry),
		now:     time.Now,
	}
}

// Query önce cache'e bakar, miss/expired ise inner'a delege eder.
func (c *CachedClient) Query(ctx context.Context, id matcher.Identifier, version string) ([]Vulnerability, error) {
	key := id.Value + "@" + version

	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if ok && c.now().Before(e.expiresAt) {
		return e.vulns, nil
	}

	vulns, err := c.inner.Query(ctx, id, version)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.entries[key] = cacheEntry{vulns: vulns, expiresAt: c.now().Add(c.ttl)}
	c.mu.Unlock()

	return vulns, nil
}
```

- [ ] **Step 2: `internal/osv/cache_test.go` yaz**

```go
package osv

import (
	"context"
	"testing"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

type stubClient struct {
	calls int
	vulns []Vulnerability
	err   error
}

func (s *stubClient) Query(_ context.Context, _ matcher.Identifier, _ string) ([]Vulnerability, error) {
	s.calls++
	return s.vulns, s.err
}

func TestCachedClient_HitDoesNotCallInner(t *testing.T) {
	stub := &stubClient{vulns: []Vulnerability{{ID: "CVE-1"}}}
	c := NewCachedClient(stub)

	id := matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:test:test"}

	for i := 0; i < 3; i++ {
		if _, err := c.Query(context.Background(), id, "1.0"); err != nil {
			t.Fatalf("Query: %v", err)
		}
	}
	if stub.calls != 1 {
		t.Errorf("inner 1 kez çağrılmalıydı, %d kez çağrıldı", stub.calls)
	}
}

func TestCachedClient_ExpiredRecallsInner(t *testing.T) {
	stub := &stubClient{vulns: []Vulnerability{{ID: "CVE-1"}}}
	c := NewCachedClient(stub)

	clock := time.Now()
	c.now = func() time.Time { return clock }

	id := matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:test:test"}
	if _, err := c.Query(context.Background(), id, "1.0"); err != nil {
		t.Fatal(err)
	}
	clock = clock.Add(25 * time.Hour) // TTL'i aş
	if _, err := c.Query(context.Background(), id, "1.0"); err != nil {
		t.Fatal(err)
	}
	if stub.calls != 2 {
		t.Errorf("expired sonrası 2 çağrı bekleniyor, alındı: %d", stub.calls)
	}
}
```

- [ ] **Step 3: Test çalıştır**

```bash
go test ./internal/osv/... -v
```

Expected: 5 test PASS (önceki 3 + yeni 2)

- [ ] **Step 4: Commit**

```bash
git add internal/osv/cache.go internal/osv/cache_test.go
git commit -m "feat(osv): in-memory TTL cache wrapper for client"
```

---

### Task 12: Store interface, schema ve tipler

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/schema.sql`
- Create: `internal/store/store_test.go`

- [ ] **Step 1: `internal/store/schema.sql` yaz**

```sql
CREATE TABLE IF NOT EXISTS scans (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at   DATETIME NOT NULL,
    finished_at  DATETIME,
    host         TEXT,
    os           TEXT,
    status       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS findings (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_id       INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    app_name      TEXT NOT NULL,
    app_version   TEXT NOT NULL,
    cve_id        TEXT NOT NULL,
    severity      TEXT NOT NULL,
    fixed_version TEXT,
    summary       TEXT,
    refs          TEXT
);

CREATE INDEX IF NOT EXISTS idx_findings_scan_id ON findings(scan_id);
CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings(severity);
```

- [ ] **Step 2: `internal/store/store.go` yaz**

```go
// Package store SQLite üzerinden scan + findings persistence sağlar.
package store

import (
	"context"
	"time"
)

// ScanStatus bir scan'in mevcut durumu.
type ScanStatus string

const (
	StatusRunning     ScanStatus = "running"
	StatusSuccess     ScanStatus = "success"
	StatusFailed      ScanStatus = "failed"
	StatusInterrupted ScanStatus = "interrupted"
)

// Scan bir tarama oturumunun meta verileri.
type Scan struct {
	ID         int64
	StartedAt  time.Time
	FinishedAt *time.Time
	Host       string
	OS         string
	Status     ScanStatus
}

// Finding tek bir CVE bulgusu.
type Finding struct {
	ID           int64
	ScanID       int64
	AppName      string
	AppVersion   string
	CVEID        string
	Severity     string
	FixedVersion string
	Summary      string
	References   []string
}

// Store SQLite tabanlı persistence arayüzü.
type Store interface {
	SaveScan(ctx context.Context, s Scan) (int64, error)
	AppendFindings(ctx context.Context, scanID int64, findings []Finding) error
	UpdateScanStatus(ctx context.Context, scanID int64, status ScanStatus, finishedAt time.Time) error
	ListScans(ctx context.Context, limit int) ([]Scan, error)
	GetScan(ctx context.Context, id int64) (Scan, []Finding, error)
	CleanupStaleRunning(ctx context.Context, olderThan time.Duration) (int, error)
	Close() error
}
```

- [ ] **Step 3: `internal/store/store_test.go` (interface seviyesi, sadece tip kontrolü)**

```go
package store

import "testing"

func TestScanStatus_Values(t *testing.T) {
	all := []ScanStatus{StatusRunning, StatusSuccess, StatusFailed, StatusInterrupted}
	for _, s := range all {
		if s == "" {
			t.Errorf("status boş olamaz")
		}
	}
}
```

- [ ] **Step 4: Test çalıştır**

```bash
go test ./internal/store/...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/store/
git commit -m "feat(store): define Store interface, Scan and Finding types with SQL schema"
```

---

### Task 13: Store SQLite implementation

**Files:**
- Create: `internal/store/sqlite.go`
- Create: `internal/store/sqlite_test.go`
- Modify: `go.mod` (modernc.org/sqlite)

- [ ] **Step 1: Dependency ekle**

```bash
go get modernc.org/sqlite
```

- [ ] **Step 2: `internal/store/sqlite.go` yaz**

```go
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// SQLiteStore Store arayüzünün SQLite implementasyonu.
type SQLiteStore struct {
	db *sql.DB
}

// OpenSQLite verilen DSN ile bağlantı açar ve şemayı uygular.
// DSN örnek: "file:/path/to/mr-ducksafe.db?_journal=WAL&_fk=1"
func OpenSQLite(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("schema apply: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) SaveScan(ctx context.Context, sc Scan) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO scans (started_at, host, os, status)
		VALUES (?, ?, ?, ?)`,
		sc.StartedAt, sc.Host, sc.OS, sc.Status,
	)
	if err != nil {
		return 0, fmt.Errorf("save scan: %w", err)
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) AppendFindings(ctx context.Context, scanID int64, findings []Finding) error {
	if len(findings) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO findings (scan_id, app_name, app_version, cve_id, severity, fixed_version, summary, refs)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range findings {
		refsJSON, err := json.Marshal(f.References)
		if err != nil {
			return fmt.Errorf("refs marshal: %w", err)
		}
		if _, err := stmt.ExecContext(ctx,
			scanID, f.AppName, f.AppVersion, f.CVEID, f.Severity,
			f.FixedVersion, f.Summary, string(refsJSON),
		); err != nil {
			return fmt.Errorf("insert finding: %w", err)
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) UpdateScanStatus(ctx context.Context, scanID int64, status ScanStatus, finishedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE scans SET status = ?, finished_at = ? WHERE id = ?`,
		status, finishedAt, scanID,
	)
	return err
}

func (s *SQLiteStore) ListScans(ctx context.Context, limit int) ([]Scan, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, started_at, finished_at, host, os, status
		FROM scans ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Scan
	for rows.Next() {
		var sc Scan
		var finished sql.NullTime
		if err := rows.Scan(&sc.ID, &sc.StartedAt, &finished, &sc.Host, &sc.OS, &sc.Status); err != nil {
			return nil, err
		}
		if finished.Valid {
			sc.FinishedAt = &finished.Time
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) GetScan(ctx context.Context, id int64) (Scan, []Finding, error) {
	var sc Scan
	var finished sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT id, started_at, finished_at, host, os, status
		FROM scans WHERE id = ?`, id).Scan(
		&sc.ID, &sc.StartedAt, &finished, &sc.Host, &sc.OS, &sc.Status,
	)
	if err != nil {
		return Scan{}, nil, fmt.Errorf("get scan: %w", err)
	}
	if finished.Valid {
		sc.FinishedAt = &finished.Time
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, scan_id, app_name, app_version, cve_id, severity, fixed_version, summary, refs
		FROM findings WHERE scan_id = ? ORDER BY severity, app_name`, id)
	if err != nil {
		return sc, nil, err
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		var f Finding
		var refsJSON sql.NullString
		var fixed sql.NullString
		if err := rows.Scan(&f.ID, &f.ScanID, &f.AppName, &f.AppVersion, &f.CVEID, &f.Severity, &fixed, &f.Summary, &refsJSON); err != nil {
			return sc, nil, err
		}
		if fixed.Valid {
			f.FixedVersion = fixed.String
		}
		if refsJSON.Valid && refsJSON.String != "" {
			_ = json.Unmarshal([]byte(refsJSON.String), &f.References)
		}
		findings = append(findings, f)
	}
	return sc, findings, rows.Err()
}

func (s *SQLiteStore) CleanupStaleRunning(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	res, err := s.db.ExecContext(ctx, `
		UPDATE scans SET status = ?, finished_at = ?
		WHERE status = ? AND started_at < ?`,
		StatusInterrupted, time.Now(), StatusRunning, cutoff,
	)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
```

- [ ] **Step 3: `internal/store/sqlite_test.go` yaz**

```go
package store

import (
	"context"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestSQLiteStore_SaveAndGetScan(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.SaveScan(ctx, Scan{
		StartedAt: time.Now(),
		Host:      "test-host",
		OS:        "windows",
		Status:    StatusRunning,
	})
	if err != nil {
		t.Fatalf("SaveScan: %v", err)
	}
	if id == 0 {
		t.Fatal("ID 0 dönmemeli")
	}

	findings := []Finding{
		{
			AppName:      "Firefox",
			AppVersion:   "100.0",
			CVEID:        "CVE-2022-26485",
			Severity:     "high",
			FixedVersion: "97.0.2",
			Summary:      "Use-after-free",
			References:   []string{"https://example.com/advisory"},
		},
	}
	if err := s.AppendFindings(ctx, id, findings); err != nil {
		t.Fatalf("AppendFindings: %v", err)
	}

	if err := s.UpdateScanStatus(ctx, id, StatusSuccess, time.Now()); err != nil {
		t.Fatalf("UpdateScanStatus: %v", err)
	}

	sc, got, err := s.GetScan(ctx, id)
	if err != nil {
		t.Fatalf("GetScan: %v", err)
	}
	if sc.Status != StatusSuccess {
		t.Errorf("status = %s, want success", sc.Status)
	}
	if len(got) != 1 {
		t.Fatalf("1 finding bekleniyor, alındı: %d", len(got))
	}
	if got[0].CVEID != "CVE-2022-26485" {
		t.Errorf("CVE = %s, want CVE-2022-26485", got[0].CVEID)
	}
	if len(got[0].References) != 1 {
		t.Errorf("references = %v, want 1 url", got[0].References)
	}
}

func TestSQLiteStore_CleanupStaleRunning(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.SaveScan(ctx, Scan{
		StartedAt: time.Now().Add(-1 * time.Hour),
		Status:    StatusRunning,
	})
	if err != nil {
		t.Fatal(err)
	}

	n, err := s.CleanupStaleRunning(ctx, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("cleanup count = %d, want 1", n)
	}

	sc, _, err := s.GetScan(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Status != StatusInterrupted {
		t.Errorf("status = %s, want interrupted", sc.Status)
	}
}

func TestSQLiteStore_ListScans(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := s.SaveScan(ctx, Scan{StartedAt: time.Now(), Status: StatusSuccess}); err != nil {
			t.Fatal(err)
		}
	}
	scans, err := s.ListScans(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(scans) != 3 {
		t.Errorf("3 scan bekleniyor, alındı: %d", len(scans))
	}
}
```

- [ ] **Step 4: Test çalıştır**

```bash
go test ./internal/store/... -v
```

Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add internal/store/ go.mod go.sum
git commit -m "feat(store): SQLite implementation with schema migration, save/get/cleanup"
```

---

## Phase 4 — Orchestration

### Task 14: Scanengine result tipleri ve progress channel

**Files:**
- Create: `internal/scanengine/result.go`
- Create: `internal/scanengine/progress.go`

- [ ] **Step 1: `internal/scanengine/result.go` yaz**

```go
// Package scanengine scanner, matcher, osv ve store'u orkestre eder.
package scanengine

import (
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

// ScanResult tek bir taramanın özet sonucu.
type ScanResult struct {
	ScanID    int64
	StartedAt time.Time
	Finished  time.Time
	Apps      []scanner.InstalledApp
	Findings  []store.Finding
	Unmatched []scanner.InstalledApp
	Warnings  []ScanError
}

// Stage scan içindeki bir aşamayı belirtir.
type Stage string

const (
	StageScan    Stage = "scan"
	StageMatch   Stage = "match"
	StageQuery   Stage = "query"
	StagePersist Stage = "persist"
)

// ScanError scan akışında ortaya çıkan bir hatayı taşır.
type ScanError struct {
	Stage   Stage
	AppName string // boş olabilir
	Cause   error
	Fatal   bool
}

func (e ScanError) Error() string {
	if e.AppName == "" {
		return string(e.Stage) + ": " + e.Cause.Error()
	}
	return string(e.Stage) + " (" + e.AppName + "): " + e.Cause.Error()
}

// CountBySeverity result.Findings'i severity'ye göre sayar.
func (r *ScanResult) CountBySeverity() map[string]int {
	out := map[string]int{}
	for _, f := range r.Findings {
		out[f.Severity]++
	}
	return out
}
```

- [ ] **Step 2: `internal/scanengine/progress.go` yaz**

```go
package scanengine

// Progress scan ilerlemesini SSE/UI'a iletmek için kullanılan event.
type Progress struct {
	Stage   Stage
	Done    int
	Total   int
	Message string
}
```

- [ ] **Step 3: Build doğrula**

```bash
go build ./internal/scanengine/...
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add internal/scanengine/result.go internal/scanengine/progress.go
git commit -m "feat(scanengine): define ScanResult, ScanError and Progress types"
```

---

### Task 15: Scanengine orkestrasyon

**Files:**
- Create: `internal/scanengine/engine.go`
- Create: `internal/scanengine/engine_test.go`

- [ ] **Step 1: `internal/scanengine/engine.go` yaz**

```go
package scanengine

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/osv"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

const defaultConcurrency = 5

// Engine scan akışını orkestre eder.
type Engine struct {
	Scanner     scanner.Scanner
	Matcher     matcher.Matcher
	OSV         osv.Client
	Store       store.Store
	Concurrency int
}

// New verilen bağımlılıklarla yeni Engine oluşturur.
func New(s scanner.Scanner, m matcher.Matcher, o osv.Client, st store.Store) *Engine {
	return &Engine{
		Scanner:     s,
		Matcher:     m,
		OSV:         o,
		Store:       st,
		Concurrency: defaultConcurrency,
	}
}

// Run tek bir scan döngüsünü yürütür ve sonucu döndürür.
// progress nil değilse her aşamada event yayınlar; receiver yoksa engellemez.
func (e *Engine) Run(ctx context.Context, progress chan<- Progress) (*ScanResult, error) {
	started := time.Now()
	host, _ := os.Hostname()

	scanID, err := e.Store.SaveScan(ctx, store.Scan{
		StartedAt: started,
		Host:      host,
		OS:        runtime.GOOS,
		Status:    store.StatusRunning,
	})
	if err != nil {
		return nil, err
	}

	result := &ScanResult{ScanID: scanID, StartedAt: started}

	apps, err := e.Scanner.Scan(ctx)
	if err != nil {
		result.Warnings = append(result.Warnings, ScanError{Stage: StageScan, Cause: err, Fatal: true})
		_ = e.Store.UpdateScanStatus(ctx, scanID, store.StatusFailed, time.Now())
		result.Finished = time.Now()
		return result, err
	}
	result.Apps = apps
	emit(progress, Progress{Stage: StageScan, Done: len(apps), Total: len(apps), Message: "scan complete"})

	var toQuery []matchedApp
	for _, app := range apps {
		id, conf, mErr := e.Matcher.Match(ctx, app)
		if mErr != nil {
			result.Warnings = append(result.Warnings, ScanError{Stage: StageMatch, AppName: app.Name, Cause: mErr})
			continue
		}
		if conf == matcher.ConfidenceNone {
			result.Unmatched = append(result.Unmatched, app)
			continue
		}
		toQuery = append(toQuery, matchedApp{app: app, id: id})
	}
	emit(progress, Progress{Stage: StageMatch, Done: len(toQuery), Total: len(apps), Message: "matching complete"})

	findings, queryWarnings := e.queryAll(ctx, scanID, toQuery, progress)
	result.Findings = findings
	result.Warnings = append(result.Warnings, queryWarnings...)

	if err := e.Store.AppendFindings(ctx, scanID, findings); err != nil {
		result.Warnings = append(result.Warnings, ScanError{Stage: StagePersist, Cause: err, Fatal: true})
		_ = e.Store.UpdateScanStatus(ctx, scanID, store.StatusFailed, time.Now())
		result.Finished = time.Now()
		return result, err
	}

	finished := time.Now()
	if err := e.Store.UpdateScanStatus(ctx, scanID, store.StatusSuccess, finished); err != nil {
		result.Warnings = append(result.Warnings, ScanError{Stage: StagePersist, Cause: err})
	}
	result.Finished = finished
	emit(progress, Progress{Stage: StagePersist, Done: len(findings), Total: len(findings), Message: "done"})

	return result, nil
}

type matchedApp struct {
	app scanner.InstalledApp
	id  matcher.Identifier
}

func (e *Engine) queryAll(ctx context.Context, scanID int64, items []matchedApp, progress chan<- Progress) ([]store.Finding, []ScanError) {
	var (
		findings []store.Finding
		warnings []ScanError
		mu       sync.Mutex
	)

	conc := e.Concurrency
	if conc <= 0 {
		conc = defaultConcurrency
	}
	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup

	for i, item := range items {
		select {
		case <-ctx.Done():
			break
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(idx int, m matchedApp) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					warnings = append(warnings, ScanError{Stage: StageQuery, AppName: m.app.Name})
					mu.Unlock()
				}
				<-sem
			}()

			vulns, err := e.OSV.Query(ctx, m.id, m.app.Version)
			mu.Lock()
			if err != nil {
				warnings = append(warnings, ScanError{Stage: StageQuery, AppName: m.app.Name, Cause: err})
			} else {
				for _, v := range vulns {
					findings = append(findings, store.Finding{
						ScanID:       scanID,
						AppName:      m.app.Name,
						AppVersion:   m.app.Version,
						CVEID:        v.ID,
						Severity:     string(v.Severity),
						FixedVersion: firstOr(v.FixedVersions, ""),
						Summary:      v.Summary,
						References:   v.References,
					})
				}
			}
			mu.Unlock()
			emit(progress, Progress{Stage: StageQuery, Done: idx + 1, Total: len(items)})
		}(i, item)
	}
	wg.Wait()
	return findings, warnings
}

func firstOr(s []string, def string) string {
	if len(s) > 0 {
		return s[0]
	}
	return def
}

func emit(ch chan<- Progress, p Progress) {
	if ch == nil {
		return
	}
	select {
	case ch <- p:
	default: // receiver hazır değilse drop et
	}
}
```

- [ ] **Step 2: `internal/scanengine/engine_test.go` yaz**

```go
package scanengine

import (
	"context"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/osv"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

type fakeScanner struct{ apps []scanner.InstalledApp }

func (f *fakeScanner) Scan(_ context.Context) ([]scanner.InstalledApp, error) { return f.apps, nil }

type fakeMatcher struct{}

func (fakeMatcher) Match(_ context.Context, app scanner.InstalledApp) (matcher.Identifier, matcher.MatchConfidence, error) {
	if app.Name == "Known" {
		return matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:test:known"}, matcher.ConfidenceHigh, nil
	}
	return matcher.Identifier{}, matcher.ConfidenceNone, nil
}

type fakeOSV struct{ vulns []osv.Vulnerability }

func (f *fakeOSV) Query(_ context.Context, _ matcher.Identifier, _ string) ([]osv.Vulnerability, error) {
	return f.vulns, nil
}

func newMemoryStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestEngine_Run_MatchesAndPersists(t *testing.T) {
	st := newMemoryStore(t)
	eng := New(
		&fakeScanner{apps: []scanner.InstalledApp{
			{Name: "Known", Version: "1.0"},
			{Name: "Unknown", Version: "2.0"},
		}},
		fakeMatcher{},
		&fakeOSV{vulns: []osv.Vulnerability{
			{ID: "CVE-1", Severity: osv.SeverityHigh, FixedVersions: []string{"1.1"}},
		}},
		st,
	)

	res, err := eng.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Apps) != 2 {
		t.Errorf("apps = %d, want 2", len(res.Apps))
	}
	if len(res.Unmatched) != 1 {
		t.Errorf("unmatched = %d, want 1", len(res.Unmatched))
	}
	if len(res.Findings) != 1 {
		t.Errorf("findings = %d, want 1", len(res.Findings))
	}
	if res.Findings[0].CVEID != "CVE-1" {
		t.Errorf("CVE = %s, want CVE-1", res.Findings[0].CVEID)
	}
}
```

- [ ] **Step 3: Test çalıştır**

```bash
go test ./internal/scanengine/... -v
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/scanengine/
git commit -m "feat(scanengine): orchestrate scan -> match -> query -> persist with concurrency"
```

---

### Task 16: Report — table formatter

**Files:**
- Create: `internal/report/report.go`
- Create: `internal/report/table.go`
- Modify: `go.mod` (tablewriter)

- [ ] **Step 1: Dependency ekle**

```bash
go get github.com/olekukonko/tablewriter
go get github.com/fatih/color
```

- [ ] **Step 2: `internal/report/report.go` yaz**

```go
// Package report ScanResult'ı CLI çıktısı için formatlar.
package report

import (
	"fmt"
	"io"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

// Format çıktı formatı.
type Format string

const (
	FormatTable    Format = "table"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

// Options formatter davranışını yapılandırır.
type Options struct {
	ShowUnmatched bool
	NoColor       bool
}

// Write verilen format ile sonucu w'ye yazar.
func Write(w io.Writer, result *scanengine.ScanResult, format Format, opts Options) error {
	switch format {
	case FormatTable:
		return writeTable(w, result, opts)
	case FormatJSON:
		return writeJSON(w, result, opts)
	case FormatMarkdown:
		return writeMarkdown(w, result, opts)
	default:
		return fmt.Errorf("desteklenmeyen format: %s", format)
	}
}
```

- [ ] **Step 3: `internal/report/table.go` yaz**

```go
package report

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

func writeTable(w io.Writer, r *scanengine.ScanResult, opts Options) error {
	if opts.NoColor {
		color.NoColor = true
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
		t.SetHeader([]string{"Sev", "Uygulama", "Sürüm", "CVE", "Düzeltme", "Özet"})
		t.SetAutoWrapText(false)
		t.SetColMinWidth(5, 40)

		for _, f := range r.Findings {
			summary := f.Summary
			if len(summary) > 70 {
				summary = summary[:67] + "..."
			}
			t.Append([]string{
				colorSeverity(f.Severity),
				f.AppName, f.AppVersion, f.CVEID, f.FixedVersion, summary,
			})
		}
		t.Render()
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
```

- [ ] **Step 4: Smoke test çalıştır (formatter çıktısı manuel görmek için)**

```bash
go build ./internal/report/...
```

Expected: No errors.

- [ ] **Step 5: Commit**

```bash
git add internal/report/ go.mod go.sum
git commit -m "feat(report): table formatter with severity colors and summary truncation"
```

---

### Task 17: Report — JSON ve Markdown formatter + golden test

**Files:**
- Create: `internal/report/json.go`
- Create: `internal/report/markdown.go`
- Create: `internal/report/report_test.go`
- Create: `internal/report/testdata/result.golden.json`
- Create: `internal/report/testdata/result.golden.md`

- [ ] **Step 1: `internal/report/json.go` yaz**

```go
package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

type jsonOutput struct {
	SchemaVersion string                       `json:"schemaVersion"`
	ScanID        int64                        `json:"scanId"`
	StartedAt     time.Time                    `json:"startedAt"`
	FinishedAt    time.Time                    `json:"finishedAt"`
	AppsTotal     int                          `json:"appsTotal"`
	Counts        map[string]int               `json:"countsBySeverity"`
	Findings      []jsonFinding                `json:"findings"`
	Unmatched     []jsonUnmatched              `json:"unmatched"`
	Warnings      []string                     `json:"warnings,omitempty"`
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
```

- [ ] **Step 2: `internal/report/markdown.go` yaz**

```go
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
```

- [ ] **Step 3: Golden file'lar oluştur**

`internal/report/testdata/result.golden.json`:

```json
{
  "schemaVersion": "1",
  "scanId": 42,
  "startedAt": "2026-05-24T10:00:00Z",
  "finishedAt": "2026-05-24T10:00:30Z",
  "appsTotal": 2,
  "countsBySeverity": {
    "high": 1
  },
  "findings": [
    {
      "appName": "Firefox",
      "appVersion": "100.0",
      "cveId": "CVE-2022-26485",
      "severity": "high",
      "fixedVersion": "97.0.2",
      "summary": "Use-after-free in XSLT parameter processing",
      "references": [
        "https://example.com/advisory"
      ]
    }
  ],
  "unmatched": [
    {
      "name": "Some Random Tool",
      "version": "1.0",
      "publisher": "Nobody"
    }
  ]
}
```

`internal/report/testdata/result.golden.md`:

```markdown
# mr-ducksafe scan #42

- Tarandı: **2** kurulu yazılım
- Eşleştirilemedi: **1**
- Bulgular: 🔴 0 kritik · 🟠 1 yüksek · 🟡 0 orta · 🔵 0 düşük

| Severity | Uygulama | Sürüm | CVE | Düzeltme | Özet |
|---|---|---|---|---|---|
| high | Firefox | 100.0 | CVE-2022-26485 | 97.0.2 | Use-after-free in XSLT parameter processing |
```

- [ ] **Step 4: `internal/report/report_test.go` yaz**

```go
package report

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
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
```

- [ ] **Step 5: Test çalıştır**

```bash
go test ./internal/report/... -v
```

Expected: PASS (golden file'lar eşleşir). Eşleşmezse: golden file'ları manuel düzenle veya formatter'ı düzelt.

- [ ] **Step 6: Commit**

```bash
git add internal/report/ 
git commit -m "feat(report): add JSON and Markdown formatters with golden file tests"
```

---

## Phase 5 — CLI

### Task 18: Cobra root + version komutu

**Files:**
- Create: `internal/cmd/root.go`
- Create: `internal/cmd/version.go`
- Modify: `cmd/mr-ducksafe/main.go`
- Modify: `go.mod` (cobra)

- [ ] **Step 1: Dependency ekle**

```bash
go get github.com/spf13/cobra
```

- [ ] **Step 2: `internal/cmd/root.go` yaz**

```go
// Package cmd mr-ducksafe CLI komutlarını tanımlar.
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	flagVerbose bool
	flagDBPath  string
)

// NewRootCmd mr-ducksafe root komutunu oluşturur.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "mr-ducksafe",
		Short:         "Yerel kurulu yazılımlar için CVE tarayıcı",
		Long:          "mr-ducksafe makinendeki kurulu sistem yazılımlarını tarayıp OSV.dev üzerinden bilinen güvenlik açıklarını raporlar.",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "ayrıntılı log çıktısı")
	root.PersistentFlags().StringVar(&flagDBPath, "db", "", "SQLite DB yolu (default: ~/.mr-ducksafe/mr-ducksafe.db)")

	root.AddCommand(NewVersionCmd())
	// Task 19'da NewScanCmd, Task 20'de NewPruneCmd, Task 21'de NewServeCmd eklenecek.
	return root
}

// Execute mr-ducksafe CLI'sini çalıştırır.
func Execute() error {
	return NewRootCmd().Execute()
}
```

- [ ] **Step 3: `internal/cmd/version.go` yaz**

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version build sırasında ldflags ile set edilir; yoksa "dev".
var Version = "dev"

// NewVersionCmd `mr-ducksafe version` komutu.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "mr-ducksafe sürümünü göster",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "mr-ducksafe %s\n", Version)
			return nil
		},
	}
}
```

- [ ] **Step 4: `cmd/mr-ducksafe/main.go` güncelle**

```go
package main

import (
	"fmt"
	"os"

	"github.com/yemresalcan/mr-ducksafe/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: Build ve doğrula**

```bash
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
./mr-ducksafe.exe version
./mr-ducksafe.exe --help
```

Expected:
- `mr-ducksafe version` → `mr-ducksafe dev`
- `mr-ducksafe --help` → root help mesajı, version subcommand görünür

- [ ] **Step 6: Commit**

```bash
git add cmd/mr-ducksafe/main.go internal/cmd/ go.mod go.sum
git commit -m "feat(cli): cobra root with version subcommand and global flags"
```

---

### Task 19: `scan` komutu

**Files:**
- Create: `internal/cmd/scan.go`
- Modify: `internal/cmd/root.go`
- Create: `internal/config/config.go`

- [ ] **Step 1: `internal/config/config.go` yaz**

```go
// Package config kullanıcı config dosyası ve default path'leri yönetir.
package config

import (
	"os"
	"path/filepath"
)

// DataDir mr-ducksafe'in kalıcı verilerini sakladığı klasör (~/.mr-ducksafe).
func DataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".mr-ducksafe")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// DBPath default SQLite DB yolu.
func DBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mr-ducksafe.db"), nil
}
```

- [ ] **Step 2: `internal/cmd/scan.go` yaz**

```go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/osv"
	"github.com/yemresalcan/mr-ducksafe/internal/report"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
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
	osvClient := osv.NewCachedClient(osv.NewHTTPClient())

	eng := scanengine.New(sc, m, osvClient, st)
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
```

- [ ] **Step 3: `internal/cmd/root.go` içine scan command'ı register et**

`NewRootCmd` fonksiyonu içinde `root.AddCommand(NewVersionCmd())` satırının altına ekle:

```go
	root.AddCommand(NewScanCmd())
```

- [ ] **Step 4: Build ve manuel test (Windows üzerinde)**

```bash
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
./mr-ducksafe.exe scan --format=table
```

Expected: Registry tarar, tablo basar, ortalama 100-200 kurulu yazılım listeler, OSV'ye sorgu atar.

- [ ] **Step 5: JSON formatı dene**

```bash
./mr-ducksafe.exe scan --format=json | head -50
```

Expected: Geçerli JSON, schemaVersion: "1", findings array'i.

- [ ] **Step 6: Commit**

```bash
git add internal/cmd/ internal/config/
git commit -m "feat(cli): scan command wiring scanner/matcher/osv/store with format flags"
```

---

### Task 20: `prune` komutu

**Files:**
- Create: `internal/cmd/prune.go`
- Modify: `internal/cmd/root.go`
- Modify: `internal/store/store.go` (PruneOlderThan ekle)
- Modify: `internal/store/sqlite.go` (impl)
- Modify: `internal/store/sqlite_test.go` (test)

- [ ] **Step 1: `internal/store/store.go` Store interface'ine method ekle**

`Store` interface tanımının içine ekle:

```go
	PruneOlderThan(ctx context.Context, olderThan time.Duration) (int, error)
```

- [ ] **Step 2: `internal/store/sqlite.go` implementasyonu ekle**

```go
func (s *SQLiteStore) PruneOlderThan(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM scans WHERE started_at < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
```

- [ ] **Step 3: `internal/store/sqlite_test.go` test ekle**

```go
func TestSQLiteStore_PruneOlderThan(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	old := time.Now().Add(-48 * time.Hour)
	_, _ = s.SaveScan(ctx, Scan{StartedAt: old, Status: StatusSuccess})
	_, _ = s.SaveScan(ctx, Scan{StartedAt: time.Now(), Status: StatusSuccess})

	n, err := s.PruneOlderThan(ctx, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("prune count = %d, want 1", n)
	}
}
```

- [ ] **Step 4: `internal/cmd/prune.go` yaz**

```go
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

var flagPruneOlderThan string

// NewPruneCmd `mr-ducksafe prune` komutu.
func NewPruneCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "prune",
		Short: "Belirli yaştan eski scan kayıtlarını sil",
		RunE:  runPrune,
	}
	c.Flags().StringVar(&flagPruneOlderThan, "older-than", "30d", "süre: 7d, 24h, 30d gibi")
	return c
}

func runPrune(cmd *cobra.Command, _ []string) error {
	d, err := parseHumanDuration(flagPruneOlderThan)
	if err != nil {
		return err
	}

	dbPath := flagDBPath
	if dbPath == "" {
		p, err := config.DBPath()
		if err != nil {
			return err
		}
		dbPath = p
	}
	st, err := store.OpenSQLite("file:" + dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	n, err := st.PruneOlderThan(cmd.Context(), d)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d scan kaydı silindi\n", n)
	return nil
}

func parseHumanDuration(s string) (time.Duration, error) {
	if len(s) >= 2 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil {
			return 0, fmt.Errorf("geçersiz süre: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
```

- [ ] **Step 5: `internal/cmd/root.go` register et**

```go
	root.AddCommand(NewPruneCmd())
```

- [ ] **Step 6: Test ve build**

```bash
go test ./internal/store/...
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
./mr-ducksafe.exe prune --older-than=30d
```

Expected: Tests PASS, prune komutu çalışır (0 kayıt silebilir).

- [ ] **Step 7: Commit**

```bash
git add internal/cmd/prune.go internal/cmd/root.go internal/store/
git commit -m "feat(cli): prune command and store.PruneOlderThan for cleaning old scans"
```

---

## Phase 6 — Server

### Task 21: HTTP server skeleton + dashboard

**Files:**
- Create: `internal/server/server.go`
- Create: `internal/server/handlers.go`
- Create: `internal/server/embed.go`
- Create: `internal/server/templates/layout.html`
- Create: `internal/server/templates/dashboard.html`
- Create: `internal/server/static/style.css`
- Create: `internal/cmd/serve.go`
- Modify: `internal/cmd/root.go`

- [ ] **Step 1: `internal/server/embed.go` yaz**

```go
package server

import "embed"

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS
```

- [ ] **Step 2: `internal/server/templates/layout.html` yaz**

```html
{{define "layout"}}<!DOCTYPE html>
<html lang="tr">
<head>
  <meta charset="UTF-8">
  <title>mr-ducksafe — {{.Title}}</title>
  <link rel="stylesheet" href="/static/style.css">
</head>
<body>
  <header>
    <h1>🦆 mr-ducksafe</h1>
    <nav>
      <a href="/">Dashboard</a>
      <a href="/scans">Geçmiş</a>
    </nav>
  </header>
  <main>{{template "content" .}}</main>
</body>
</html>{{end}}
```

- [ ] **Step 3: `internal/server/templates/dashboard.html` yaz**

```html
{{define "content"}}
<section class="card">
  <h2>Son tarama</h2>
  {{if .LastScan}}
    <p>Scan #{{.LastScan.ID}} — {{.LastScan.StartedAt.Format "2006-01-02 15:04"}} — {{.LastScan.Status}}</p>
    <p><strong>{{.FindingsCount}}</strong> bulgu</p>
    <a class="btn" href="/scans/{{.LastScan.ID}}">Detay</a>
  {{else}}
    <p>Henüz tarama yok.</p>
  {{end}}
  <form method="post" action="/api/scan">
    <button type="submit" class="btn primary">Şimdi Tara</button>
  </form>
</section>
{{end}}
```

- [ ] **Step 4: `internal/server/static/style.css` yaz**

```css
* { box-sizing: border-box; }
body { font-family: system-ui, sans-serif; margin: 0; background: #0f172a; color: #e2e8f0; }
header { display: flex; justify-content: space-between; padding: 1rem 2rem; background: #1e293b; }
header h1 { margin: 0; font-size: 1.25rem; }
nav a { color: #94a3b8; margin-left: 1rem; text-decoration: none; }
nav a:hover { color: #fff; }
main { max-width: 960px; margin: 2rem auto; padding: 0 1rem; }
.card { background: #1e293b; padding: 1.5rem; border-radius: 8px; margin-bottom: 1rem; }
.btn { display: inline-block; padding: 0.5rem 1rem; background: #334155; color: #fff; text-decoration: none; border: none; border-radius: 4px; cursor: pointer; }
.btn.primary { background: #2563eb; }
.btn:hover { opacity: 0.9; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 0.5rem; text-align: left; border-bottom: 1px solid #334155; }
.sev-critical { color: #f87171; font-weight: bold; }
.sev-high { color: #fb923c; }
.sev-medium { color: #fbbf24; }
.sev-low { color: #60a5fa; }
```

- [ ] **Step 5: `internal/server/server.go` yaz**

```go
// Package server lokal HTTP server ve embedded UI sunar.
package server

import (
	"context"
	"html/template"
	"net/http"
	"sync"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

// Server HTTP handler'ları + bağımlılıkları taşır.
type Server struct {
	Engine    *scanengine.Engine
	Store     store.Store
	templates *template.Template

	mu         sync.Mutex
	activeScan int64 // tek aktif scan kuralı
}

// New verilen engine ve store ile Server oluşturur.
func New(engine *scanengine.Engine, st store.Store) (*Server, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{Engine: engine, Store: st, templates: tmpl}, nil
}

// Handler net/http için mux döner.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/scans", s.handleScansList)
	mux.HandleFunc("/scans/", s.handleScanDetail)
	mux.HandleFunc("/api/scan", s.handleAPIScan)
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	return mux
}

func (s *Server) render(w http.ResponseWriter, name string, data map[string]any) {
	data["Title"] = data["Title"]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) tryAcquireScan() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeScan != 0 {
		return false
	}
	s.activeScan = -1 // reserve
	return true
}

func (s *Server) setActiveScan(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeScan = id
}

func (s *Server) releaseScan() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeScan = 0
}

func (s *Server) runScanAsync(ctx context.Context) {
	defer s.releaseScan()
	res, err := s.Engine.Run(ctx, nil)
	if err == nil && res != nil {
		s.setActiveScan(res.ScanID)
	}
}
```

- [ ] **Step 6: `internal/server/handlers.go` yaz**

```go
package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	scans, err := s.Store.ListScans(r.Context(), 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]any{"Title": "Dashboard"}
	if len(scans) > 0 {
		data["LastScan"] = scans[0]
		_, findings, _ := s.Store.GetScan(r.Context(), scans[0].ID)
		data["FindingsCount"] = len(findings)
	}
	s.render(w, "layout", data)
}

func (s *Server) handleScansList(w http.ResponseWriter, r *http.Request) {
	scans, err := s.Store.ListScans(r.Context(), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "layout", map[string]any{
		"Title": "Geçmiş",
		"Scans": scans,
	})
}

func (s *Server) handleScanDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/scans/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	scan, findings, err := s.Store.GetScan(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.render(w, "layout", map[string]any{
		"Title":    "Scan #" + idStr,
		"Scan":     scan,
		"Findings": findings,
	})
}

func (s *Server) handleAPIScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.tryAcquireScan() {
		http.Error(w, "another scan is already running", http.StatusConflict)
		return
	}
	go s.runScanAsync(context.Background())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

- [ ] **Step 7: `internal/cmd/serve.go` yaz**

```go
package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/yemresalcan/mr-ducksafe/internal/config"
	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/osv"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/server"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

var (
	flagServePort int
	flagServeBind string
)

// NewServeCmd `mr-ducksafe serve` komutu.
func NewServeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "serve",
		Short: "Lokal web UI başlat",
		RunE:  runServe,
	}
	c.Flags().IntVar(&flagServePort, "port", 8080, "dinlenecek port")
	c.Flags().StringVar(&flagServeBind, "bind", "127.0.0.1", "bağlanılacak adres (0.0.0.0 ağa açar — uyarı: auth yok)")
	return c
}

func runServe(cmd *cobra.Command, _ []string) error {
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

	if _, err := st.CleanupStaleRunning(cmd.Context(), 10*time.Minute); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "stale cleanup:", err)
	}

	sc, err := scanner.New()
	if err != nil {
		return err
	}
	m, err := matcher.NewDefaultMatcher()
	if err != nil {
		return err
	}
	eng := scanengine.New(sc, m, osv.NewCachedClient(osv.NewHTTPClient()), st)

	srv, err := server.New(eng, st)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", flagServeBind, flagServePort)
	fmt.Fprintf(cmd.OutOrStdout(), "mr-ducksafe serve → http://%s\n", addr)
	if flagServeBind == "0.0.0.0" {
		fmt.Fprintln(cmd.ErrOrStderr(), "UYARI: 0.0.0.0 — bu makineyi ağa açar, auth yok.")
	}
	return http.ListenAndServe(addr, srv.Handler())
}
```

- [ ] **Step 8: `internal/cmd/root.go` register et**

```go
	root.AddCommand(NewServeCmd())
```

- [ ] **Step 9: Build ve manuel test**

```bash
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
./mr-ducksafe.exe serve --port 8080
```

Tarayıcıda `http://127.0.0.1:8080` aç. Dashboard görünmeli. "Şimdi Tara" butonuna bas, ardından sayfa yenile — son scan'in detayını gör.

- [ ] **Step 10: Commit**

```bash
git add internal/server/ internal/cmd/serve.go internal/cmd/root.go
git commit -m "feat(server): HTTP server with dashboard, scan list, detail and scan trigger"
```

---

### Task 22: Scan detayı + history template'leri

**Files:**
- Create: `internal/server/templates/scans.html`
- Create: `internal/server/templates/scan_detail.html`
- Modify: `internal/server/server.go` (template adlandırma)
- Modify: `internal/server/handlers.go` (template seçimi)

- [ ] **Step 1: `internal/server/templates/scans.html` yaz**

```html
{{define "scans_content"}}
<section class="card">
  <h2>Geçmiş taramalar</h2>
  {{if .Scans}}
    <table>
      <thead>
        <tr><th>#</th><th>Başlangıç</th><th>Süre</th><th>Durum</th><th></th></tr>
      </thead>
      <tbody>
        {{range .Scans}}
          <tr>
            <td>{{.ID}}</td>
            <td>{{.StartedAt.Format "2006-01-02 15:04"}}</td>
            <td>{{if .FinishedAt}}{{.FinishedAt.Sub .StartedAt}}{{else}}—{{end}}</td>
            <td><span class="sev-{{.Status}}">{{.Status}}</span></td>
            <td><a class="btn" href="/scans/{{.ID}}">Detay</a></td>
          </tr>
        {{end}}
      </tbody>
    </table>
  {{else}}
    <p>Henüz tarama yok.</p>
  {{end}}
</section>
{{end}}
```

- [ ] **Step 2: `internal/server/templates/scan_detail.html` yaz**

```html
{{define "scan_detail_content"}}
<section class="card">
  <h2>Scan #{{.Scan.ID}}</h2>
  <p>{{.Scan.StartedAt.Format "2006-01-02 15:04:05"}} — {{.Scan.Status}}</p>
  <p><strong>{{len .Findings}}</strong> bulgu</p>
</section>

{{if .Findings}}
<section class="card">
  <h3>Bulgular</h3>
  <table>
    <thead>
      <tr><th>Sev</th><th>Uygulama</th><th>Sürüm</th><th>CVE</th><th>Düzeltme</th><th>Özet</th></tr>
    </thead>
    <tbody>
      {{range .Findings}}
        <tr>
          <td class="sev-{{.Severity}}">{{.Severity}}</td>
          <td>{{.AppName}}</td>
          <td>{{.AppVersion}}</td>
          <td>{{.CVEID}}</td>
          <td>{{.FixedVersion}}</td>
          <td>{{.Summary}}</td>
        </tr>
      {{end}}
    </tbody>
  </table>
</section>
{{end}}
{{end}}
```

- [ ] **Step 3: `internal/server/server.go` render fonksiyonunu güncelle (template ismi parametre olsun)**

`render` fonksiyonunu değiştir:

```go
func (s *Server) render(w http.ResponseWriter, layoutName, contentName string, data map[string]any) {
	// İlgili content template'i layout'la birleştir
	data["Title"] = data["Title"]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Clone'la içeriği "content"'a alias'la
	tmpl, err := s.templates.Clone()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := tmpl.New("content").Parse(`{{template "` + contentName + `" .}}`); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, layoutName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

- [ ] **Step 4: `internal/server/handlers.go` çağrılarını güncelle**

```go
// handleDashboard içinde:
	s.render(w, "layout", "content", data)
// handleScansList içinde:
	s.render(w, "layout", "scans_content", map[string]any{...})
// handleScanDetail içinde:
	s.render(w, "layout", "scan_detail_content", map[string]any{...})
```

Tam handlers.go güncel hali:

```go
package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	scans, err := s.Store.ListScans(r.Context(), 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]any{"Title": "Dashboard"}
	if len(scans) > 0 {
		data["LastScan"] = scans[0]
		_, findings, _ := s.Store.GetScan(r.Context(), scans[0].ID)
		data["FindingsCount"] = len(findings)
	}
	s.render(w, "layout", "content", data)
}

func (s *Server) handleScansList(w http.ResponseWriter, r *http.Request) {
	scans, err := s.Store.ListScans(r.Context(), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "layout", "scans_content", map[string]any{
		"Title": "Geçmiş",
		"Scans": scans,
	})
}

func (s *Server) handleScanDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/scans/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	scan, findings, err := s.Store.GetScan(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.render(w, "layout", "scan_detail_content", map[string]any{
		"Title":    "Scan #" + idStr,
		"Scan":     scan,
		"Findings": findings,
	})
}

func (s *Server) handleAPIScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.tryAcquireScan() {
		http.Error(w, "another scan is already running", http.StatusConflict)
		return
	}
	go s.runScanAsync(context.Background())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

- [ ] **Step 5: Build ve manuel test**

```bash
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
./mr-ducksafe.exe serve
```

Tarayıcıda `/`, `/scans`, `/scans/1` sayfalarının doğru render edildiğini doğrula.

- [ ] **Step 6: Commit**

```bash
git add internal/server/
git commit -m "feat(server): scan list and scan detail pages with shared layout"
```

---

### Task 23: Server smoke testleri

**Files:**
- Create: `internal/server/server_test.go`

- [ ] **Step 1: `internal/server/server_test.go` yaz**

```go
package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/osv"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

type stubScanner struct{}

func (stubScanner) Scan(_ context.Context) ([]scanner.InstalledApp, error) { return nil, nil }

type stubMatcher struct{}

func (stubMatcher) Match(_ context.Context, _ scanner.InstalledApp) (matcher.Identifier, matcher.MatchConfidence, error) {
	return matcher.Identifier{}, matcher.ConfidenceNone, nil
}

type stubOSV struct{}

func (stubOSV) Query(_ context.Context, _ matcher.Identifier, _ string) ([]osv.Vulnerability, error) {
	return nil, nil
}

func newTestServer(t *testing.T) (*Server, store.Store) {
	t.Helper()
	st, err := store.OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	eng := scanengine.New(stubScanner{}, stubMatcher{}, stubOSV{}, st)
	srv, err := New(eng, st)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return srv, st
}

func TestDashboard_ReturnsHTML(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %s, want text/html", rec.Header().Get("Content-Type"))
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "mr-ducksafe") {
		t.Errorf("body 'mr-ducksafe' içermiyor")
	}
}

func TestAPIScan_RejectsGET(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/scan", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestAPIScan_SecondCallReturns409(t *testing.T) {
	srv, _ := newTestServer(t)
	if !srv.tryAcquireScan() {
		t.Fatal("first acquire failed")
	}
	defer srv.releaseScan()

	req := httptest.NewRequest(http.MethodPost, "/api/scan", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}

func TestScansList_ReturnsHTML(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/scans", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
```

- [ ] **Step 2: Test çalıştır**

```bash
go test ./internal/server/... -v
```

Expected: 4 test PASS

- [ ] **Step 3: Commit**

```bash
git add internal/server/server_test.go
git commit -m "test(server): smoke tests for dashboard, scan list and concurrency rule"
```

---

## Phase 7 — Polish

### Task 24: End-to-end integration testi

**Files:**
- Create: `internal/scanengine/integration_test.go`
- Create: `testdata/osv_responses/empty.json`

- [ ] **Step 1: `testdata/osv_responses/empty.json` yaz**

```json
{"vulns": []}
```

- [ ] **Step 2: `internal/scanengine/integration_test.go` yaz**

```go
package scanengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/osv"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

type fixedScanner struct{ apps []scanner.InstalledApp }

func (f *fixedScanner) Scan(_ context.Context) ([]scanner.InstalledApp, error) { return f.apps, nil }

func TestIntegration_FirefoxKnownCVECaught(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/osv_responses/firefox_100.json")
	if err != nil {
		t.Fatalf("fixture okuma: %v", err)
	}
	emptyFixture, err := os.ReadFile("../../testdata/osv_responses/empty.json")
	if err != nil {
		t.Fatalf("empty fixture: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// payload'a göre fixture seç
		w.Header().Set("Content-Type", "application/json")
		body := make([]byte, 1024)
		n, _ := r.Body.Read(body)
		req := string(body[:n])
		if strings.Contains(req, "firefox") {
			_, _ = w.Write(fixture)
			return
		}
		_, _ = w.Write(emptyFixture)
	}))
	defer srv.Close()

	st, err := store.OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	m, err := matcher.NewDefaultMatcher()
	if err != nil {
		t.Fatal(err)
	}
	osvClient := osv.NewHTTPClient().WithEndpoint(srv.URL)

	eng := New(
		&fixedScanner{apps: []scanner.InstalledApp{
			{Name: "Mozilla Firefox 100.0 (x64)", Version: "100.0", Publisher: "Mozilla"},
			{Name: "Brand New Tool", Version: "1.0", Publisher: "Acme"},
		}},
		m,
		osvClient,
		st,
	)

	res, err := eng.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(res.Findings) != 1 {
		t.Fatalf("1 finding bekleniyor, alındı: %d", len(res.Findings))
	}
	if res.Findings[0].CVEID != "CVE-2022-26485" {
		t.Errorf("CVE = %s, want CVE-2022-26485", res.Findings[0].CVEID)
	}
	if res.Findings[0].FixedVersion != "97.0.2" {
		t.Errorf("FixedVersion = %s, want 97.0.2", res.Findings[0].FixedVersion)
	}
	if len(res.Unmatched) != 1 || res.Unmatched[0].Name != "Brand New Tool" {
		t.Errorf("unmatched beklenmedik: %+v", res.Unmatched)
	}
}
```

- [ ] **Step 3: Test çalıştır**

```bash
go test ./internal/scanengine/... -v
```

Expected: Tüm testler dahil yeni integration PASS

- [ ] **Step 4: Commit**

```bash
git add internal/scanengine/integration_test.go testdata/osv_responses/empty.json
git commit -m "test(scanengine): integration test for Firefox known CVE end-to-end flow"
```

---

### Task 25: README'yi gerçek kullanım talimatlarıyla güncelle

**Files:**
- Modify: `README.md`

- [ ] **Step 1: README'yi yeniden yaz**

```markdown
# mr-ducksafe

Yerel makinede kurulu sistem yazılımları için bilinen güvenlik açıklarını (CVE) raporlayan tek dosya CLI ve lokal web aracı. Veriler [OSV.dev](https://osv.dev)'den çekilir. Hiçbir telemetri toplanmaz; tarama sonuçları sadece lokal SQLite'a yazılır.

**Platform desteği:** MVP — Windows. Linux ve macOS v0.2'de planlanıyor.

## Hızlı Başlangıç

### Binary indir

[GitHub Releases](https://github.com/yemresalcan/mr-ducksafe/releases) sayfasından `mr-ducksafe-<os>-<arch>.zip` indir, çıkart, `PATH`'e ekle.

### Kaynaktan build et

Go 1.22+ gerekli.

```sh
git clone https://github.com/yemresalcan/mr-ducksafe
cd mr-ducksafe
go build -o mr-ducksafe ./cmd/mr-ducksafe
```

## Kullanım

### CLI tarama

```sh
mr-ducksafe scan                              # tablo, renkli
mr-ducksafe scan --format=json                # JSON (CI/script)
mr-ducksafe scan --format=markdown            # markdown (PR/issue)
mr-ducksafe scan --show-unmatched             # eşleştirilemeyen yazılımları da göster
mr-ducksafe scan --fail-on=high               # high+ bulgu varsa exit 1
```

### Lokal web UI

```sh
mr-ducksafe serve --port 8080
# http://127.0.0.1:8080 aç
```

Dashboard'dan "Şimdi Tara" butonu, geçmiş, scan detay görüntülenebilir. UI default olarak sadece `127.0.0.1`'e bağlanır. Ağa açmak için `--bind 0.0.0.0` (auth yok, dikkatli kullan).

### Eski scan kayıtlarını sil

```sh
mr-ducksafe prune --older-than=30d
```

## Exit Code'ları

| Code | Anlam |
|---|---|
| 0 | Başarılı, eşik aşılmadı |
| 1 | Eşik aşıldı (`--fail-on`) |
| 2 | Tarama hatası (OSV erişilemedi, scanner başarısız) |
| 64 | Kullanım hatası (yanlış flag) |

## Gizlilik

- Tarama tamamen lokaldir.
- OSV.dev'e gönderilen tek veri: `{ecosystem, name, version}` — makine adı, kullanıcı bilgisi, IP içermez (TLS üzerinden HTTPS).
- Sonuçlar `~/.mr-ducksafe/mr-ducksafe.db` içinde saklanır.

## Geliştirme

```sh
go test ./...                  # tüm testler
go test ./... -race            # race detector ile
golangci-lint run              # lint
```

Tasarım: [docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md](docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md)
Implementasyon planı: [docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md](docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md)

## Lisans

Henüz belirlenmedi (MIT veya Apache-2.0). MVP build'inden önce eklenecek.

## Topluluk

Yeni yazılım eşleştirmesi eklemek için `internal/matcher/data/mappings.yaml` dosyasına PR aç.
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: update README with installation, usage, exit codes and privacy notes"
```

---

### Task 26: Release build script ve doğrulama

**Files:**
- Create: `scripts/build.sh`
- Create: `scripts/build.ps1`

- [ ] **Step 1: `scripts/build.sh` yaz (Linux/macOS host için cross-build)**

```sh
#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-dev}"
DIST_DIR="dist"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

build() {
  local goos="$1" goarch="$2" ext="$3"
  local out="$DIST_DIR/mr-ducksafe-${goos}-${goarch}${ext}"
  echo "Building $out"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build \
    -ldflags "-s -w -X github.com/yemresalcan/mr-ducksafe/internal/cmd.Version=$VERSION" \
    -o "$out" ./cmd/mr-ducksafe
}

build windows amd64 .exe
build linux   amd64 ""
build darwin  amd64 ""
build darwin  arm64 ""

echo "Done. Artifacts in $DIST_DIR/"
ls -lh "$DIST_DIR"
```

- [ ] **Step 2: `scripts/build.ps1` yaz (Windows host için)**

```powershell
param(
  [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"
$distDir = "dist"

if (Test-Path $distDir) { Remove-Item -Recurse -Force $distDir }
New-Item -ItemType Directory -Path $distDir | Out-Null

$targets = @(
  @{ OS = "windows"; Arch = "amd64"; Ext = ".exe" }
  @{ OS = "linux";   Arch = "amd64"; Ext = ""     }
  @{ OS = "darwin";  Arch = "amd64"; Ext = ""     }
  @{ OS = "darwin";  Arch = "arm64"; Ext = ""     }
)

foreach ($t in $targets) {
  $out = Join-Path $distDir "mr-ducksafe-$($t.OS)-$($t.Arch)$($t.Ext)"
  Write-Host "Building $out"
  $env:GOOS   = $t.OS
  $env:GOARCH = $t.Arch
  $env:CGO_ENABLED = "0"
  go build -ldflags "-s -w -X github.com/yemresalcan/mr-ducksafe/internal/cmd.Version=$Version" -o $out ./cmd/mr-ducksafe
  if ($LASTEXITCODE -ne 0) { throw "Build failed for $($t.OS)/$($t.Arch)" }
}

Write-Host "Done. Artifacts in $distDir/"
Get-ChildItem $distDir
```

- [ ] **Step 3: Build script'ini lokalde çalıştır (Windows)**

```bash
./scripts/build.ps1 -Version "0.1.0-dev"
./dist/mr-ducksafe-windows-amd64.exe version
```

Expected: `mr-ducksafe 0.1.0-dev`, ayrıca dist/ içinde 4 binary olur, her biri < 20 MB.

- [ ] **Step 4: Boyutu kontrol et**

```bash
ls -lh dist/
```

Eğer Windows binary > 20 MB ise:
- `ldflags -s -w` zaten var (debug symbol kaldırır)
- UPX ile sıkıştırma opsiyonel ama virus scanner false positive verir, MVP'de gerek yok

- [ ] **Step 5: Commit**

```bash
git add scripts/
git commit -m "build: cross-platform build scripts for windows/linux/macos targets"
```

---

## Plan Sonu

Tüm task'lar tamamlandığında:

- ✅ Tek binary, cross-platform build
- ✅ `mr-ducksafe scan` Windows Registry'den 50+ yaygın yazılımı eşleştirir, CVE bulur
- ✅ `mr-ducksafe serve` lokal dashboard sunar, on-demand scan + history gösterir
- ✅ Test-after pragmatik test coverage: matcher, osv, store, scanengine, server smoke
- ✅ Integration test: bilinen Firefox CVE'sinin yakalandığını doğrular
- ✅ CI: Windows + Linux + macOS matrix, golangci-lint + gosec
- ✅ Spec'teki başarı kriterleri karşılanır

**Spec'ten ertelenen (v0.2 için):**
- **SSE-tabanlı canlı progress UI'ı** — spec bölüm 6.2'de tanımlıydı, MVP'de pratik basitlik için ertelendi. Şu an UI: "Scan Now" → redirect → manuel refresh. Mevcut `Engine.Run(ctx, progress chan)` imzası destekliyor; sadece `/api/scans/{id}/events` endpoint + HTMX SSE template güncellemesi gerekir.

**Sonraki sürüm (v0.2) için aday özellikler:**
- Linux scanner (dpkg, rpm, pacman)
- macOS scanner (Homebrew, `/Applications`)
- SSE canlı progress (yukarıda)
- Install script'leri ve resmi domain
- Mapping listesinin 50'den 200+'e çıkarılması (topluluk PR'larıyla)
- Offline mod (lokal CVE DB cache)
