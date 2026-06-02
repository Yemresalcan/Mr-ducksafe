<div align="center">

# 🦆 mr-ducksafe

**Windows kurulu yazılımları için NVD CVE tarayıcı — tek binary, terminalden.**

500+ CVE değil → ~10 aksiyon maddesi. Risk insan dilinde, çözüm tek komut.

[![CI](https://github.com/yemresalcan/mr-ducksafe/actions/workflows/ci.yml/badge.svg)](https://github.com/yemresalcan/mr-ducksafe/actions/workflows/ci.yml)
[![Release](https://github.com/yemresalcan/mr-ducksafe/actions/workflows/release.yml/badge.svg)](https://github.com/yemresalcan/mr-ducksafe/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/yemresalcan/mr-ducksafe/branch/master/graph/badge.svg)](https://codecov.io/gh/yemresalcan/mr-ducksafe)
[![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/platform-Windows-blue)]()
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Data](https://img.shields.io/badge/data-NVD%202.0-red)](https://nvd.nist.gov/)

</div>

---

## Neden?

Bilgisayarındaki Adobe Acrobat, Chrome, Office, Java... bunların **eski sürümlerinde** ciddi güvenlik açıkları var. Kimse manuel takip etmek istemez. `mr-ducksafe` bunu otomatikleştirir:

- 🔍 **Tek `.exe`** — kurulum, runtime, hesap yok
- 🌐 **Doğru veri kaynağı:** [NVD CVE API 2.0](https://nvd.nist.gov/) (resmi, CPE tabanlı, masaüstü için ideal)
- 📊 **Rehberli özet:** uygulama bazında gruplu, severity'ye göre sıralı
- 🧠 **İnsan dilinde risk:** "Use-after-free in XSLT" değil → "bellek hatası yoluyla uygulama ele geçirilebilir"
- 🛠️ **Aksiyon odaklı:** her bulgu için `winget upgrade` veya güncelleme linki
- 🗂️ **Lokal geçmiş:** her tarama SQLite'a yazılır, `history` ile erişilir
- ⚙️ **CI dostu:** `--format=json`, `--output FILE`, `--fail-on=high`, `NO_COLOR` desteği
- 🔒 **Gizlilik:** tarama tamamen lokal; sadece anonim `{vendor, product, version}` NVD'ye gider

**Platform:** Windows. Linux/macOS v0.2'de planlanıyor.

---

## Kurulum

### Hazır binary (önerilen)

[GitHub Releases](https://github.com/yemresalcan/mr-ducksafe/releases) üzerinden işletim sistemine uygun arşivi indir, `mr-ducksafe.exe` (veya `mr-ducksafe`) dosyasını PATH'e koy.

Her release için **SHA256 checksum** ve **SBOM** (`*.sbom.json`, CycloneDX) yayınlanıyor.

### Scoop (Windows)

```powershell
# Manifest URL ile direkt kurulum (bucket gerekmez):
scoop install https://github.com/yemresalcan/mr-ducksafe/releases/latest/download/mr-ducksafe.json
```

> Scoop manifest sonraki sürümde release artifact'i olarak yayınlanacak.

### Kaynaktan build

Go 1.25+ gerekli.

```sh
git clone https://github.com/yemresalcan/mr-ducksafe
cd mr-ducksafe
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
```

### NVD API key (önerilen — 10x hızlı)

[Ücretsiz key al](https://nvd.nist.gov/developers/request-an-api-key), sonra:

```powershell
setx NVD_API_KEY "<key>"
```
> Yeni terminal aç. `setx` mevcut oturumu etkilemez.

---

## Hızlı Kullanım

```sh
mr-ducksafe scan                       # rehberli özet
mr-ducksafe scan --details             # tüm CVE tablosu
mr-ducksafe scan --app "Chrome"        # sadece Chrome
mr-ducksafe scan -o report.json --format=json  # CI/script için dosyaya yaz
mr-ducksafe scan --fail-on=high        # high+ bulgu → exit 1 (CI)
mr-ducksafe history                    # geçmiş tarama kayıtları
mr-ducksafe prune --older-than=30d     # eski kayıtları sil
```

---

## Örnek Çıktı

```
mr-ducksafe — güncellenmesi gereken 14 yazılım (521 açık)

  KRIT  Google Chrome  148.0.7778.179
          1 kritik · 1 yüksek · 6 orta · 4 düşük
          Aciliyet: Acilen güncelle — aktif sömürülebilir.
          Risk:     Bellek hatası yoluyla uygulama ele geçirilebilir.
          Çözüm:    winget upgrade --id Google.Chrome

  KRIT  AnyDesk  ad 9.6.1
          3 kritik · 6 yüksek · 1 orta
          Aciliyet: Acilen güncelle — aktif sömürülebilir.
          Risk:     Uzaktan, internet üzerinden bilgisayarınızda kod çalıştırılabilir.
          Çözüm:    Uygulamayı en son sürüme güncelleyin.

  KRIT  Microsoft OneDrive  26.078.0426.0002
          1 kritik · 6 yüksek · 2 orta
          Aciliyet: Acilen güncelle — aktif sömürülebilir.
          Risk:     Saldırgan sistemdeki yetkilerini yükseltebilir.
          Çözüm:    Uygulamayı en son sürüme güncelleyin.

  KRIT  Node.js  22.17.0
          1 kritik · 3 yüksek · 1 orta
          Aciliyet: Acilen güncelle — aktif sömürülebilir.
          Risk:     Uzaktan, internet üzerinden bilgisayarınızda kod çalıştırılabilir.
          Çözüm:    winget upgrade --id OpenJS.NodeJS

  YÜKS  Adobe Acrobat Reader DC - Turkish  20.012.20043
          247 yüksek · 135 orta · 27 düşük
          Aciliyet: Yakında güncelle — ciddi risk.
          Risk:     Uzaktan, internet üzerinden bilgisayarınızda kod çalıştırılabilir.
          Çözüm:    winget upgrade --id Adobe.Acrobat.Reader.64-bit

  ...

  Detay:  mr-ducksafe scan --details        (tüm CVE listesi)
          mr-ducksafe scan --app "<isim>"   (tek uygulama)
```

> **Eski Adobe Acrobat'ın 409 açığının çoğu gerçektir** — 2020 sürümünden beri güncellenmediği için. Araç tam da bunu yakalamalı.

---

## Komutlar

### `scan` — taramayı çalıştır

```
$ mr-ducksafe scan --help

Kurulu yazılımları tarayıp CVE raporu üret

Flags:
      --app string       sadece adı bu metni içeren uygulamayı göster
      --details          özet yerine tüm CVE'leri tablo halinde göster
      --fail-on string   eşik aşıldığında exit 1: critical | high | medium | low
      --format string    çıktı formatı: table | json | markdown (default "table")
  -o, --output string    raporu bu dosyaya yaz (varsayılan: stdout)
      --show-unmatched   eşleştirilemeyen yazılımları da göster
```

### `history` — geçmiş taramaları listele

```
$ mr-ducksafe history --limit 5

ID  Tarih                Durum    Süre  Bulgu
12  2026-06-03 14:21:08  success  18s   521
11  2026-06-02 09:14:22  success  21s   519
10  2026-06-01 18:02:11  failed   4s    -
```

### `prune` — eski kayıtları sil

```sh
mr-ducksafe prune --older-than=30d
```

### `version`

```sh
$ mr-ducksafe version
mr-ducksafe v0.1.0
```

### Global flag'ler

| Flag | Açıklama |
|---|---|
| `--db PATH` | SQLite DB yolu (default: `~/.mr-ducksafe/mr-ducksafe.db`) |
| `--no-color` | Renkli çıktıyı kapat (`NO_COLOR` env de onurlandırılır) |
| `--verbose` | Ayrıntılı log |

---

## Exit Code'ları

| Code | Anlam |
|---|---|
| 0 | Başarılı, `--fail-on` eşiği aşılmadı |
| 1 | Eşik aşıldı (`--fail-on`) veya CLI hatası |
| 2 | Tarama hatası (NVD erişilemedi, scanner başarısız) |

---

## Gizlilik

- Tarama tamamen **lokal**.
- NVD'ye giden tek istek: `GET ?virtualMatchString=cpe:2.3:a:vendor:product` (HTTPS).
- Makine adı, kullanıcı bilgisi, IP **içermez**.
- Sonuçlar `~/.mr-ducksafe/mr-ducksafe.db` içinde lokal SQLite'ta.
- Hiçbir telemetri yok.

---

## Mimari (özet)

```
cmd/mr-ducksafe ──► internal/cmd (Cobra: scan/history/prune/version)
                         │
                         ▼
                  internal/scanengine ──► orkestrasyon (goroutine + semaphore)
                         │
        ┌────────────────┼────────────────┬────────────────┐
        ▼                ▼                ▼                ▼
   internal/      internal/        internal/        internal/
    scanner        matcher          vuln             store
   (Win Reg)    (50 mapping)      (Client iface)    (SQLite)
                                       ├─ nvd       (~/.mr-ducksafe/...)
                                       └─ osv

                  internal/report  ──► table | json | markdown | özet
                  internal/advice  ──► severity + keyword → risk metni
                  internal/tui     ──► banner + spinner
```

**Anahtar prensip:** Veri kaynağı (`vuln.Client`) interface arkasında. NVD ve OSV ikisi de implement eder, scanengine fark etmez.

---

## Bilinen Sınırlamalar (MVP)

- **50 uygulama mapping'i** — top yazılımları kapsar; geri kalan "eşleştirilemedi" görünür. Topluluk PR'larıyla genişler.
- **Versiyon şeması farklı uygulamalarda** "Düzeltme" sütunu boş kalabilir (NVD `versionEndIncluding` kullandığında — Adobe gibi).
- **Sadece kategori-bazlı öğreticilik (Seviye A):** Her CVE için değil, severity + anahtar kelimeden türeyen açıklamalar. LLM destekli Seviye B v0.2'de planlanıyor.
- **Build numarası kullanan apps** (örn. eski JetBrains) için `version_from_name` regex'i `mappings.yaml`'a eklenmeli (örnek mevcut).

---

## Geliştirme

```sh
go test ./... -race                     # tüm testler
go test ./... -coverprofile=cover.out   # coverage
golangci-lint run                       # lint

# Lokal cross-platform release smoke (goreleaser kurulu olmalı)
goreleaser release --snapshot --clean

# Manuel cross-platform build
.\scripts\build.ps1 -Version "0.1.0"
./scripts/build.sh 0.1.0
```

**Dokümanlar:**
- Tasarım: [docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md](docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md)
- CLI-only pivot: [docs/superpowers/specs/2026-06-03-cli-only-pivot-design.md](docs/superpowers/specs/2026-06-03-cli-only-pivot-design.md)
- Implementasyon planı: [docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md](docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md)

---

## Katkı: Yeni Mapping Eklemek

`internal/matcher/data/mappings.yaml` dosyasına PR aç. Her giriş:

```yaml
- registry_name: "App Name"          # veya registry_name_pattern (regex)
  publisher: "Vendor"
  identifier:
    type: cpe
    value: "cpe:2.3:a:vendor:product"
  version_normalize: "semver"        # semver | fourpart | calver
  version_from_name: "(\\d+\\.\\d+)" # opsiyonel — isimden versiyon çıkar
  winget_id: "Vendor.Product"        # opsiyonel — remediation komutu
  update_url: "https://..."          # opsiyonel — manuel güncelleme linki
```

CPE'leri [nvd.nist.gov/products/cpe/search](https://nvd.nist.gov/products/cpe/search) üzerinden doğrula.

Bug, feature veya yeni mapping önerisi için [issue template'lerini](.github/ISSUE_TEMPLATE/) kullan.

---

## Yol Haritası

**v0.1 (mevcut MVP)** ✅
- Windows Registry tarayıcı
- NVD CPE + version range eşleştirme
- 50 curated mapping
- Rehberli özet + winget remediation + Seviye A açıklama
- CLI (`scan` / `history` / `prune`), SQLite geçmiş, JSON/Markdown çıktı
- Multi-platform release (Windows/Linux/macOS binary) + SBOM

**v0.2 (planlanan)**
- Linux scanner (dpkg, rpm, pacman)
- macOS scanner (Homebrew, `/Applications`)
- Mapping listesi 200+'e (topluluk PR'larıyla)
- "Düzeltme" sütununda `versionEndIncluding` desteği
- Scoop bucket + Homebrew tap

**v0.3+ (fikir)**
- Seviye B: `--explain` flag'i + opsiyonel LLM ile CVE-başına açıklama
- Browser extension / VS Code extension tarama
- `--auto-fix` (sorumluluk uyarısıyla)

---

## Lisans

[MIT](LICENSE) — Copyright (c) 2026 Yunus Emre Salcan

---

<div align="center">

**Açıkları bulup düzeltmek bu kadar basit olmamalıydı.** 🦆

</div>
