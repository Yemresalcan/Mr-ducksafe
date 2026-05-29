// Package advice güvenlik bulgularını insan diline çeviren kural-tabanlı
// (Seviye A) öğretici metinler üretir. LLM kullanmaz, tamamen offline.
package advice

import "strings"

// Urgency severity'ye göre aciliyet çerçevesi döner.
func Urgency(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "Acilen güncelle — aktif sömürülebilir."
	case "high":
		return "Yakında güncelle — ciddi risk."
	case "medium":
		return "Fırsat buldukça güncelle."
	case "low":
		return "Düşük öncelik."
	default:
		return "Güncelleme önerilir."
	}
}

// riskKeyword bir anahtar kelime grubunu insan dili açıklamaya eşler.
type riskKeyword struct {
	keywords []string
	message  string
}

// Sıra önemli: ilk eşleşen kullanılır (en ciddi/spesifik üstte).
var riskRules = []riskKeyword{
	{[]string{"remote code execution", "arbitrary code", "rce"},
		"Uzaktan, internet üzerinden bilgisayarınızda kod çalıştırılabilir."},
	{[]string{"privilege escalation", "elevation of privilege"},
		"Saldırgan sistemdeki yetkilerini yükseltebilir."},
	{[]string{"use-after-free", "buffer overflow", "out-of-bounds", "memory corruption"},
		"Bellek hatası yoluyla uygulama ele geçirilebilir."},
	{[]string{"cross-site scripting", "xss"},
		"Kötü niyetli bir script tarayıcınızda çalıştırılabilir."},
	{[]string{"sql injection"},
		"Veritabanına yetkisiz erişim/manipülasyon mümkün olabilir."},
	{[]string{"information disclosure", "information leak", "sensitive data"},
		"Gizli bilgileriniz sızdırılabilir."},
	{[]string{"denial of service", "dos", "crash"},
		"Uygulama çökertilerek hizmet engellenebilir."},
	{[]string{"spoofing", "bypass"},
		"Güvenlik kontrolleri atlatılabilir."},
}

// RiskLine bulgu özetlerini tarayıp en uygun risk açıklamasını döner.
// Hiçbir anahtar kelime eşleşmezse severity bazlı genel bir cümle döner.
func RiskLine(worstSeverity string, summaries []string) string {
	joined := strings.ToLower(strings.Join(summaries, " "))
	for _, r := range riskRules {
		for _, kw := range r.keywords {
			if strings.Contains(joined, kw) {
				return r.message
			}
		}
	}
	return genericRisk(worstSeverity)
}

func genericRisk(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "high":
		return "Bu sürümde ciddi güvenlik açıkları var; saldırgan sisteminize zarar verebilir."
	case "medium":
		return "Bu sürümde orta seviyeli güvenlik açıkları bulunuyor."
	default:
		return "Bu sürümde bilinen güvenlik açıkları bulunuyor."
	}
}
