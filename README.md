# mr-ducksafe 🦆

Windows makinendeki kurulu sistem yazılımlarını **NVD CVE veritabanına** karşı tarayan, açıkları **insan dilinde** ve **uygulanabilir çözüm önerileriyle** raporlayan tek-dosya Go aracı.

- 🔍 Tek `.exe` — kurulum yok, runtime yok
- 🌐 Veri kaynağı: [NVD CVE API 2.0](https://nvd.nist.gov/) (resmi, CPE tabanlı)
- 📊 Rehber görünümü: 500+ CVE değil, ~10 aksiyon maddesi
- 🛠️ Her bulgu için `winget upgrade` komutu veya güncelleme linki
- 🖥️ İsteğe bağlı lokal web UI (`mr-ducksafe serve`)
- 🔒 Tarama tamamen lokal; sadece OSV/NVD'ye `{vendor, product, version}` gider

**Platform:** MVP — Windows. Linux/macOS v0.2'de planlanıyor.

---

## Hızlı Başlangıç

### Kaynaktan build

Go 1.22+ gerekli.

```sh
git clone https://github.com/yemresalcan/mr-ducksafe
cd mr-ducksafe
go build -o mr-ducksafe.exe ./cmd/mr-ducksafe
```

### NVD API key (önerilen)

Key'siz rate limit: 5 istek / 30 saniye (yavaş).
Key'li: 50 / 30 saniye (~10x hızlı).

[Ücretsiz key al](https://nvd.nist.gov/developers/request-an-api-key), sonra:

```powershell
setx NVD_API_KEY "<key>"
```
(Yeni terminal açmak gerek — `setx` mevcut oturumu etkilemez.)

---

## Kullanım

### Tarama (rehberli özet — varsayılan)

```sh
mr-ducksafe scan
```

Çıktı: uygulama bazında gruplu, severity'ye göre sıralı, her biri için **risk açıklaması + çözüm komutu**.

```
mr-ducksafe — güncellenmesi gereken 6 yazılım (521 açık)

  KRIT  Google Chrome  148.0.7778.179
          1 kritik · 1 yüksek · 6 orta
          Aciliyet: Acilen güncelle — aktif sömürülebilir.
          Risk:     Bellek hatası yoluyla uygulama ele geçirilebilir.
          Çözüm:    winget upgrade --id Google.Chrome
  ...
```

### Detay görünümleri

```sh
mr-ducksafe scan --details            # tüm CVE tablosu
mr-ducksafe scan --app "Chrome"       # sadece adı eşleşen uygulama
mr-ducksafe scan --show-unmatched     # eşleştirilemeyenleri de göster
```

### Çıktı formatları

```sh
mr-ducksafe scan --format=json        # script/CI için
mr-ducksafe scan --format=markdown    # PR/issue'ya yapıştırılabilir
```

### CI modu

```sh
mr-ducksafe scan --fail-on=high       # high+ bulgu → exit code 1
```

### Lokal web UI

```sh
mr-ducksafe serve --port 8080
# http://127.0.0.1:8080 aç
```

Dashboard'dan "Şimdi Tara" butonu, geçmiş liste, scan detay görüntülenebilir.
Default `127.0.0.1`'e bind olur. Ağa açmak için `--bind 0.0.0.0` (auth yok, dikkatli kullan).

### Eski scan kayıtlarını sil

```sh
mr-ducksafe prune --older-than=30d
```

### Sürüm

```sh
mr-ducksafe version
```

---

## Exit Code'ları

| Code | Anlam |
|---|---|
| 0 | Başarılı, `--fail-on` eşiği aşılmadı |
| 1 | Eşik aşıldı (`--fail-on`) |
| 2 | Tarama hatası (NVD erişilemedi, scanner başarısız) |
| 64 | Kullanım hatası (yanlış flag) |

## Gizlilik

- Tarama tamamen lokaldir.
- NVD'ye giden tek veri: `?virtualMatchString=cpe:2.3:a:vendor:product` (HTTPS).
- Makine adı, kullanıcı bilgisi, IP içermez (TLS doğrulamalı).
- Sonuçlar `~/.mr-ducksafe/mr-ducksafe.db` içinde lokal SQLite'ta.

## Bilinen Sınırlamalar (MVP)

- **50 uygulama mapping'i:** Top yazılımları kapsar; geri kalan "eşleştirilemedi" olarak görünür. Topluluk PR'larıyla genişler.
- **Versiyon şeması farklı uygulamalar** (örn. Adobe Acrobat DC continuous track) için "Düzeltme" sütunu boş kalabilir — NVD `versionEndIncluding` kullandığında.
- **Sadece kategori-bazlı öğreticilik (Seviye A):** Her CVE için tek tek değil, severity + anahtar kelimeden türeyen açıklamalar. LLM destekli Seviye B v0.2'de planlanıyor.

## Geliştirme

```sh
go test ./...                  # tüm testler
go test ./... -v               # verbose
golangci-lint run              # lint
```

Tasarım: [docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md](docs/superpowers/specs/2026-05-23-mr-ducksafe-design.md)
Plan: [docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md](docs/superpowers/plans/2026-05-24-mr-ducksafe-mvp.md)
Rehberli görünüm tasarımı: [docs/superpowers/specs/2026-05-28-guided-remediation-design.md](docs/superpowers/specs/2026-05-28-guided-remediation-design.md)

## Lisans

Henüz belirlenmedi (MIT veya Apache-2.0). MVP build'inden önce eklenecek.

## Yeni Mapping Eklemek

`internal/matcher/data/mappings.yaml` dosyasına PR aç. Her giriş:

```yaml
- registry_name: "App Name"      # veya registry_name_pattern (regex)
  publisher: "Vendor"
  identifier:
    type: cpe
    value: "cpe:2.3:a:vendor:product"
  version_normalize: "semver"    # semver | fourpart | calver
  version_from_name: "(\\d+\\.\\d+)"  # opsiyonel — isimden versiyon çıkar
  winget_id: "Vendor.Product"    # opsiyonel — remediation komutu
  update_url: "https://..."      # opsiyonel — manuel güncelleme linki
```
