package scanner

import "testing"

func TestInstalledApp_ZeroValue(t *testing.T) {
	var app InstalledApp
	if app.Name != "" || app.Version != "" {
		t.Errorf("zero value bekleniyor, alındı: %+v", app)
	}
}

func TestErrUnsupportedOS(t *testing.T) {
	if ErrUnsupportedOS == nil {
		t.Fatal("ErrUnsupportedOS nil olmamalı")
	}
	if ErrUnsupportedOS.Error() == "" {
		t.Error("error mesajı boş olmamalı")
	}
}
