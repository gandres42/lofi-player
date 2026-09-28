package tui

import (
	"strings"
	"unicode"
)

// cleanText makes externally sourced text (stream metadata, station
// names) safe to lay out in the frame.
//
// lipgloss measures width per grapheme cluster, so a ZWJ emoji like
// "🧟\u200d♀\ufe0f" (zombie + ZWJ + female sign + VS16) counts as 2
// cells. Many terminals (VTE, xterm.js) draw each code point
// separately and use 3–4. A line that renders wider than lipgloss
// computed wraps, and Bubble Tea's renderer — which moves the cursor
// up by the number of lines it thinks it drew — then re-prints the
// frame one row lower on every tick. Dropping the invisible joiners
// and modifiers leaves only code points whose width both sides agree
// on; the emoji degrade to their base glyphs ("🧟♀") rather than
// disappearing.
//
// Control characters (newlines, tabs, etc.) become spaces so metadata
// can't add rows to the layout either.
func cleanText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\u200d', // zero-width joiner
			r >= '\ufe00' && r <= '\ufe0f', // variation selectors
			r >= 0x1f3fb && r <= 0x1f3ff,   // skin-tone modifiers
			r >= 0xe0020 && r <= 0xe007f,   // tag characters (subdivision flags)
			r == '\u20e3':                  // combining enclosing keycap
			continue
		case unicode.IsControl(r):
			b.WriteByte(' ')
		default:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
