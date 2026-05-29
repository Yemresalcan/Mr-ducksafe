// Package vuln güvenlik açığı veri sağlayıcıları için tarafsız tipler
// ve arayüz tanımlar. OSV ve NVD gibi sağlayıcılar bu arayüzü uygular.
package vuln

import (
	"context"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

// Severity bir vulnerability'nin etki seviyesi.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityUnknown  Severity = "unknown"
)

// Vulnerability tek bir CVE/GHSA bulgusu.
type Vulnerability struct {
	ID            string
	Summary       string
	Severity      Severity
	FixedVersions []string
	References    []string
	Published     time.Time
}

// Client bir güvenlik açığı veri kaynağını sorgulayan arayüz.
type Client interface {
	Query(ctx context.Context, id matcher.Identifier, version string) ([]Vulnerability, error)
}
