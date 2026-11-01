// Package config kullanıcı config dosyası ve default path'leri yönetir.
package config

import (
	"os"
	"path/filepath"
)

// DataDir mr-ducksafe'in kalıcı verilerini sakladığı klasör (~/.mr-ducksafe).
func DataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".mr-ducksafe")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	return dir, nil
}

// DBPath default SQLite DB yolu.
func DBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mr-ducksafe.db"), nil
}
