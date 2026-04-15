package cutie

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const minValueWidth = 26

func stripAnsi(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && ((s[j] >= '0' && s[j] <= '9') || s[j] == ';') {
				j++
			}
			if j < len(s) && s[j] == 'm' {
				i = j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func computeDimensions(groups [][][2]string, title string) (keyCol, innerWidth int) {
	maxKeyLen := 16
	for _, group := range groups {
		for _, pair := range group {
			if l := utf8.RuneCountInString(pair[0]); l > maxKeyLen {
				maxKeyLen = l
			}
		}
	}
	keyCol = maxKeyLen + 2
	innerWidth = 2 + keyCol + minValueWidth
	if titleWidth := utf8.RuneCountInString(title) + 4; titleWidth > innerWidth {
		innerWidth = titleWidth
	}
	return
}

func bannerRow(key, value string, keyCol, innerWidth int) string {
	valueWidth := innerWidth - 2 - keyCol
	keyPad := strings.Repeat(" ", max(0, keyCol-utf8.RuneCountInString(key)))

	display := value
	if utf8.RuneCountInString(display) > valueWidth {
		display = string([]rune(display)[:valueWidth-1]) + "…"
	}
	rightPad := strings.Repeat(" ", max(0, valueWidth-utf8.RuneCountInString(display)))

	return fmt.Sprintf("%s│%s  %s%s%s%s%s%s%s%s%s│%s",
		cCyan, cReset,
		cBlue, key, cReset, keyPad,
		cWhite, display, cReset, rightPad,
		cCyan, cReset,
	)
}

func printBanner(title string, groups [][][2]string) {
	keyCol, innerWidth := computeDimensions(groups, title)
	line := strings.Repeat("─", innerWidth)
	divider := fmt.Sprintf("%s├%s┤%s", cCyan, line, cReset)

	titleLen := utf8.RuneCountInString(title)
	titlePad := (innerWidth - titleLen) / 2
	titleRow := fmt.Sprintf("%s│%s%s%s%s%s%s%s%s│%s",
		cCyan, cReset,
		strings.Repeat(" ", titlePad),
		cBold, cWhite, title, cReset,
		strings.Repeat(" ", innerWidth-titlePad-titleLen),
		cCyan, cReset,
	)

	w := stdout
	fprintln(w, "")
	fprintln(w, fmt.Sprintf("%s╭%s╮%s", cCyan, line, cReset))
	fprintln(w, titleRow)
	fprintln(w, divider)
	for i, group := range groups {
		for _, pair := range group {
			fprintln(w, bannerRow(pair[0], pair[1], keyCol, innerWidth))
		}
		if i < len(groups)-1 {
			fprintln(w, divider)
		}
	}
	fprintln(w, fmt.Sprintf("%s╰%s╯%s", cCyan, line, cReset))
	fprintln(w, "")
}

// PrintBanner renders a styled startup box to stdout showing all config keys.
func PrintBanner(title string, config Attrs) {
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	group := make([][2]string, len(keys))
	for i, k := range keys {
		group[i] = [2]string{k, fmt.Sprintf("%v", config[k])}
	}
	printBanner(title, [][][2]string{group})
}

// PrintBannerPick renders a styled startup box showing only the picked keys (no dividers).
func PrintBannerPick(title string, config Attrs, keys []string) {
	group := make([][2]string, len(keys))
	for i, k := range keys {
		group[i] = [2]string{k, fmt.Sprintf("%v", config[k])}
	}
	printBanner(title, [][][2]string{group})
}

// PrintBannerGrouped renders a styled startup box with grouped keys and dividers between groups.
func PrintBannerGrouped(title string, config Attrs, groups [][]string) {
	gs := make([][][2]string, len(groups))
	for i, keys := range groups {
		g := make([][2]string, len(keys))
		for j, k := range keys {
			g[j] = [2]string{k, fmt.Sprintf("%v", config[k])}
		}
		gs[i] = g
	}
	printBanner(title, gs)
}

// PrintListening writes a styled "server is up" line to stdout.
func PrintListening(addr string, handlers ...int) {
	suffix := ""
	if len(handlers) > 0 {
		suffix = fmt.Sprintf("  %s·  %d handlers%s", cDim, handlers[0], cReset)
	}
	fprintln(stdout, fmt.Sprintf("  %s▶%s  %s%s%s%s%s", cCyan, cReset, cBold, cWhite, addr, cReset, suffix))
	fprintln(stdout, "")
}
