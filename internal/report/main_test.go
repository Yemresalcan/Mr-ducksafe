package report

import (
	"os"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/i18n"
)

// The existing assertions and golden files are written against the Turkish
// output; English output is covered by the *_EN tests.
func TestMain(m *testing.M) {
	i18n.Set(i18n.TR)
	os.Exit(m.Run())
}
