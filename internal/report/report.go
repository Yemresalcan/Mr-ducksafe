// Package report ScanResult'ı CLI çıktısı için formatlar.
package report

import (
	"fmt"
	"io"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

// Format çıktı formatı.
type Format string

const (
	FormatTable    Format = "table"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

// Options formatter davranışını yapılandırır.
type Options struct {
	ShowUnmatched bool
	NoColor       bool
}

// Write verilen format ile sonucu w'ye yazar.
func Write(w io.Writer, result *scanengine.ScanResult, format Format, opts Options) error {
	switch format {
	case FormatTable:
		return writeTable(w, result, opts)
	case FormatJSON:
		return writeJSON(w, result, opts)
	case FormatMarkdown:
		return writeMarkdown(w, result, opts)
	default:
		return fmt.Errorf("desteklenmeyen format: %s", format)
	}
}
