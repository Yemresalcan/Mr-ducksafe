// Package nvd NVD (National Vulnerability Database) CVE API 2.0'a sorgu
// yapan vuln.Client implementasyonudur. CPE tabanlı olduğu için masaüstü
// yazılımları için OSV'den daha doğru sonuç verir.
package nvd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

const defaultEndpoint = "https://services.nvd.nist.gov/rest/json/cves/2.0"

// Client NVD CVE API'sine sorgu yapar.
type Client struct {
	endpoint string
	http     *http.Client
	apiKey   string
	limiter  *rate.Limiter
}

// NewClient ortam değişkeninden API key okuyarak yeni Client döner.
// NVD_API_KEY varsa daha yüksek rate limit kullanılır.
func NewClient() *Client {
	key := os.Getenv("NVD_API_KEY")
	// Rate limit: key'siz 5/30sn (~6sn ara), key'li 50/30sn (~0.7sn ara).
	interval := 6 * time.Second
	if key != "" {
		interval = 700 * time.Millisecond
	}
	return &Client{
		endpoint: defaultEndpoint,
		http:     &http.Client{Timeout: 30 * time.Second},
		apiKey:   key,
		limiter:  rate.NewLimiter(rate.Every(interval), 1),
	}
}

// HasAPIKey rate limit ipucu göstermek için kullanılır.
func (c *Client) HasAPIKey() bool { return c.apiKey != "" }

// WithEndpoint test amaçlı endpoint override.
func (c *Client) WithEndpoint(u string) *Client {
	c.endpoint = u
	c.limiter = rate.NewLimiter(rate.Inf, 1) // test'te rate limit yok
	return c
}

// Query vuln.Client arayüzünü uygular. Sadece CPE identifier destekler.
func (c *Client) Query(ctx context.Context, id matcher.Identifier, version string) ([]vuln.Vulnerability, error) {
	if id.Type != matcher.IdentifierTypeCPE {
		return nil, nil // NVD CPE-tabanlı; purl desteklenmiyor
	}
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	body, err := c.fetch(ctx, id.Value)
	if err != nil {
		return nil, err
	}

	normVer, _ := matcher.NormalizeVersion(version)
	return parseNVD(body, id.Value, normVer)
}

func (c *Client) fetch(ctx context.Context, cpe string) ([]byte, error) {
	q := url.Values{}
	q.Set("virtualMatchString", cpe)
	reqURL := c.endpoint + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "mr-ducksafe/0.1")
	if c.apiKey != "" {
		req.Header.Set("apiKey", c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NVD status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// --- NVD JSON şeması (sadeleştirilmiş) ---

type nvdResponse struct {
	Vulnerabilities []struct {
		CVE nvdCVE `json:"cve"`
	} `json:"vulnerabilities"`
}

type nvdCVE struct {
	ID           string `json:"id"`
	Descriptions []struct {
		Lang  string `json:"lang"`
		Value string `json:"value"`
	} `json:"descriptions"`
	Metrics struct {
		V31 []nvdMetric `json:"cvssMetricV31"`
		V30 []nvdMetric `json:"cvssMetricV30"`
		V2  []nvdMetric `json:"cvssMetricV2"`
	} `json:"metrics"`
	Configurations []struct {
		Nodes []struct {
			CpeMatch []nvdCpeMatch `json:"cpeMatch"`
		} `json:"nodes"`
	} `json:"configurations"`
}

type nvdMetric struct {
	CvssData struct {
		BaseScore    float64 `json:"baseScore"`
		BaseSeverity string  `json:"baseSeverity"`
	} `json:"cvssData"`
	BaseSeverity string `json:"baseSeverity"`
}

type nvdCpeMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	Criteria              string `json:"criteria"`
	VersionStartIncluding string `json:"versionStartIncluding"`
	VersionStartExcluding string `json:"versionStartExcluding"`
	VersionEndIncluding   string `json:"versionEndIncluding"`
	VersionEndExcluding   string `json:"versionEndExcluding"`
}

func parseNVD(body []byte, ourCPE, version string) ([]vuln.Vulnerability, error) {
	var r nvdResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("NVD response parse: %w", err)
	}

	vendor, product := vendorProduct(ourCPE)
	var out []vuln.Vulnerability

	for _, item := range r.Vulnerabilities {
		cve := item.CVE
		fixed, affected := matchVersion(cve, vendor, product, version)
		if !affected {
			continue
		}
		out = append(out, vuln.Vulnerability{
			ID:            cve.ID,
			Summary:       englishDesc(cve),
			Severity:      severityOf(cve),
			FixedVersions: fixed,
		})
	}
	return out, nil
}

// matchVersion CVE'nin cpeMatch aralıklarından bizim version'ı kontrol eder.
// affected=true ise etkilendik; fixed varsa düzeltme sürümlerini döner.
func matchVersion(cve nvdCVE, vendor, product, version string) (fixed []string, affected bool) {
	seen := map[string]bool{}
	for _, conf := range cve.Configurations {
		for _, node := range conf.Nodes {
			for _, m := range node.CpeMatch {
				if !m.Vulnerable {
					continue
				}
				mv, mp := vendorProduct(m.Criteria)
				if mv != vendor || mp != product {
					continue
				}
				if versionInRange(version, m) {
					affected = true
					if m.VersionEndExcluding != "" && !seen[m.VersionEndExcluding] {
						fixed = append(fixed, m.VersionEndExcluding)
						seen[m.VersionEndExcluding] = true
					}
				}
			}
		}
	}
	return fixed, affected
}

func versionInRange(v string, m nvdCpeMatch) bool {
	if v == "" {
		return false
	}
	if m.VersionStartIncluding != "" && compareVersions(v, m.VersionStartIncluding) < 0 {
		return false
	}
	if m.VersionStartExcluding != "" && compareVersions(v, m.VersionStartExcluding) <= 0 {
		return false
	}
	if m.VersionEndIncluding != "" && compareVersions(v, m.VersionEndIncluding) > 0 {
		return false
	}
	if m.VersionEndExcluding != "" && compareVersions(v, m.VersionEndExcluding) >= 0 {
		return false
	}
	// Hiç bound yoksa: criteria'da spesifik version var mı?
	if m.VersionStartIncluding == "" && m.VersionStartExcluding == "" &&
		m.VersionEndIncluding == "" && m.VersionEndExcluding == "" {
		cpeVer := cpeVersion(m.Criteria)
		if cpeVer != "" && cpeVer != "*" && cpeVer != "-" {
			return compareVersions(v, cpeVer) == 0
		}
		// version "*" + bound yok → tüm sürümler etkilenir
		return true
	}
	return true
}

// compareVersions iki version string'ini sayısal segment bazında karşılaştırır.
// -1: a<b, 0: a==b, 1: a>b
func compareVersions(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		var av, bv int
		if i < len(as) {
			av, _ = strconv.Atoi(numericPrefix(as[i]))
		}
		if i < len(bs) {
			bv, _ = strconv.Atoi(numericPrefix(bs[i]))
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

// numericPrefix "6099abc" → "6099", "abc" → "".
func numericPrefix(s string) string {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return s[:i]
}

// vendorProduct cpe:2.3:a:vendor:product:... → (vendor, product).
func vendorProduct(cpe string) (string, string) {
	parts := strings.Split(cpe, ":")
	if len(parts) >= 5 {
		return parts[3], parts[4]
	}
	return "", ""
}

// cpeVersion cpe:2.3:a:vendor:product:VERSION:... → VERSION.
func cpeVersion(cpe string) string {
	parts := strings.Split(cpe, ":")
	if len(parts) >= 6 {
		return parts[5]
	}
	return ""
}

func englishDesc(cve nvdCVE) string {
	for _, d := range cve.Descriptions {
		if d.Lang == "en" {
			return d.Value
		}
	}
	if len(cve.Descriptions) > 0 {
		return cve.Descriptions[0].Value
	}
	return ""
}

func severityOf(cve nvdCVE) vuln.Severity {
	sev := ""
	switch {
	case len(cve.Metrics.V31) > 0:
		sev = pickSeverity(cve.Metrics.V31[0])
	case len(cve.Metrics.V30) > 0:
		sev = pickSeverity(cve.Metrics.V30[0])
	case len(cve.Metrics.V2) > 0:
		sev = pickSeverity(cve.Metrics.V2[0])
	}
	switch strings.ToUpper(sev) {
	case "CRITICAL":
		return vuln.SeverityCritical
	case "HIGH":
		return vuln.SeverityHigh
	case "MEDIUM":
		return vuln.SeverityMedium
	case "LOW":
		return vuln.SeverityLow
	default:
		return vuln.SeverityUnknown
	}
}

func pickSeverity(m nvdMetric) string {
	if m.CvssData.BaseSeverity != "" {
		return m.CvssData.BaseSeverity
	}
	return m.BaseSeverity
}
