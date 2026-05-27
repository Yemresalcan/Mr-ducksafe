package report

import (
	"errors"
	"io"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

// writeMarkdown Task 17'de gerçek implementasyonla değiştirilecek.
func writeMarkdown(_ io.Writer, _ *scanengine.ScanResult, _ Options) error {
	return errors.New("markdown format henüz implement edilmedi (Task 17)")
}
