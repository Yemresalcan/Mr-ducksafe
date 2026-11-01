// Package advice güvenlik bulgularını insan diline çeviren kural-tabanlı
// (Seviye A) öğretici metinler üretir. LLM kullanmaz, tamamen offline.
package advice

import (
	"strings"

	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
)

// Urgency severity'ye göre aciliyet çerçevesi döner.
func Urgency(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "high", "medium", "low":
		return i18n.T("urg." + strings.ToLower(severity))
	default:
		return i18n.T("urg.unknown")
	}
}

// riskKeyword bir anahtar kelime grubunu insan dili açıklamaya (i18n anahtarı) eşler.
type riskKeyword struct {
	keywords []string
	message  string // i18n key
}

// Sıra önemli: ilk eşleşen kullanılır (en ciddi/spesifik üstte).
var riskRules = []riskKeyword{
	{[]string{"remote code execution", "arbitrary code", "rce"},
		"risk.rce"},
	{[]string{"privilege escalation", "elevation of privilege"},
		"risk.priv"},
	{[]string{"use-after-free", "buffer overflow", "out-of-bounds", "memory corruption"},
		"risk.memory"},
	{[]string{"cross-site scripting", "xss"},
		"risk.xss"},
	{[]string{"sql injection"},
		"risk.sqli"},
	{[]string{"information disclosure", "information leak", "sensitive data"},
		"risk.leak"},
	{[]string{"denial of service", "dos", "crash"},
		"risk.dos"},
	{[]string{"spoofing", "bypass"},
		"risk.bypass"},
}

// RiskLine bulgu özetlerini tarayıp en uygun risk açıklamasını döner.
// Hiçbir anahtar kelime eşleşmezse severity bazlı genel bir cümle döner.
func RiskLine(worstSeverity string, summaries []string) string {
	joined := strings.ToLower(strings.Join(summaries, " "))
	for _, r := range riskRules {
		for _, kw := range r.keywords {
			if strings.Contains(joined, kw) {
				return i18n.T(r.message)
			}
		}
	}
	return genericRisk(worstSeverity)
}

func genericRisk(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "high":
		return i18n.T("risk.generic.high")
	case "medium":
		return i18n.T("risk.generic.medium")
	default:
		return i18n.T("risk.generic.low")
	}
}
