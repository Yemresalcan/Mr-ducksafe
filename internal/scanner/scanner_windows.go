//go:build windows

package scanner

import (
	"context"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const sourceWindowsRegistry = "windows-registry"

var uninstallKeys = []struct {
	root registry.Key
	path string
}{
	{registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Uninstall`},
	{registry.LOCAL_MACHINE, `Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
	{registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall`},
}

type windowsScanner struct{}

func newPlatformScanner() (Scanner, error) {
	return &windowsScanner{}, nil
}

func (s *windowsScanner) Scan(ctx context.Context) ([]InstalledApp, error) {
	var apps []InstalledApp
	for _, k := range uninstallKeys {
		if err := ctx.Err(); err != nil {
			return apps, err
		}
		batch, err := readUninstallKey(k.root, k.path)
		if err != nil {
			// kısmi başarısızlığı tolere et — bir hive'ı okuyamamak fatal değil
			continue
		}
		apps = append(apps, batch...)
	}
	return apps, nil
}

func readUninstallKey(root registry.Key, path string) ([]InstalledApp, error) {
	parent, err := registry.OpenKey(root, path, registry.READ)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer parent.Close()

	subKeys, err := parent.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("list subkeys %s: %w", path, err)
	}

	apps := make([]InstalledApp, 0, len(subKeys))
	for _, name := range subKeys {
		app, ok := readSubKey(parent, name)
		if ok {
			apps = append(apps, app)
		}
	}
	return apps, nil
}

func readSubKey(parent registry.Key, name string) (InstalledApp, bool) {
	sub, err := registry.OpenKey(parent, name, registry.READ)
	if err != nil {
		return InstalledApp{}, false
	}
	defer sub.Close()

	display, _, err := sub.GetStringValue("DisplayName")
	if err != nil || display == "" {
		return InstalledApp{}, false
	}
	version, _, _ := sub.GetStringValue("DisplayVersion")
	publisher, _, _ := sub.GetStringValue("Publisher")

	return InstalledApp{
		Name:      display,
		Version:   version,
		Publisher: publisher,
		Source:    sourceWindowsRegistry,
	}, true
}
