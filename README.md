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
