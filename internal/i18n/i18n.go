// Package i18n holds every user-facing CLI string in English and Turkish.
//
// The active language is process-wide and is chosen once at start-up (see
// internal/cmd/lang.go); everything else just calls T / Tf.
package i18n

import (
	"fmt"
	"strings"
)

// Lang is a supported output language.
type Lang string

const (
	// EN is English (the default).
	EN Lang = "en"
	// TR is Turkish.
	TR Lang = "tr"
)

var current = EN

// Set changes the active language. Unknown values fall back to English.
func Set(l Lang) {
	if l == TR {
		current = TR
		return
	}
	current = EN
}

// Current returns the active language.
func Current() Lang { return current }

// Parse understands the spellings a user is likely to type.
func Parse(s string) (Lang, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "en", "eng", "english", "ingilizce", "1":
		return EN, true
	case "tr", "tur", "turkish", "türkçe", "turkce", "2":
		return TR, true
	}
	return EN, false
}

// T returns the message for key in the active language, falling back to
// English and finally to the key itself.
func T(key string) string {
	m, ok := messages[key]
	if !ok {
		return key
	}
	if s, ok := m[current]; ok {
		return s
	}
	return m[EN]
}

// Tf is T followed by fmt.Sprintf.
func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}
