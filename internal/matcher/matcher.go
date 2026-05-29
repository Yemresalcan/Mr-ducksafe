// Package matcher InstalledApp'i sorgulanabilir CVE identifier'ına
// (CPE veya purl) eşleştirir.
package matcher

import (
	"context"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
)

// IdentifierType bir Identifier'ın hangi şemada olduğunu belirtir.
type IdentifierType string

const (
	IdentifierTypePURL IdentifierType = "purl"
	IdentifierTypeCPE  IdentifierType = "cpe"
)

// Identifier sorgulanabilir bir tekil yazılım koordinatı.
type Identifier struct {
	Type      IdentifierType
	Value     string
	Ecosystem string // OSV ecosystem hint, boş olabilir
	Version   string // sorguda kullanılacak çözülmüş versiyon (isimden veya app.Version)
	WingetID  string // remediation: winget paket id'si, boş olabilir
	UpdateURL string // remediation: manuel güncelleme linki, boş olabilir
}

// MatchConfidence eşleştirme güven seviyesi.
type MatchConfidence int

const (
	ConfidenceNone MatchConfidence = iota
	ConfidenceLow
	ConfidenceMedium
	ConfidenceHigh
)

func (c MatchConfidence) String() string {
	switch c {
	case ConfidenceHigh:
		return "high"
	case ConfidenceMedium:
		return "medium"
	case ConfidenceLow:
		return "low"
	default:
		return "none"
	}
}

// Matcher InstalledApp -> Identifier eşleştirme arayüzü.
type Matcher interface {
	Match(ctx context.Context, app scanner.InstalledApp) (Identifier, MatchConfidence, error)
}
