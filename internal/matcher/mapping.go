package matcher

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"gopkg.in/yaml.v3"
)

// mappingEntry mappings.yaml içindeki tek bir kayıt.
type mappingEntry struct {
	RegistryName        string `yaml:"registry_name"`
	RegistryNamePattern string `yaml:"registry_name_pattern"`
	Publisher           string `yaml:"publisher"`
	Identifier          struct {
		Type  string `yaml:"type"`
		Value string `yaml:"value"`
	} `yaml:"identifier"`
	EcosystemHint    string `yaml:"ecosystem_hint"`
	VersionNormalize string `yaml:"version_normalize"`

	pattern *regexp.Regexp // pattern compiled cache
}

type mappingFile struct {
	Mappings []mappingEntry `yaml:"mappings"`
}

// MappingMatcher YAML mapping tabanlı Matcher implementasyonu.
type MappingMatcher struct {
	entries []mappingEntry
}

// NewMappingMatcher verilen YAML byte slice'ından MappingMatcher oluşturur.
func NewMappingMatcher(yamlData []byte) (*MappingMatcher, error) {
	var f mappingFile
	if err := yaml.Unmarshal(yamlData, &f); err != nil {
		return nil, fmt.Errorf("mappings YAML parse: %w", err)
	}

	for i := range f.Mappings {
		if pat := f.Mappings[i].RegistryNamePattern; pat != "" {
			re, err := regexp.Compile(pat)
			if err != nil {
				return nil, fmt.Errorf("mapping pattern %q derlenemedi: %w", pat, err)
			}
			f.Mappings[i].pattern = re
		}
	}
	return &MappingMatcher{entries: f.Mappings}, nil
}

// NewDefaultMatcher gömülü mapping verisiyle Matcher döner.
func NewDefaultMatcher() (*MappingMatcher, error) {
	return NewMappingMatcher(EmbeddedMappings())
}

// Match scanner.Matcher arayüzünü uygular.
func (m *MappingMatcher) Match(_ context.Context, app scanner.InstalledApp) (Identifier, MatchConfidence, error) {
	for _, e := range m.entries {
		if e.matches(app) {
			return Identifier{
				Type:      IdentifierType(e.Identifier.Type),
				Value:     e.Identifier.Value,
				Ecosystem: e.EcosystemHint,
			}, ConfidenceHigh, nil
		}
	}
	return Identifier{}, ConfidenceNone, nil
}

func (e mappingEntry) matches(app scanner.InstalledApp) bool {
	switch {
	case e.RegistryName != "":
		if !strings.EqualFold(e.RegistryName, app.Name) {
			return false
		}
	case e.pattern != nil:
		if !e.pattern.MatchString(app.Name) {
			return false
		}
	default:
		return false
	}
	return true
}
