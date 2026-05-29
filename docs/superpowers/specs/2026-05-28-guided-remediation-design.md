# Rehberli Remediation + Öğretici Görünüm — Design (Seviye A)

**Tarih:** 2026-05-28
**Durum:** Onaylandı (kullanıcı), uygulamaya hazır
**Bağlam:** MVP'nin (Task 1-20b) üzerine eklenir. mr-ducksafe'i "rapor aracı"ndan "rehber araç"a taşır.

---

## 1. Amaç

Kullanıcı taramayı bitirdiğinde:
1. **Boğulmasın** — 521 CVE satırı yerine ~6 uygulama bazlı aksiyon
2. **Anlasın** — "CVE-2022-26485" yerine insan dilinde risk açıklaması (Seviye A: kural-tabanlı, LLM yok)
3. **Düzeltsin** — her uygulama için net bir sonraki adım (winget komutu / indirme linki)

## 2. Seviye seçimi

- **Seviye A (bu spec):** Statik/kural-tabanlı açıklamalar. Offline, maliyetsiz, "tek binary" felsefesine uygun. Açıklamalar severity + kategori bazında.
- **Seviye B (gelecek sürüm):** LLM destekli, CVE-başına kişiselleştirilmiş açıklama. `--explain` flag'i + opsiyonel API key. Bu spec kapsamı dışında.

## 3. Yeni varsayılan görünüm (grouped summary)

Bulgular **uygulama bazında** gruplanır, en yüksek severity'ye göre sıralanır:

```
mr-ducksafe — guncellenmesi gereken 6 yazilim (521 acik)

  KRITIK  Adobe Acrobat Reader DC        20.012.20043
          409 acik · 7 kritik · 180 yuksek
          Risk:  Kotu niyetli bir dosya actirilarak bilgisayarinda
                 uzaktan kod calistirilabilir.
          Cozum: winget upgrade --id Adobe.Acrobat.Reader.64-bit

  YUKSEK  Google Chrome                  120.0.6099.130
          12 acik · 12 yuksek
          Risk:  Kotu niyetli bir web sitesi tarayicinizda kod
                 calistirabilir veya verinize erisebilir.
          Cozum: winget upgrade --id Google.Chrome

  ...

  Detay:  mr-ducksafe scan --details        (tum CVE listesi)
          mr-ducksafe scan --app "Chrome"   (tek uygulama)
```

`--show-unmatched` ve `--format=json|markdown` davranışları korunur (JSON tam veriyi içerir).

## 4. Öğretici katman (Seviye A kuralları)

`internal/advice` paketi. Girdi: worst severity + bulgu özetlerindeki anahtar kelimeler. Çıktı: tek satır Türkçe risk açıklaması.

**Severity → aciliyet çerçevesi:**
| Severity | Çerçeve |
|---|---|
| critical | "Acilen guncelle — aktif somurulabilir." |
| high | "Yakinda guncelle — ciddi risk." |
| medium | "Firsat buldukca guncelle." |
| low | "Dusuk oncelik." |

**Anahtar kelime → risk cümlesi** (CVE özetlerinde aranır, ilk eşleşen kullanılır):
| Anahtar | Açıklama |
|---|---|
| "remote code execution", "rce", "arbitrary code" | "...uzaktan kod calistirilabilir." |
| "cross-site scripting", "xss" | "...tarayicinizda kotu niyetli script calisabilir." |
| "use-after-free", "buffer overflow", "memory" | "...bellek hatasi yoluyla ele gecirme mumkun." |
| "privilege escalation" | "...saldirgan yetkilerini yukseltebilir." |
| "denial of service", "dos" | "...uygulama cokertilerek hizmet engellenebilir." |
| "information disclosure", "leak" | "...gizli bilgileriniz sizdirilabilir." |
| (eşleşme yok) | severity bazlı genel cümle |

## 5. Remediation verisi

`mappings.yaml`'a opsiyonel alanlar:
```yaml
- registry_name: "Google Chrome"
  ...
  winget_id: "Google.Chrome"
  update_url: "https://www.google.com/chrome/"
```
- `winget_id` varsa → "winget upgrade --id X" gösterilir
- yoksa `update_url` varsa → "Guncelle: <url>"
- ikisi de yoksa → "Uygulamayi en son surume guncelleyin."

## 6. Mimari değişiklikler

- **matcher:** `mappingEntry`'ye `WingetID`, `UpdateURL`. `Identifier`'a `WingetID`, `UpdateURL` (Match'te doldurulur).
- **scanengine:** `ScanResult`'a `Remediation map[string]Remediation` (appName → {WingetID, UpdateURL}). Engine match sırasında doldurur. (Persist edilmez — sunum verisi.)
- **internal/advice:** `RiskLine(worstSeverity, summaries) string` ve `Urgency(severity) string`.
- **report:** yeni `writeSummary` (varsayılan); mevcut tablo `--details` altına taşınır. Gruplama + advice + remediation render.
- **cmd/scan:** `--details` (tam tablo), `--app <isim>` (filtre) flag'leri.

## 7. Test stratejisi

- `advice`: severity + keyword → beklenen cümle (table-driven)
- `report`: summary görünümü golden file (gruplama + remediation + risk satırı)
- `matcher`: winget_id/update_url'in Identifier'a aktığı test

## 8. Task dökümü

| Task | İçerik |
|---|---|
| **R1** | mappings remediation alanları + matcher/Identifier plumbing + ScanResult.Remediation |
| **R2** | `internal/advice` paketi (severity + keyword → Türkçe risk) |
| **R3** | report summary görünümü (varsayılan) + golden test |
| **R4** | scan `--details` / `--app` flag'leri + wiring |

## 9. Kapsam dışı (gelecek)

- Seviye B (LLM açıklama)
- `--auto-fix` (winget'i otomatik çalıştırma)
- Tüm 50 mapping için winget_id (başta popüler ~15'i; gerisi topluluk PR)
