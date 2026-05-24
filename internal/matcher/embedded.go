package matcher

import _ "embed"

//go:embed data/mappings.yaml
var embeddedMappings []byte

// EmbeddedMappings paketle gömülmüş varsayılan mapping verisini döner.
func EmbeddedMappings() []byte {
	return embeddedMappings
}
