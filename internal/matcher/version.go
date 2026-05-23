package matcher

import (
	"regexp"
	"strings"
)

// VersionScheme bir version string'inin format şemasını belirtir.
type VersionScheme string

const (
	SchemeSemver  VersionScheme = "semver"
	SchemeFour    VersionScheme = "fourpart"
	SchemeCalver  VersionScheme = "calver"
	SchemeUnknown VersionScheme = "unknown"
)

var (
	semverRe   = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)([-+].+)?$`)
	fourPartRe = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+)$`)
	calverRe   = regexp.MustCompile(`^(20\d{2})\.(\d{1,2})\.(\d{1,2})$`)
)

// NormalizeVersion bir version string'ini OSV'ye yollamaya uygun
// normalize edilmiş haliyle ve tespit edilen şemayla döner.
func NormalizeVersion(raw string) (string, VersionScheme) {
	v := strings.TrimSpace(raw)
	v = strings.TrimPrefix(v, "v")

	// Calver önce (daha spesifik: yıl 20XX ile başlar) — yoksa
	// "2024.01.5" semver olarak match olur.
	if calverRe.MatchString(v) {
		return v, SchemeCalver
	}
	if m := fourPartRe.FindStringSubmatch(v); m != nil {
		// Dört parça (örn: 120.0.6099.130) — ilk üç parçayı semver olarak kullan
		return m[1] + "." + m[2] + "." + m[3], SchemeFour
	}
	if m := semverRe.FindStringSubmatch(v); m != nil {
		return m[1] + "." + m[2] + "." + m[3], SchemeSemver
	}
	return v, SchemeUnknown
}
