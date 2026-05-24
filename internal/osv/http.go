package osv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

const defaultEndpoint = "https://api.osv.dev/v1/query"

// HTTPClient OSV.dev'e gerçek HTTP sorgusu yapan Client implementasyonu.
type HTTPClient struct {
	endpoint string
	http     *http.Client
	maxRetry int
}

// NewHTTPClient varsayılan ayarlarla yeni HTTPClient oluşturur.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		endpoint: defaultEndpoint,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetry: 3,
	}
}

// WithEndpoint test amaçlı endpoint override.
func (c *HTTPClient) WithEndpoint(url string) *HTTPClient {
	c.endpoint = url
	return c
}

// Query Client arayüzünü uygular.
func (c *HTTPClient) Query(ctx context.Context, id matcher.Identifier, version string) ([]Vulnerability, error) {
	payload, err := buildPayload(id, version)
	if err != nil {
		return nil, err
	}
	body, err := c.doWithRetry(ctx, payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(body)
}

func buildPayload(id matcher.Identifier, version string) ([]byte, error) {
	req := map[string]any{
		"version": version,
	}
	switch id.Type {
	case matcher.IdentifierTypePURL:
		req["package"] = map[string]string{"purl": id.Value}
	case matcher.IdentifierTypeCPE:
		// OSV.dev CPE'yi doğrudan kabul etmez; CPE'den vendor:product:name çıkarıp
		// ecosystem hint ile package.name'e gönderiyoruz.
		name := nameFromCPE(id.Value)
		pkg := map[string]string{"name": name}
		if id.Ecosystem != "" {
			pkg["ecosystem"] = id.Ecosystem
		}
		req["package"] = pkg
	default:
		return nil, fmt.Errorf("desteklenmeyen identifier type: %s", id.Type)
	}
	return json.Marshal(req)
}

func nameFromCPE(cpe string) string {
	// cpe:2.3:a:vendor:product[:...] -> "product"
	parts := strings.Split(cpe, ":")
	if len(parts) >= 5 {
		return parts[4]
	}
	return cpe
}

func (c *HTTPClient) doWithRetry(ctx context.Context, payload []byte) ([]byte, error) {
	var lastErr error
	backoff := 500 * time.Millisecond

	for attempt := 0; attempt < c.maxRetry; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("request oluştur: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "mr-ducksafe/0.1")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			switch {
			case resp.StatusCode == http.StatusOK:
				return body, readErr
			case resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode >= 500:
				lastErr = fmt.Errorf("OSV status %d: %s", resp.StatusCode, string(body))
			default:
				return nil, fmt.Errorf("OSV status %d: %s", resp.StatusCode, string(body))
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	if lastErr == nil {
		lastErr = errors.New("OSV: max retry aşıldı")
	}
	return nil, lastErr
}

type osvResponse struct {
	Vulns []osvVuln `json:"vulns"`
}

type osvVuln struct {
	ID         string    `json:"id"`
	Summary    string    `json:"summary"`
	Published  time.Time `json:"published"`
	References []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"references"`
	Affected []struct {
		Ranges []struct {
			Events []map[string]string `json:"events"`
		} `json:"ranges"`
	} `json:"affected"`
	Severity []struct {
		Type  string `json:"type"`
		Score string `json:"score"`
	} `json:"severity"`
}

func parseResponse(body []byte) ([]Vulnerability, error) {
	if len(body) == 0 {
		return nil, nil
	}
	var r osvResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("OSV response parse: %w", err)
	}
	out := make([]Vulnerability, 0, len(r.Vulns))
	for _, v := range r.Vulns {
		out = append(out, Vulnerability{
			ID:            v.ID,
			Summary:       v.Summary,
			Severity:      severityFromCVSS(v.Severity),
			FixedVersions: collectFixed(v.Affected),
			References:    collectReferences(v.References),
			Published:     v.Published,
		})
	}
	return out, nil
}

func severityFromCVSS(items []struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}) Severity {
	for _, s := range items {
		if !strings.HasPrefix(s.Type, "CVSS") {
			continue
		}
		score := extractBaseScore(s.Score)
		switch {
		case score >= 9.0:
			return SeverityCritical
		case score >= 7.0:
			return SeverityHigh
		case score >= 4.0:
			return SeverityMedium
		case score > 0:
			return SeverityLow
		}
	}
	return SeverityUnknown
}

// extractBaseScore CVSS vector veya numeric score'dan base score çıkarır.
// MVP için kaba bir yaklaşım: vector parse etmek yerine HTTP impl'i sadece
// numeric score'u doğru okur. CVSS vector görünce kötümser olup High kabul
// ediyoruz (security için "false high" "false low"dan daha güvenli).
func extractBaseScore(s string) float64 {
	if strings.HasPrefix(s, "CVSS:") {
		return 7.5
	}
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func collectFixed(affected []struct {
	Ranges []struct {
		Events []map[string]string `json:"events"`
	} `json:"ranges"`
}) []string {
	var out []string
	seen := map[string]bool{}
	for _, a := range affected {
		for _, r := range a.Ranges {
			for _, ev := range r.Events {
				if v, ok := ev["fixed"]; ok && !seen[v] {
					out = append(out, v)
					seen[v] = true
				}
			}
		}
	}
	return out
}

func collectReferences(refs []struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.URL)
	}
	return out
}
