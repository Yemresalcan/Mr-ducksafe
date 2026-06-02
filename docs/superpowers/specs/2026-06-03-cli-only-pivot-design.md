# CLI-Only Pivot + CV-Ready Polish — Design

**Tarih:** 2026-06-03
**Konu:** Web UI'ı kaldırıp projeyi salt terminal/CLI'a indirgemek, ardından CV'de profesyonel görünmesini sağlayacak release/QoL eksiklerini tamamlamak.

## Amaç

`mr-ducksafe` CV'ye eklenecek. Hedef sunum: **"Go ile yazılmış, Windows kurulu yazılımları için NVD CVE tarayıcı — terminal aracı, otomatik release pipeline'ı, SQLite tabanlı yerel geçmiş."** Web UI eklenmişti ama olmadı; tek bir komutla çalışan saf bir CLI olarak konumlanması daha güçlü bir hikâye.

## Kapsam

İki blok iş var:

1. **Web UI'ın temizliği** — `serve` komutu ve `internal/server/` paketi tamamen kalkar.
2. **CV için profesyonelleştirme** — lisans, release otomasyonu, eksik CLI komutları, CI hijyeni.

## Mevcut Durum (özet)

```
cmd/mr-ducksafe        → main entry
internal/cmd           → cobra komutları: scan, serve, prune, version
internal/scanner       → Windows Registry tarayıcı (build tag'lı)
internal/matcher       → 50 mapping (YAML embedded), CPE eşleştirme
internal/vuln + nvd    → NVD 2.0 client + cache
internal/scanengine    → orkestrasyon (goroutine + semaphore)
internal/report        → table | json | markdown + özet
internal/advice        → severity + keyword → insan dili
internal/store         → SQLite (scans + findings)
internal/tui           → banner + spinner
internal/server        → HTTP + embedded HTML/CSS  (KALDIRILACAK)
.github/workflows/ci.yml → test (3 OS) + lint
scripts/build.{ps1,sh} → manuel cross-platform build
```

## Bölüm 1 — Web UI Temizliği

### Silinecek dosyalar

```
internal/server/server.go
internal/server/handlers.go
internal/server/embed.go
internal/server/server_test.go
internal/server/templates/layout.html
internal/server/templates/dashboard.html
internal/server/templates/scans.html
internal/server/templates/scan_detail.html
internal/server/static/style.css
internal/cmd/serve.go
```

### Düzenlenecek dosyalar

- `internal/cmd/root.go` — `root.AddCommand(NewServeCmd())` satırı kalkar.
- `internal/store/store.go` — `CleanupStaleRunning` interface metodu kalkar (sadece serve kullanıyor).
- `internal/store/sqlite.go` — aynı metodun implementasyonu kalkar.
- `internal/store/store_test.go` ve `sqlite_test.go` — `CleanupStaleRunning` referansları temizlenir.
- `go.sum` — kullanılmayan transitive bağımlılıklar `go mod tidy` ile düşer.

### README'den çıkarılacaklar

- "🖥️ **Opsiyonel web UI**" özellik maddesi
- `### serve — lokal web UI` bölümünün tamamı
- `Available Commands` listesinden `serve` satırı
- Mimari diyagramından `internal/server` satırı + altındaki "lokal HTTP + embedded UI" açıklaması
- Roadmap'teki "SSE canlı progress (web UI'da spinner)" satırı

## Bölüm 2 — CV için Profesyonelleştirme

### 2.1 LICENSE (MIT)

- Kök dizine `LICENSE` (standart MIT, copyright `Yunus Emre Salcan, 2026`).
- README'deki "Lisans: Henüz belirlenmedi" bölümü "MIT — bkz. [LICENSE](LICENSE)" olur.

### 2.2 Version ldflags inject

- `internal/cmd/version.go` içinde `Version = "dev"` zaten var.
- Build scriptlerine ve GoReleaser config'ine `-ldflags "-X github.com/yemresalcan/mr-ducksafe/internal/cmd.Version={{.Version}}"` eklenir.
- `mr-ducksafe version` release binary'sinde `v0.1.0` döner.

### 2.3 GoReleaser + Release workflow

**`.goreleaser.yaml`:**
- builds: windows/amd64, linux/amd64, darwin/amd64, darwin/arm64
- archives: zip (windows), tar.gz (diğerleri)
- checksum: sha256
- changelog: conventional commits (feat:, fix:, docs:…) gruplanmış
- sboms: syft ile artifact başına SBOM (cyclonedx-json)
- scoops: opsiyonel, Scoop bucket repo'su gerektirir → bu turda **kapalı bırak**, sadece manifest README'de göster

**`.github/workflows/release.yml`:**
- trigger: `push` tag `v*`
- job: `goreleaser/goreleaser-action@v6` → `release --clean`
- `GITHUB_TOKEN` yeterli (Homebrew/Scoop tap kullanılmayacak)

### 2.4 Coverage CI + Codecov badge

- `ci.yml` test step zaten `-coverprofile=coverage.out` üretiyor.
- `codecov/codecov-action@v5` ile upload eklenir (sadece ubuntu-latest matrisinde).
- README'ye Codecov badge eklenir.

### 2.5 `history` komutu

`mr-ducksafe history [--limit N]`:

```
ID   Tarih                Durum    Süre   Bulgu
12   2026-06-03 14:21:08  success  18s    521
11   2026-06-02 09:14:22  success  21s    519
10   2026-06-01 18:02:11  failed   4s     -
```

- `store.ListScans` zaten var.
- Bulgu sayısı için `GetScan` yapmak yerine küçük bir `CountFindings(scanID)` metodu eklenir (tek COUNT query).
- Varsayılan limit: 20.

### 2.6 `scan --output FILE`

- `report.Write` zaten `io.Writer` alıyor.
- Flag eklenir; boşsa stdout, doluysa `os.Create(path)`.
- Hata: zaten varsa overwrite (CI dostu).

### 2.7 NO_COLOR / `--no-color` desteği

- `internal/tui` ve `report/table` `fatih/color` kullanıyor.
- `fatih/color` paketi zaten `NO_COLOR` env'i otomatik onurlandırır — ek iş gerekmez, ama README'de **belirt**.
- `--no-color` global flag eklenir: set edilirse `color.NoColor = true`.

### 2.8 JSON çıktısına `schema_version`

- `report/json.go` top-level struct'a `SchemaVersion string \`json:"schema_version"\`` eklenir → `"1.0"`.
- `result.golden.json` güncellenir.
- Tüketicilere "kırılan değişiklik olursa bumplanır" sözü.

### 2.9 Issue/PR template'leri

- `.github/ISSUE_TEMPLATE/bug_report.yml`
- `.github/ISSUE_TEMPLATE/feature_request.yml`
- `.github/ISSUE_TEMPLATE/new_mapping.yml` — `mappings.yaml`'a yeni uygulama eklemek için form
- `.github/PULL_REQUEST_TEMPLATE.md` — kısa checklist (test, lint, mapping verified)

### 2.10 Scoop manifest

- Release sonrası kullanıcılar `scoop install <url>/scoop/mr-ducksafe.json` ile kursun.
- `dist/scoop/mr-ducksafe.json` template README'de gösterilir; GoReleaser üzerinden otomatik bucket publish bu turda **kapalı**.
- Bu adımın amacı CV'de "package manager distribution thinking" sinyali; tam otomasyon v0.2'ye bırakılır.

### 2.11 SBOM

- GoReleaser `sboms:` bloğu yapar (syft). Ek kod yok.

## Test Stratejisi

- `go test ./...` ve `golangci-lint run` her şeyi geçmeli.
- Web UI silindikten sonra `internal/server` referansı kalmamalı — `go vet ./...` doğrular.
- `report/testdata/result.golden.json` schema_version eklendiği için **yeniden onaylanır** (test yeniden yazılır, sonuç sabit kalır).
- `history` ve `--output` için unit test (history ListScans wrapper'ı, --output dosya yazımı).
- Release workflow tag'lı testte (`v0.0.0-test` gibi pre-release tag, ya da `goreleaser release --snapshot --clean` ile lokal smoke).

## Olmayacaklar (YAGNI)

- Linux/macOS scanner — roadmap'te kalıyor, bu turda yok.
- LLM destekli açıklama (Seviye B) — roadmap'te.
- Auth eklenmiş yeni bir web UI — proje pivot ediyor, geri dönmüyor.
- Docker image — Windows CLI için minimal değer; eklenirse sonra.
- Homebrew tap, Scoop bucket otomasyonu — manuel manifest yeterli.

## Çalışma Sırası

1. Spec commit (bu dosya).
2. Web UI sil + `go mod tidy` + testler yeşil.
3. LICENSE + version ldflags + README lisans/komut satırları.
4. `history` + `scan --output` + `--no-color` + `schema_version`.
5. GoReleaser config + release workflow + Codecov badge.
6. Issue/PR template'leri + Scoop manifest snippet.
7. README'yi son düzene sok (badges, kurulum bölümü, history komutu, no-color).
8. Final `go test ./...` + `golangci-lint run` + commit zinciri.

## Commit Stratejisi

Tek büyük commit yerine konuya göre küçük commitler (conventional commits):

```
chore: remove web UI (serve command + internal/server package)
chore: add MIT LICENSE
build: inject version via ldflags in goreleaser and build scripts
feat(cli): history command listing past scans from local DB
feat(cli): scan --output FILE flag
feat(cli): --no-color flag honoring NO_COLOR env
feat(report): add schema_version to JSON output
ci: upload coverage to Codecov
build(release): goreleaser config and GitHub Actions release workflow
docs: add issue/PR templates and Scoop install snippet
docs: README polish — remove serve, add new commands, badges, MIT
```
