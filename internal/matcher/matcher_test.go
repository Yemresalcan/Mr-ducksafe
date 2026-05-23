package matcher

import "testing"

func TestMatchConfidence_String(t *testing.T) {
	cases := []struct {
		c    MatchConfidence
		want string
	}{
		{ConfidenceHigh, "high"},
		{ConfidenceMedium, "medium"},
		{ConfidenceLow, "low"},
		{ConfidenceNone, "none"},
	}
	for _, tc := range cases {
		if got := tc.c.String(); got != tc.want {
			t.Errorf("MatchConfidence(%d).String() = %q, want %q", tc.c, got, tc.want)
		}
	}
}
