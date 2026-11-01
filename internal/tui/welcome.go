package tui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// glyphs is a tiny 5×5 bitmap font, just enough letters for the wordmark.
// '#' is a filled pixel; rows are top to bottom.
var glyphs = map[rune][5]string{
	'M': {"#...#", "##.##", "#.#.#", "#...#", "#...#"},
	'R': {"####.", "#...#", "####.", "#..#.", "#...#"},
	'.': {"..", "..", "..", "..", "#."},
	'D': {"####.", "#...#", "#...#", "#...#", "####."},
	'U': {"#...#", "#...#", "#...#", "#...#", ".###."},
	'C': {".####", "#....", "#....", "#....", ".####"},
	'K': {"#...#", "#..#.", "###..", "#..#.", "#...#"},
	'S': {".####", "#....", ".###.", "....#", "####."},
	'A': {".###.", "#...#", "#####", "#...#", "#...#"},
	'F': {"#####", "#....", "####.", "#....", "#...."},
	'E': {"#####", "#....", "####.", "#....", "#####"},
}

// Wordmark renders text with the bitmap font using half-block characters,
// so each 5-pixel-high letter takes three terminal rows and looks square.
func Wordmark(text string) []string {
	// assemble the 5 pixel rows (plus an empty 6th so rows pair up)
	var px [6]strings.Builder
	for i, r := range strings.ToUpper(text) {
		g, ok := glyphs[r]
		if !ok {
			g = [5]string{"...", "...", "...", "...", "..."}
		}
		for row := 0; row < 6; row++ {
			if i > 0 {
				px[row].WriteByte('.')
			}
			if row < 5 {
				px[row].WriteString(g[row])
			} else {
				px[row].WriteString(strings.Repeat(".", len(g[0])))
			}
		}
	}
	lines := make([]string, 0, 3)
	for row := 0; row < 6; row += 2 {
		top, bot := px[row].String(), px[row+1].String()
		var b strings.Builder
		for c := 0; c < len(top); c++ {
			switch {
			case top[c] == '#' && bot[c] == '#':
				b.WriteRune('█')
			case top[c] == '#':
				b.WriteRune('▀')
			case bot[c] == '#':
				b.WriteRune('▄')
			default:
				b.WriteByte(' ')
			}
		}
		lines = append(lines, strings.TrimRight(b.String(), " "))
	}
	return lines
}

// Welcome is the first-run splash: the MR.DUCKSAFE wordmark, a tagline and
// the version. The language question is printed right below it.
func Welcome(version, tagline string) string {
	duck := color.New(color.FgYellow, color.Bold)
	var b strings.Builder
	b.WriteString("\n")
	for _, l := range Wordmark("MR.DUCKSAFE") {
		b.WriteString("  " + duck.Sprint(l) + "\n")
	}
	fmt.Fprintf(&b, "\n  %s  %s\n\n",
		color.New(color.Faint).Sprint(tagline),
		color.New(color.Faint).Sprint("· "+version))
	return b.String()
}

// Question formats a CLI-style question with numbered options; the first
// option is the default (taken on a bare Enter).
func Question(title string, options []string, defaultHint string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  %s %s\n", color.New(color.FgCyan, color.Bold).Sprint("?"), color.New(color.Bold).Sprint(title))
	for i, o := range options {
		hint := ""
		if i == 0 && defaultHint != "" {
			hint = " " + color.New(color.Faint).Sprint(defaultHint)
		}
		fmt.Fprintf(&b, "    %s %s%s\n", color.New(color.FgYellow).Sprintf("%d)", i+1), o, hint)
	}
	b.WriteString("  " + color.New(color.FgCyan, color.Bold).Sprint("›") + " ")
	return b.String()
}
