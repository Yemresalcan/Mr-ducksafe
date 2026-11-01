package i18n

type msg = map[Lang]string

var messages = map[string]msg{
	// ---- language selection -------------------------------------------
	"lang.saved": {
		EN: "Language set to English. Change it any time with: mr-ducksafe lang tr",
		TR: "Dil Türkçe olarak ayarlandı. İstediğin zaman değiştir: mr-ducksafe lang en",
	},
	"lang.short":   {EN: "Show or change the output language (en | tr)", TR: "Çıktı dilini göster veya değiştir (en | tr)"},
	"lang.current": {EN: "Current language: %s", TR: "Geçerli dil: %s"},
	"lang.invalid": {EN: "unknown language %q (use en or tr)", TR: "bilinmeyen dil %q (en veya tr kullan)"},
	"flag.lang":    {EN: "output language for this run: en | tr", TR: "bu çalıştırma için çıktı dili: en | tr"},

	// ---- root / global flags ------------------------------------------
	"root.short": {EN: "CVE scanner for locally installed software", TR: "Yerel kurulu yazılımlar için CVE tarayıcı"},
	"root.long": {
		EN: "mr-ducksafe scans the software installed on your machine and reports known vulnerabilities from the NVD.",
		TR: "mr-ducksafe makinendeki kurulu sistem yazılımlarını tarayıp NVD üzerinden bilinen güvenlik açıklarını raporlar.",
	},
	"flag.verbose": {EN: "verbose logging", TR: "ayrıntılı log çıktısı"},
	"flag.db":      {EN: "SQLite DB path (default: ~/.mr-ducksafe/mr-ducksafe.db)", TR: "SQLite DB yolu (varsayılan: ~/.mr-ducksafe/mr-ducksafe.db)"},
	"flag.nocolor": {EN: "disable colored output (NO_COLOR is honored too)", TR: "renkli çıktıyı kapat (NO_COLOR env de onurlandırılır)"},
	"err.prefix":   {EN: "error:", TR: "hata:"},
	"err.opendb":   {EN: "open DB: %w", TR: "DB aç: %w"},

	// ---- double-click mode --------------------------------------------
	"pause.enter": {EN: "\nPress Enter to close this window... ", TR: "\nPencereyi kapatmak için Enter'a basın... "},

	// ---- version --------------------------------------------------------
	"version.short": {EN: "Show the mr-ducksafe version", TR: "mr-ducksafe sürümünü göster"},

	// ---- scan -------------------------------------------------------------
	"scan.short":         {EN: "Scan installed software and produce a CVE report", TR: "Kurulu yazılımları tarayıp CVE raporu üret"},
	"flag.format":        {EN: "output format: table | json | markdown", TR: "çıktı formatı: table | json | markdown"},
	"flag.failon":        {EN: "exit 1 when the threshold is crossed: critical | high | medium | low", TR: "eşik aşıldığında exit 1: critical | high | medium | low"},
	"flag.showunmatched": {EN: "also show software that could not be matched", TR: "eşleştirilemeyen yazılımları da göster"},
	"flag.details":       {EN: "show every CVE as a table instead of the summary", TR: "özet yerine tüm CVE'leri tablo halinde göster"},
	"flag.app":           {EN: "only show applications whose name contains this text", TR: "sadece adı bu metni içeren uygulamayı göster"},
	"flag.output":        {EN: "write the report to this file (default: stdout)", TR: "raporu bu dosyaya yaz (varsayılan: stdout)"},
	"scan.tip.nokey": {
		EN: "tip: NVD_API_KEY is not set — the scan may be slow (5 requests/30s).",
		TR: "ipucu: NVD_API_KEY ortam değişkeni ayarlanmamış — tarama yavaş olabilir (5 istek/30sn).",
	},
	"scan.tip.getkey": {
		EN: "     Free key: https://nvd.nist.gov/developers/request-an-api-key",
		TR: "       Ücretsiz key: https://nvd.nist.gov/developers/request-an-api-key",
	},
	"scan.failed":       {EN: "scan failed:", TR: "tarama başarısız:"},
	"scan.err.output":   {EN: "open output file: %w", TR: "output dosyası aç: %w"},
	"scan.written":      {EN: "Report written: %s", TR: "Rapor yazıldı: %s"},
	"progress.scanning": {EN: "Scanning installed software...", TR: "Kurulu yazılımlar taranıyor..."},
	"progress.scanned":  {EN: "Scan complete — %d programs", TR: "Tarama tamamlandı — %d program"},
	"progress.matching": {EN: "Matching...", TR: "Eşleştiriliyor..."},
	"progress.matched":  {EN: "Matching complete — %d matches", TR: "Eşleştirme tamamlandı — %d eşleşme"},
	"progress.nvdstart": {EN: "Querying NVD...", TR: "NVD sorgusu başlıyor..."},
	"progress.nvd":      {EN: "Querying NVD... %d/%d", TR: "NVD sorgusu... %d/%d"},
	"progress.done":     {EN: "Done — %d findings", TR: "Tamamlandı — %d bulgu"},
	"banner.sub":        {EN: "CVE scanner for installed software", TR: "yerel kurulu yazilim CVE tarayici"},

	// ---- history / prune ------------------------------------------------
	"history.short":  {EN: "List past scans", TR: "Geçmiş tarama kayıtlarını listele"},
	"flag.limit":     {EN: "how many records to show", TR: "kaç kayıt gösterilsin"},
	"history.empty":  {EN: "No scans yet. Start with `mr-ducksafe scan`.", TR: "Henüz tarama kaydı yok. `mr-ducksafe scan` ile başla."},
	"history.header": {EN: "ID\tDate\tStatus\tTime\tFindings", TR: "ID\tTarih\tDurum\tSüre\tBulgu"},
	"prune.short":    {EN: "Delete scans older than a given age", TR: "Belirli yaştan eski scan kayıtlarını sil"},
	"flag.olderthan": {EN: "age: 7d, 24h, 30d, ...", TR: "süre: 7d, 24h, 30d gibi"},
	"prune.done":     {EN: "%d scan records deleted", TR: "%d scan kaydı silindi"},
	"prune.invalid":  {EN: "invalid duration: %s", TR: "geçersiz süre: %s"},

	// ---- summary report ---------------------------------------------------
	"sum.clean":      {EN: "%s %d installed programs scanned, no vulnerabilities found.", TR: "%s %d kurulu yazılım tarandı, açık bulunamadı."},
	"sum.title":      {EN: "mr-ducksafe — %s applications need updating (%d vulnerabilities)", TR: "mr-ducksafe — güncellenmesi gereken %s yazılım (%d açık)"},
	"sum.urgency":    {EN: "Urgency:", TR: "Aciliyet:"},
	"sum.risk":       {EN: "Risk:   ", TR: "Risk:    "},
	"sum.fix":        {EN: "Fix:    ", TR: "Çözüm:   "},
	"sum.details":    {EN: "  Details: mr-ducksafe scan --details        (full CVE list)\n           mr-ducksafe scan --app \"<name>\"   (single application)", TR: "  Detay:  mr-ducksafe scan --details        (tüm CVE listesi)\n          mr-ducksafe scan --app \"<isim>\"   (tek uygulama)"},
	"sum.unmatched":  {EN: "\nSoftware that could not be matched (%d):", TR: "\nEşleştirilemeyen yazılımlar (%d):"},
	"sum.update":     {EN: "Update: ", TR: "Güncelle: "},
	"sum.updateapp":  {EN: "Update the application to the latest version.", TR: "Uygulamayı en son sürüme güncelleyin."},
	"count.critical": {EN: "%d critical", TR: "%d kritik"},
	"count.high":     {EN: "%d high", TR: "%d yüksek"},
	"count.medium":   {EN: "%d medium", TR: "%d orta"},
	"count.low":      {EN: "%d low", TR: "%d düşük"},

	// ---- severity badges (4 columns wide) ------------------------------
	"sev.critical": {EN: "CRIT", TR: "KRIT"},
	"sev.high":     {EN: "HIGH", TR: "YÜKS"},
	"sev.medium":   {EN: "MED ", TR: "ORTA"},
	"sev.low":      {EN: "LOW ", TR: "DÜŞK"},
	"sev.unknown":  {EN: "UNK ", TR: "BLNM"},

	// ---- table report -----------------------------------------------------
	"tbl.title":     {EN: "mr-ducksafe scan #%d — %d installed programs scanned (%d unmatched)", TR: "mr-ducksafe scan #%d — %d kurulu yazılım tarandı (%d eşleştirilemedi)"},
	"tbl.findings":  {EN: "Findings:", TR: "Bulgular:"},
	"tbl.critical":  {EN: "CRITICAL", TR: "KRITIK"},
	"tbl.high":      {EN: "HIGH", TR: "YÜKSEK"},
	"tbl.medium":    {EN: "MEDIUM", TR: "ORTA"},
	"tbl.low":       {EN: "LOW", TR: "DÜŞÜK"},
	"tbl.none":      {EN: "✓ No vulnerabilities found.", TR: "✓ Açık güvenlik bulgusu yok."},
	"tbl.h.app":     {EN: "Application", TR: "Uygulama"},
	"tbl.h.ver":     {EN: "Version", TR: "Sürüm"},
	"tbl.h.fix":     {EN: "Fixed in", TR: "Düzeltme"},
	"tbl.h.sum":     {EN: "Summary", TR: "Özet"},
	"tbl.unmatched": {EN: "\nSoftware that could not be matched (%d):", TR: "\nEşleştirilemeyen yazılımlar (%d):"},
	"tbl.warnings":  {EN: "\n%d warnings", TR: "\n%d uyarı"},

	// ---- markdown report --------------------------------------------------
	"md.scanned":   {EN: "- Scanned: **%d** installed programs", TR: "- Tarandı: **%d** kurulu yazılım"},
	"md.unmatched": {EN: "- Unmatched: **%d**", TR: "- Eşleştirilemedi: **%d**"},
	"md.findings":  {EN: "- Findings: 🔴 %d critical · 🟠 %d high · 🟡 %d medium · 🔵 %d low", TR: "- Bulgular: 🔴 %d kritik · 🟠 %d yüksek · 🟡 %d orta · 🔵 %d düşük"},
	"md.none":      {EN: "✅ No vulnerabilities found.", TR: "✅ Açık güvenlik bulgusu yok."},
	"md.header":    {EN: "| Severity | Application | Version | CVE | Fixed in | Summary |", TR: "| Severity | Uygulama | Sürüm | CVE | Düzeltme | Özet |"},

	// ---- advice (plain-language risk) ----------------------------------
	"urg.critical":        {EN: "Update immediately — actively exploitable.", TR: "Acilen güncelle — aktif sömürülebilir."},
	"urg.high":            {EN: "Update soon — serious risk.", TR: "Yakında güncelle — ciddi risk."},
	"urg.medium":          {EN: "Update when you get a chance.", TR: "Fırsat buldukça güncelle."},
	"urg.low":             {EN: "Low priority.", TR: "Düşük öncelik."},
	"urg.unknown":         {EN: "An update is recommended.", TR: "Güncelleme önerilir."},
	"risk.rce":            {EN: "Code can be executed on your computer remotely, over the internet.", TR: "Uzaktan, internet üzerinden bilgisayarınızda kod çalıştırılabilir."},
	"risk.priv":           {EN: "An attacker can elevate their privileges on the system.", TR: "Saldırgan sistemdeki yetkilerini yükseltebilir."},
	"risk.memory":         {EN: "The application can be hijacked through a memory error.", TR: "Bellek hatası yoluyla uygulama ele geçirilebilir."},
	"risk.xss":            {EN: "A malicious script can run in your browser.", TR: "Kötü niyetli bir script tarayıcınızda çalıştırılabilir."},
	"risk.sqli":           {EN: "Unauthorized database access or manipulation may be possible.", TR: "Veritabanına yetkisiz erişim/manipülasyon mümkün olabilir."},
	"risk.leak":           {EN: "Your sensitive information can be leaked.", TR: "Gizli bilgileriniz sızdırılabilir."},
	"risk.dos":            {EN: "The application can be crashed to deny service.", TR: "Uygulama çökertilerek hizmet engellenebilir."},
	"risk.bypass":         {EN: "Security controls can be bypassed.", TR: "Güvenlik kontrolleri atlatılabilir."},
	"risk.generic.high":   {EN: "This version has serious vulnerabilities; an attacker could harm your system.", TR: "Bu sürümde ciddi güvenlik açıkları var; saldırgan sisteminize zarar verebilir."},
	"risk.generic.medium": {EN: "This version has medium-severity vulnerabilities.", TR: "Bu sürümde orta seviyeli güvenlik açıkları bulunuyor."},
	"risk.generic.low":    {EN: "This version has known vulnerabilities.", TR: "Bu sürümde bilinen güvenlik açıkları bulunuyor."},
}
