package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// halloweenTitle is the media-title mpv reports for a real YouTube
// live stream; the "🧟\u200d♀\ufe0f" ZWJ sequence made the frame scroll.
const halloweenTitle = "Halloween lofi radio  🧟\u200d♀\ufe0f - spooky beats to get chills to"

// codepointWidth sums per-code-point widths, which is how terminals
// without grapheme-cluster support (VTE, xterm.js) lay out a line.
func codepointWidth(s string) int {
	w := 0
	for _, r := range s {
		w += lipgloss.Width(string(r))
	}
	return w
}

var sgrPattern = regexp.MustCompile("\x1b\\[[0-9;:]*[A-Za-z]")

func TestCleanText(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"plain", "Lofi Girl 24/7", "Lofi Girl 24/7"},
		{"zwj emoji", halloweenTitle, "Halloween lofi radio  🧟♀ - spooky beats to get chills to"},
		{"skin tone", "chill 👍🏽 beats", "chill 👍 beats"},
		{"variation selector", "made with ❤\ufe0f", "made with ❤"},
		{"keycap", "top 1\ufe0f\u20e3", "top 1"},
		{"zwj flag", "pride 🏳\ufe0f\u200d🌈", "pride 🏳🌈"},
		{"control chars", "line one\nline two\ttab\r", "line one line two tab"},
		{"accents kept", "Café del Mar", "Café del Mar"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cleanText(tc.in)
			if got != tc.want {
				t.Fatalf("cleanText(%q) = %q, want %q", tc.in, got, tc.want)
			}
			if lw, cw := lipgloss.Width(got), codepointWidth(got); lw != cw {
				t.Errorf("width mismatch for %q: lipgloss %d, per-code-point %d", got, lw, cw)
			}
		})
	}
}

// TestView_EmojiTitleDoesNotOverflow guards the "frame keeps scrolling
// up" bug: no rendered line may exceed the terminal width when measured
// the way a terminal without grapheme clustering would.
func TestView_EmojiTitleDoesNotOverflow(t *testing.T) {
	m := fixture()
	m.cfg.Stations[0].Name = "spooky 🧟\u200d♀\ufe0f"
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(Model)
	m.playingIdx, m.playing = 0, true
	updated, _ = m.Update(MetadataChangedMsg{Title: halloweenTitle})
	m = updated.(Model)

	for i, line := range strings.Split(m.View(), "\n") {
		if w := codepointWidth(sgrPattern.ReplaceAllString(line, "")); w > m.width {
			t.Errorf("line %d is %d cells wide in a per-code-point terminal (width %d):\n%s", i, w, m.width, line)
		}
	}
}
