# mr-ducksafe — Design Spec

**Tarih:** 2026-05-23
**Durum:** Taslak (kullanıcı onayı bekliyor)
**Yazar:** Yunus Emre Salcan
**Versiyon:** 0.1.0 (MVP)

---

## 1. Özet

`mr-ducksafe`, kullanıcının makinesinde kurulu olan sistem yazılımlarını tarayıp her biri için bilinen güvenlik açıklarını (CVE) raporlayan, Go ile yazılmış tek dosyalık bir araçtır. İki modda çalışır:

- **CLI modu** — `mr-ducksafe scan` terminalde tablo veya JSON çıktısı üretir
- **Server modu** — `mr-ducksafe serve` lokal bir HTTP server başlatır, tarayıcıdan on-demand tarama yapılır ve geçmiş görüntülenir

Çıkarımları sadece kullanıcıya gösterir — otomatik güncelleme uygulamaz, hiçbir telemetri toplamaz, sadece OSV.dev'e sorgu gönderir.

## 2. Hedefler

- Windows üzerinde kurulu sistem yazılımlarını listelemek ve bilinen CVE'lerle eşleştirmek
- Yanlış pozitifi düşük tutmak — eşleştirilemeyen yazılımları gizlemek yerine şeffafça göstermek
- Tek binary olarak dağıtılabilir, harici bağımlılık (Docker, runtime, lib) gerektirmemek
- CI/script entegrasyonu için anlamlı exit code'lar
- Linux desteği için temiz bir genişleme yolu (interface tabanlı scanner)
- Çekirdek mantığın testlenebilir olması (interface arkasında dependency'ler)

## 3. Hedef Olmayanlar (YAGNI)

- Otomatik güncelleme/fix uygulama (kullanıcı kararını ve sorumluluğunu üstlenmek istemiyoruz)
- Authentication / multi-user web UI (lokal araç, multi-tenant değil)
- Cloud backend, telemetri, kullanıcı hesabı
- Browser extension, VS Code extension, npm/pip paket taraması (gelecek sürümlerde değerlendirilebilir)
- Container/SBOM taraması (bu alan Grype/Trivy tarafından zaten çözülmüş)
- Real-time monitoring / arka plan daemon

## 4. Mimari

### 4.1 Üst seviye yapı

Tek Go binary, üç komut: `scan`, `serve`, `version`. `cmd/` katmanı sadece CLI parsing ve çıktı sunumu yapar; tüm iş mantığı `internal/` altındadır.

```
cmd/mr-ducksafe/          → Cobra root, subcommands, config
internal/
  scanner/             → OS-spesifik kurulu yazılım tespiti (interface)
    windows/           → Registry tabanlı (build tag: windows)
    linux/             → (sonraki sürüm) dpkg/rpm parser
  matcher/             → InstalledApp → CPE/purl identifier
  osv/                 → OSV.dev API client (cache + retry)
  store/               → SQLite persist (pure-Go driver)
  scanengine/          → Orkestrasyon: scan → match → query → persist
  report/              → CLI table/json/markdown formatter
  server/              → HTTP + embedded UI (html/template + HTMX)
  config/              → ~/.mr-ducksafe/config.yaml okuyucu
data/
  mappings.yaml        → Curated app → CPE eşleştirme (gömülü)
docs/
  superpowers/specs/   → Bu doküman ve takip eden plan
```

### 4.2 Katman bağımlılık kuralı

`cmd/` ve `internal/server` aynı seviyededir; ikisi de `scanengine` üzerinden iş yaptırır. `scanengine` somut tipleri değil interface'leri kabul eder. Yan paketler birbirini import etmez — sadece `scanengine` ve `cmd/server` birden fazla paketi tüketir.

## 5. Paket Sorumlulukları

### 5.1 `internal/scanner`

OS'tan kurulu yazılım listesini çıkarır.

```go
type Scanner interface {
    Scan(ctx context.Context) ([]InstalledApp, error)
}

type InstalledApp struct {
    Name      string  // "Google Chrome"
    Version   string  // "120.0.6099.130"
    Publisher string  // "Google LLC"
    Source    string  // "windows-registry" | "dpkg" | "rpm"
}
```

`New() Scanner` runtime OS'una göre doğru implementasyonu döner. Windows için iki Registry yolundan okur:
- `HKLM\Software\Microsoft\Windows\CurrentVersion\Uninstall`
- `HKLM\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`

Bağımlılık: stdlib + `golang.org/x/sys/windows/registry`.

### 5.2 `internal/matcher`

`InstalledApp` → sorgulanabilir identifier eşleştirir.

```go
type Matcher interface {
    Match(app InstalledApp) (Identifier, MatchConfidence, error)
}

type Identifier struct {
    Type      IdentifierType  // PURL | CPE
    Value     string
    Ecosystem string           // OSV ecosystem hint
}

type MatchConfidence int  // High | Medium | Low | None
```

Üç katmanlı strateji (detay: bölüm 7).

Bağımlılık: stdlib + embed (mapping YAML gömülü).

### 5.3 `internal/osv`

OSV.dev API client.

```go
type Client interface {
    Query(ctx context.Context, id Identifier, version string) ([]Vulnerability, error)
}

type Vulnerability struct {
    ID            string    // "CVE-2024-1234" veya "GHSA-..."
    Summary       string
    Severity      Severity  // Critical | High | Medium | Low | Unknown
    FixedVersions []string
    References    []string
    Published     time.Time
}
```

Endpoint: `POST https://api.osv.dev/v1/query`. Concurrent semaphore (default 5), exponential backoff (max 3 retry), in-memory response cache (24h TTL).

Bağımlılık: `net/http` + stdlib.

### 5.4 `internal/store`

SQLite üzerinden persist. Driver: `modernc.org/sqlite` (pure-Go, CGO yok — cross-compile için kritik).

```go
type Store interface {
    SaveScan(ctx context.Context, scan Scan) (int64, error)
    AppendFindings(ctx context.Context, scanID int64, f []Finding) error
    UpdateScanStatus(ctx context.Context, scanID int64, status ScanStatus) error
    ListScans(ctx context.Context, limit int) ([]Scan, error)
    GetScan(ctx context.Context, id int64) (Scan, []Finding, error)
    CleanupStaleRunning(ctx context.Context, olderThan time.Duration) (int, error)
}
```

Şema:
```sql
CREATE TABLE scans (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at   DATETIME NOT NULL,
    finished_at  DATETIME,
    host         TEXT,
    os           TEXT,
    status       TEXT NOT NULL  -- 'running' | 'success' | 'failed' | 'interrupted'
);

CREATE TABLE findings (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_id       INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    app_name      TEXT NOT NULL,
    app_version   TEXT NOT NULL,
    cve_id        TEXT NOT NULL,
    severity      TEXT NOT NULL,
    fixed_version TEXT,
    summary       TEXT,
    references    TEXT  -- JSON array
);

CREATE INDEX idx_findings_scan_id ON findings(scan_id);
CREATE INDEX idx_findings_severity ON findings(severity);
```

Default konum: `~/.mr-ducksafe/mr-ducksafe.db`.

### 5.5 `internal/scanengine`

Orkestrasyon. Tek public fonksiyon:

```go
type Engine struct {
    scanner Scanner
    matcher Matcher
    osv     Client
    store   Store
    sem     int  // concurrent OSV query limit
}

type ScanResult struct {
    ScanID    int64
    Apps      []InstalledApp
    Findings  []Finding
    Unmatched []InstalledApp
    Warnings  []ScanError
    Started   time.Time
    Finished  time.Time
}

func (e *Engine) Run(ctx context.Context, progress chan<- Progress) (*ScanResult, error)
```

Akış:
1. `store.SaveScan(status=running)` → scanID
2. `scanner.Scan()` → apps
3. Her app için `matcher.Match()` (sync, hızlı)
4. Match olanlar için concurrent (semaphore) `osv.Query()`
5. Findings + warnings'i `store.AppendFindings()`
6. `store.UpdateScanStatus(success|failed)`
7. ScanResult döndür

`progress` kanalı opsiyonel — server SSE için kullanır, CLI dinlemez.

### 5.6 `internal/report`

CLI çıktı formatter.

```go
type Format string  // "table" | "json" | "markdown"

func Format(result *ScanResult, format Format, opts Options) (string, error)
```

Tablo renkli severity sütunu, kritik/yüksek üstte. JSON stabil schema (versiyonlu). Markdown PR/issue'ya yapıştırılabilir formatta.

Bağımlılık: `github.com/olekukonko/tablewriter` + `github.com/fatih/color`.

### 5.7 `internal/server`

HTTP server + embedded UI. Default bind `127.0.0.1:8080`.

Endpoints:
| Method | Path | Açıklama |
|---|---|---|
| GET | `/` | Dashboard (son scan özeti) |
| GET | `/scans` | Geçmiş listesi |
| GET | `/scans/{id}` | Scan detayı (findings + unmatched) |
| POST | `/api/scan` | Yeni scan tetikle (async) |
| GET | `/api/scans/{id}/events` | SSE progress stream |
| GET | `/api/scans/{id}/findings.json` | JSON export |
| GET | `/static/*` | Embedded CSS/HTMX |

Eşzamanlılık: tek aktif scan kuralı. İkinci scan denemesi → 409 Conflict.

UI teknolojisi: `html/template` + HTMX (CDN değil, gömülü). JS build pipeline yok. CSP strict, harici kaynak yüklenmez.

Bağımlılık: stdlib (`net/http`, `html/template`, `embed`) + minimal.

### 5.8 `cmd/mr-ducksafe`

Cobra root + subcommands. Config dosyası `~/.mr-ducksafe/config.yaml` (opsiyonel) okur. Flag'ler:

```
mr-ducksafe scan [--json|--markdown] [--fail-on=critical|high|medium] [--show-unmatched] [--verbose]
mr-ducksafe serve [--port=8080] [--bind=127.0.0.1] [--db=path]
mr-ducksafe version [--check-updates]
mr-ducksafe prune --older-than=30d
```

Exit code'lar:
| Code | Anlam |
|---|---|
| 0 | Başarılı, `--fail-on` eşiğini aşan bulgu yok |
| 1 | Başarılı, eşik aşıldı (CI fail için) |
| 2 | Tarama hatası (OSV unreachable, scanner çöktü) |
| 64 | Kullanım hatası (yanlış flag) |

## 6. Veri Akışı

### 6.1 CLI scan

```
cmd/mr-ducksafe scan
  → scanengine.Run()
      → scanner.Scan() ........... []InstalledApp
      → matcher.Match() (loop) ... (id, confidence)
      → osv.Query() (concurrent) . []Vulnerability
      → store.SaveScan/Append
  → report.Format(result, table)
  → stdout
  → exit code (severity threshold check)
```

### 6.2 Server scan (UI butonu)

```
Browser → POST /api/scan
  → server: scanID := store.SaveScan(running)
  → server: go scanengine.Run(ctx, progressCh)
  → response: {scanID, streamURL}

Browser → GET /api/scans/{id}/events (SSE)
  → server consumes progressCh
  → stream: "scanned: 45/187", "querying OSV: 12/45", "done: 18 vulns"
  → Browser HTMX SSE extension tabloyu canlı günceller
```

### 6.3 Veri yaşam döngüsü

- Scan kayıtları SQLite'ta sınırsız. `prune` komutu eski kayıtları siler.
- OSV cache in-memory, process restart'ta silinir, TTL 24h.
- Hiçbir veri dışarı gitmez (OSV sorguları hariç). OSV payload'u sadece `{ecosystem, name, version}` içerir — kullanıcı kimliği, makine adı yok.

## 7. CVE Eşleştirme Stratejisi

Üç katmanlı:

```
1. Curated mapping (data/mappings.yaml — gömülü)
   "Google Chrome" → pkg:generic/google/chrome  [High]

2. Normalize edilmiş publisher + name
   "Mozilla Corp." + "Firefox 121.0" → vendor=mozilla, product=firefox  [Medium]

3. Eşleşemedi → unmatched listesine ekle  [None]
```

### 7.1 `mappings.yaml` format

```yaml
- registry_name: "Google Chrome"
  publisher: "Google LLC"
  identifier:
    type: purl
    value: "pkg:generic/google/chrome"
  ecosystem_hint: "Chrome"
  version_normalize: "semver"

- registry_name_pattern: "^Mozilla Firefox.*"
  publisher: "Mozilla"
  identifier:
    type: cpe
    value: "cpe:2.3:a:mozilla:firefox"
  version_normalize: "semver"
```

### 7.2 MVP kapsam

Top 50 yaygın Windows yazılımı için curated mapping: Chrome, Firefox, Edge, VLC, 7-Zip, Notepad++, Office, Acrobat Reader, Java JRE/JDK, Python, Node.js, Git, OBS, Discord, Slack, Zoom, WinRAR, PuTTY, FileZilla, vb. Bunlar gerçek dünyada kurulu yazılım vuln'larının büyük çoğunluğunu kapsar.

### 7.3 Version normalization

Windows version'ları kaotik (`120.0.6099.130`, `2024.01.5`, `21H2`, `v1.2-beta`). `internal/matcher/version.go` strateji uygular: `semver`, `calver`, `fourpart`, `vendor-specific`. Bilinmeyen format → `Low` confidence + warning.

### 7.4 Unmatched şeffaflığı

- UI'da "23 program eşleştirilemedi [Göster]" expandable bandı
- CLI'da default gizli, `--show-unmatched` ile listelenir
- Sonuç JSON'ında her zaman `unmatched: [...]` alanı var

### 7.5 Topluluk katkı yolu

`mappings.yaml` ayrı dosya — PR ile yeni mapping eklenir. Bu, projenin tek geliştiriciyle sürdürülebilir kalmasını sağlar.

## 8. Hata Yönetimi ve Güvenlik

### 8.1 Temel kural

Sessiz başarısızlık yasak. Kısmi başarılar şeffaf raporlanır.

| Senaryo | Davranış |
|---|---|
| Registry kısmen okunamadı (izin) | Devam et, `warnings: ["N anahtar okunamadı"]` |
| OSV.dev erişilemez (offline, 5xx) | Scan fail, exit 2, "OSV unreachable" |
| OSV bir app için 429 | Backoff + retry max 3, sonra app error state |
| Version parse edilemedi | Unmatched listesine + sebep loglanır |
| SQLite write hatası | Scan abort, yarım kayıt yazılmaz |
| Goroutine panic | `recover` → app skip, panic logged |

### 8.2 Hata tipi

```go
type ScanError struct {
    Stage    string  // "scan" | "match" | "query" | "persist"
    AppName  string  // boş olabilir
    Cause    error
    Fatal    bool    // true → scan abort
}
```

`ScanResult.Warnings []ScanError` ile döner.

### 8.3 Loglama

`log/slog` (Go 1.21+ stdlib). Default `INFO`, `--debug` ile `DEBUG`. Hassas veri (path'ler, makine adı) opt-in.

### 8.4 Güvenlik

- **Tarama read-only.** Sistem dosyası yazılmaz, Registry yazılmaz.
- **Web UI sadece localhost** default. `--bind 0.0.0.0` açık flag ile + uyarı: "Bu makineyi ağa açar, auth yok."
- **CSRF:** Form POST'larında SameSite cookie + token. SSE GET'tir, kapsam dışı.
- **OSV sorguları HTTPS only,** TLS doğrulama zorunlu, downgrade yok.
- **Embedded UI dış kaynak yüklemez** (CDN yok, CSP strict).

### 8.5 Crash recovery

`SaveScan(running)` → process çökerse SQLite'ta `running` kalır. `serve` başlangıcında `CleanupStaleRunning(10*time.Minute)` çağrılır → eski `running` scan'ler `interrupted` olarak işaretlenir.

## 9. Test Stratejisi

Yaklaşım: **test-after pragmatik + golden file ağırlıklı**. TDD disiplini yok — kod yazılır, çalıştığı görülür, sonra anlamlı davranışlar için test yazılır.

### 9.1 Hangi paketler test edilir

| Paket | Test odağı | Tip |
|---|---|---|
| `matcher` | 50 mapping, regex, version normalize, confidence seviyeleri | Unit |
| `osv` | API response parse (fixture JSON), retry, error → ScanError | Unit + golden |
| `scanengine` | Mock bağımlılıklarla orkestrasyon, partial failure | Unit |
| `store` | In-memory SQLite, scan + findings round-trip, cleanup | Unit |
| `report` | Table/JSON/markdown çıktıları | Golden file |
| `scanner/windows` | Registry'den minimum 1 entry okuma (sadece Windows CI) | Smoke |
| `server` | Endpoint smoke (200 + JSON schema), tek aktif scan kuralı | Smoke |

### 9.2 Test edilmeyen

- Trivial getter/setter
- Cobra wiring (manuel doğrulanır)
- Her endpoint'in detayı (smoke test yeterli)

### 9.3 Integration testi

`internal/scanengine/integration_test.go` — gerçek `matcher` + gerçek `store` + mock OSV (httptest server). Kritik senaryo: "Firefox 100 → bilinen CVE-X yakalandı mı?"

### 9.4 CI matrix

```
GitHub Actions:
  windows-latest: go test ./... (full)
  ubuntu-latest:  go test ./... (scanner/windows skip)
  macos-latest:   pure-Go paketler (matcher, osv, store, report)

golangci-lint: PR zorunlu
gosec:         güvenlik linter, PR zorunlu
```

### 9.5 Test verisi

- `testdata/registry_dump.json` — fake Windows Registry örneği
- `testdata/osv_responses/*.json` — gerçek OSV response'lardan sansürlü kopyalar
- `testdata/mappings_test.yaml` — minimal mapping seti

## 10. Teknoloji Stack

| Kategori | Seçim | Gerekçe |
|---|---|---|
| Dil | Go 1.22+ | Cross-platform tek binary, security tool ekosistemi |
| CLI framework | `spf13/cobra` | Endüstri standardı, subcommand desteği |
| Config | `spf13/viper` | YAML + env var + flag birleştirme |
| SQLite | `modernc.org/sqlite` | Pure-Go, CGO yok, cross-compile basit |
| HTTP server | stdlib `net/http` | Yeterli, ekstra dependency'ye gerek yok |
| HTML | stdlib `html/template` + HTMX | JS build pipeline yok |
| Table output | `olekukonko/tablewriter` | Olgun, renkli severity destekler |
| Logging | stdlib `log/slog` | Go 1.21+ resmi yapılandırılmış log |
| Windows Registry | `golang.org/x/sys/windows/registry` | Resmi Go x repo |
| Lint | `golangci-lint` + `gosec` | Standart Go toolchain |

## 11. Dağıtım

- **MVP:** GitHub Releases üzerinden Windows/Linux/macOS için statik build (CGO yok, kolay).
- **Kurulum:** İlk sürüm sadece "release sayfasından binary indir, PATH'e koy". Install script'leri (PowerShell `iwr | iex`, Bash `curl | sh`) ve özel domain v0.2+ için ertelendi (hosting + domain kararı gerektirir).
- **Lisans:** Karar ertelendi — MVP build'inden önce MIT veya Apache-2.0 arasında seçim yapılacak (bkz. açık sorular).
- **Server modu dağıtımı:** Repo clone + `go build`, veya release binary indir, `mr-ducksafe serve` çalıştır.

## 12. Açık Sorular / Gelecek

- Linux scanner implementasyonu (v0.2): dpkg, rpm, pacman parser'ları
- macOS desteği (v0.3): Homebrew + `/Applications` taraması
- Lisans seçimi (MIT vs Apache-2.0)
- Offline modu (lokal CVE DB) — kullanıcı talebine göre v0.4+
- Mapping topluluk PR şablonu ve onay süreci
- Web UI auth (multi-user server kullanımı için) — talep gelirse

## 13. Başarı Kriterleri (MVP)

- Windows 10/11'de `mr-ducksafe scan` 187 kurulu yazılımdan en az %70'ini eşleştirir
- En az 50 yaygın yazılım için curated mapping mevcut
- Bilinen vuln testi: Firefox 100 kurulu makinede en az 5 CVE bulur
- `mr-ducksafe serve` lokal browser'da çalışır, "Scan Now" butonu işler, geçmiş listelenir
- Tek binary < 20 MB, ilk scan < 30 saniye (100 app, internet ortalama hız)
- CI matrix yeşil (Windows + Linux + macOS)
