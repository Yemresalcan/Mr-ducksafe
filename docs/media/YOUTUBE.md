# YouTube yükleme metinleri

Videolar: [`mr-ducksafe-youtube.mp4`](mr-ducksafe-youtube.mp4) (1920×1080, 1:26) ·
[`mr-ducksafe-short.mp4`](mr-ducksafe-short.mp4) (1080×1920, 0:35) ·
Thumbnail: [`thumbnail.png`](thumbnail.png) (1280×720)

Yükledikten sonra `README.md` ve `README.en.md` içindeki "Tanıtım Videosu / Promo Video"
bölümünde `docs/media/*.mp4` linklerini YouTube URL'leriyle değiştir.

---

## 1. YouTube videosu (yatay)

**Başlık** (≤ 100 karakter)

```
Bilgisayarında 525 Güvenlik Açığı mı Var? Tek Komutla Bul: mr-ducksafe (Ücretsiz CVE Tarayıcı)
```

Alternatifler:
- `525 Açık → 14 Aksiyon: Windows'ta Güncel Olmayan Yazılımları Tek Komutla Bul`
- `Windows İçin Ücretsiz ve Açık Kaynak CVE Tarayıcı — mr-ducksafe Tanıtım`

**Açıklama**

```
Bilgisayarındaki Chrome, Adobe Acrobat, Office, AnyDesk, Node.js… eski sürümleri yüzlerce bilinen güvenlik açığı (CVE) taşıyor. mr-ducksafe, Windows'ta kurulu yazılımları tarar, resmi NVD CVE veritabanıyla eşleştirir ve sonucu "525 açık" yerine "14 aksiyon" olarak, sade dilde ve çözüm komutuyla birlikte verir.

🦆 GitHub: https://github.com/yemresalcan/mr-ducksafe

✅ Tek .exe — kurulum, runtime, hesap yok
✅ Resmi veri: NVD CVE API 2.0 (CPE + sürüm aralığı)
✅ Riskler sade dilde: "bellek hatası yoluyla uygulama ele geçirilebilir"
✅ Çözüm hazır: winget upgrade --id Google.Chrome
✅ CI/CD: --fail-on=high, --format=json, exit kodları
✅ %100 yerel tarama, telemetri yok — NVD'ye sadece {vendor, product, version} gider
✅ Ücretsiz, açık kaynak (MIT)

⏱️ Bölümler
0:00 Giriş
0:05 Sorun: eski sürümler
0:13 mr-ducksafe nedir?
0:18 Nasıl çalışır?
0:28 Demo: mr-ducksafe scan
0:43 Sade dilde risk + çözüm
0:50 Demo: --app ve --details
0:58 CI/CD: --fail-on ve JSON
1:07 Geçmiş ve gizlilik
1:17 Kurulum ve bağlantılar

📦 Kurulum
• GitHub Releases → mr-ducksafe.exe
• go install github.com/yemresalcan/mr-ducksafe/cmd/mr-ducksafe@latest

🔑 Ücretsiz NVD API key (10x hızlı tarama): https://nvd.nist.gov/developers/request-an-api-key

#güvenlik #cybersecurity #CVE #Windows #açıkkaynak #golang #devsecops
```

**Etiketler**

```
mr-ducksafe, cve tarayıcı, cve scanner, güvenlik açığı, zafiyet taraması, vulnerability scanner, nvd, national vulnerability database, windows güvenlik, winget, devsecops, siber güvenlik, cybersecurity, açık kaynak, open source, golang, go cli, ci cd güvenlik, yazılım güncelleme, sbom
```

- Kategori: **Bilim ve Teknoloji**
- Dil: **Türkçe**
- Thumbnail: `thumbnail.png`

---

## 2. YouTube Shorts (dikey)

**Başlık**

```
Bilgisayarında kaç güvenlik açığı var? 🦆 #shorts
```

**Açıklama**

```
Tek komut: mr-ducksafe scan
525 bilinen açık → 14 net aksiyon, sade dilde risk ve winget çözüm komutu.
Ücretsiz & açık kaynak: https://github.com/yemresalcan/mr-ducksafe

#shorts #cybersecurity #güvenlik #windows #cve #opensource #devsecops
```

> Shorts'ta altyazı/başlık alanı ekranın alt ~%20'sini kaplar; videodaki önemli
> içerik bu bölgenin dışında tutuldu.
