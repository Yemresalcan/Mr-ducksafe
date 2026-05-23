// Package scanner kurulu sistem yazılımlarının listesini çıkaran
// OS-spesifik scanner'ları tanımlar.
package scanner

import (
	"context"
	"errors"
)

// InstalledApp bir kurulu yazılımın tespit edilen meta verilerini taşır.
type InstalledApp struct {
	Name      string
	Version   string
	Publisher string
	Source    string
}

// Scanner OS-spesifik kurulu yazılım listeleme arayüzü.
type Scanner interface {
	Scan(ctx context.Context) ([]InstalledApp, error)
}

// ErrUnsupportedOS runtime OS desteklenmiyorsa döner.
var ErrUnsupportedOS = errors.New("scanner: bu işletim sistemi henüz desteklenmiyor")

// New runtime OS'una göre uygun Scanner implementasyonunu döner.
func New() (Scanner, error) {
	return newPlatformScanner()
}
