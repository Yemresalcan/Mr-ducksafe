package matcher

import "testing"

func TestNormalizeVersion(t *testing.T) {
	cases := []struct {
		raw        string
		wantValue  string
		wantScheme VersionScheme
	}{
		{"1.2.3", "1.2.3", SchemeSemver},
		{"v1.2.3", "1.2.3", SchemeSemver},
		{"1.2.3-beta", "1.2.3", SchemeSemver},
		{"120.0.6099.130", "120.0.6099", SchemeFour},
		{"2024.01.5", "2024.01.5", SchemeCalver},
		{"21H2", "21H2", SchemeUnknown},
		{"  1.0.0  ", "1.0.0", SchemeSemver},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			gotValue, gotScheme := NormalizeVersion(tc.raw)
			if gotValue != tc.wantValue {
				t.Errorf("value = %q, want %q", gotValue, tc.wantValue)
			}
			if gotScheme != tc.wantScheme {
				t.Errorf("scheme = %q, want %q", gotScheme, tc.wantScheme)
			}
		})
	}
}
