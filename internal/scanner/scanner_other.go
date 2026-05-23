//go:build !windows

package scanner

func newPlatformScanner() (Scanner, error) {
	return nil, ErrUnsupportedOS
}
