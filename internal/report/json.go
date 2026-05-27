package report

import (
	"errors"
	"io"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
)

// writeJSON Task 17'de gerçek implementasyonla değiştirilecek.
func writeJSON(_ io.Writer, _ *scanengine.ScanResult, _ Options) error {
	return errors.New("json format henüz implement edilmedi (Task 17)")
}
