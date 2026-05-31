<div align="center">

# 🦆 mr-ducksafe

**Windows kurulu yazılımları için NVD CVE tarayıcı — tek dosya, rehberli.**

500+ CVE değil → ~10 aksiyon maddesi. Risk insan dilinde, çözüm tek komut.

[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/platform-Windows-blue)]()
[![Status](https://img.shields.io/badge/status-MVP-yellow)]()
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
- 🖥️ **Opsiyonel web UI:** `mr-ducksafe serve` → lokal dashboard
- 🔒 **Gizlilik:** tarama tamamen lokal; sadece anonim `{vendor, product, version}` NVD'ye gider

**Platform:** Windows. Linux/macOS v0.2'de planlanıyor.

---

## Hızlı Başlangıç

### 1. Build

Go 1.22+ gerekli.

```sh
git clone https://github.com/yemresalcan/mr-ducksafe
cd mr-ducksafe
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
```

### 2. NVD API key (önerilen — 10x hızlı)

[Ücretsiz key al](https://nvd.nist.gov/developers/request-an-api-key), sonra:

```powershell
setx NVD_API_KEY "<key>"
```
> Yeni terminal aç. `setx` mevcut oturumu etkilemez.

### 3. Tara

```sh
mr-ducksafe scan
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
      --show-unmatched   eşleştirilemeyen yazılımları da göster
```

**Örnekler:**

```sh
mr-ducksafe scan                      # rehberli özet (varsayılan)
mr-ducksafe scan --details            # tüm CVE tablosu
mr-ducksafe scan --app "Chrome"       # sadece Chrome
mr-ducksafe scan --show-unmatched     # eşleştirilemeyenleri de göster
mr-ducksafe scan --format=json        # CI/script için
mr-ducksafe scan --format=markdown    # PR/issue'ya yapıştırılabilir
mr-ducksafe scan --fail-on=high       # high+ bulgu → exit 1 (CI)
```

### `serve` — lokal web UI

```sh
mr-ducksafe serve --port 8080
# tarayıcıdan http://127.0.0.1:8080
```

Dashboard'dan "Şimdi Tara" butonu, geçmiş liste, scan detay.
Default `127.0.0.1`. Ağa açmak için `--bind 0.0.0.0` (auth yok — dikkat).

### `prune` — eski kayıtları sil

```sh
mr-ducksafe prune --older-than=30d    # 30 günden eski scan'leri sil
```

### `version`

```sh
$ mr-ducksafe version
mr-ducksafe dev
```

### Tüm komutlar

```
$ mr-ducksafe --help

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  prune       Belirli yaştan eski scan kayıtlarını sil
  scan        Kurulu yazılımları tarayıp CVE raporu üret
  serve       Lokal web UI başlat
  version     mr-ducksafe sürümünü göster
```

---

## Exit Code'ları

| Code | Anlam |
|---|---|
| 0 | Başarılı, `--fail-on` eşiği aşılmadı |
| 1 | Eşik aşıldı (`--fail-on`) |
| 2 | Tarama hatası (NVD erişilemedi, scanner başarısız) |
| 64 | Kullanım hatası (yanlış flag) |

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
cmd/mr-ducksafe ──► internal/cmd (Cobra: scan/serve/prune/version)
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
                  internal/server  ──► lokal HTTP + embedded UI
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
go test ./...                           # tüm testler
go test ./... -v                        # verbose
golangci-lint run                       # lint

# Cross-platform release build (Windows host)
.\scripts\build.ps1 -Version "0.1.0"
# dist/ altında: windows-amd64, linux-amd64, darwin-amd64, darwin-arm64 (~15 MB her biri)
```

**Dokümanlar:**
- Tasarım: [docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md](docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md)
- Implementasyon planı: [docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md](docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md)
- Rehberli görünüm tasarımı: [docs/superpowers/specs/2026-05-28-guided-remediation-design.md](docs/superpowers/specs/2026-05-28-guided-remediation-design.md)

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

---

## Yol Haritası

**v0.1 (mevcut MVP)** ✅
- Windows Registry tarayıcı
- NVD CPE + version range eşleştirme
- 50 curated mapping
- Rehberli özet + winget remediation + Seviye A açıklama
- CLI + lokal web UI

**v0.2 (planlanan)**
- Linux scanner (dpkg, rpm, pacman)
- macOS scanner (Homebrew, `/Applications`)
- SSE canlı progress (web UI'da spinner)
- Mapping listesi 200+'e (topluluk PR'larıyla)
- "Düzeltme" sütununda `versionEndIncluding` desteği

**v0.3+ (fikir)**
- Seviye B: `--explain` flag'i + opsiyonel LLM ile CVE-başına açıklama
- Browser extension / VS Code extension tarama
- `--auto-fix` (sorumluluk uyarısıyla)

---

## Lisans

Henüz belirlenmedi (MIT veya Apache-2.0). MVP build'inden önce eklenecek.

---

<div align="center">

**Açıkları bulup düzeltmek bu kadar basit olmamalıydı.** 🦆

</div>
