package store

import "testing"

func TestScanStatus_Values(t *testing.T) {
	all := []ScanStatus{StatusRunning, StatusSuccess, StatusFailed, StatusInterrupted}
	for _, s := range all {
		if s == "" {
			t.Errorf("status boş olamaz")
		}
	}
}
