package cmd

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestAutoScanArgs(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		explorer bool
		want     []string
	}{
		{"double-click without args runs scan", nil, true, []string{"scan"}},
		{"terminal without args keeps help", nil, false, nil},
		{"double-click with args is left alone", []string{"history"}, true, nil},
		{"terminal with args is left alone", []string{"scan", "--details"}, false, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := autoScanArgs(c.args, c.explorer); !reflect.DeepEqual(got, c.want) {
				t.Errorf("autoScanArgs(%v, %v) = %v, want %v", c.args, c.explorer, got, c.want)
			}
		})
	}
}

func TestWaitForEnter(t *testing.T) {
	var out bytes.Buffer
	waitForEnter(strings.NewReader("\n"), &out)
	if !strings.Contains(out.String(), "Enter") {
		t.Errorf("prompt not shown: %q", out.String())
	}
	// EOF (no console attached) must not block or panic.
	waitForEnter(strings.NewReader(""), &out)
}
