package cutie

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// JSONHandlerOptions configures a JSONHandler.
type JSONHandlerOptions struct {
	Expand bool  // Multi-line output (default: false)
	Color  *bool // Colorize output (default: auto-detect TTY)
}

// JSONHandler writes structured JSON log records.
// Suitable for production environments (CloudWatch, ELK, OpenTelemetry).
type JSONHandler struct {
	expand bool
	color  bool
}

// NewJSONHandler creates a JSONHandler with the given options.
// Pass nil for defaults (compact, auto-detect color).
func NewJSONHandler(opts *JSONHandlerOptions) *JSONHandler {
	h := &JSONHandler{}
	if opts != nil {
		h.expand = opts.Expand
		if opts.Color != nil {
			h.color = *opts.Color
		} else {
			h.color = isTTY(os.Stdout)
		}
	} else {
		h.color = isTTY(os.Stdout)
	}
	return h
}

type kv struct {
	key string
	val any
}

func (h *JSONHandler) Handle(record LogRecord) {
	entries := []kv{
		{"level", string(record.Level)},
		{"msg", record.Msg},
		{"time", record.Time.Format("2006-01-02T15:04:05.000Z07:00")},
	}

	if len(record.Attrs) > 0 {
		keys := make([]string, 0, len(record.Attrs))
		for k := range record.Attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			entries = append(entries, kv{k, record.Attrs[k]})
		}
	}

	shouldExpand := h.expand
	if record.Expand != nil {
		shouldExpand = *record.Expand
	}

	var line string
	if h.color {
		if shouldExpand {
			line = colorizeExpanded(entries, record.Level)
		} else {
			line = colorizeCompact(entries, record.Level)
		}
	} else {
		if shouldExpand {
			line = plainExpanded(entries)
		} else {
			line = plainCompact(entries)
		}
	}

	fprintln(writerForLevel(record.Level), line)
}

// --- plain JSON (no color) ---

func jsonValue(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func plainCompact(entries []kv) string {
	var b strings.Builder
	b.WriteByte('{')
	for i, e := range entries {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%s:%s", jsonValue(e.key), jsonValue(e.val))
	}
	b.WriteByte('}')
	return b.String()
}

func plainExpanded(entries []kv) string {
	var b strings.Builder
	b.WriteByte('{')
	for i, e := range entries {
		b.WriteByte('\n')
		fmt.Fprintf(&b, "  %s: %s", jsonValue(e.key), jsonValue(e.val))
		if i < len(entries)-1 {
			b.WriteByte(',')
		}
	}
	b.WriteString("\n}")
	return b.String()
}

// --- colorized JSON ---

func colorizeValue(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`%s"%s"%s`, cCyan, v, cReset)
	case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return fmt.Sprintf("%s%v%s", cYellow, v, cReset)
	case bool:
		return fmt.Sprintf("%s%v%s", cYellow, v, cReset)
	case nil:
		return fmt.Sprintf("%snull%s", cDim, cReset)
	default:
		b, _ := json.Marshal(v)
		return fmt.Sprintf("%s%s%s", cWhite, string(b), cReset)
	}
}

func colorizeKeyValue(key string, value any, level Level) string {
	var coloredKey string
	if key == "level" {
		coloredKey = fmt.Sprintf(`%s"%s"%s`, cDim, key, cReset)
	} else {
		coloredKey = fmt.Sprintf(`%s"%s"%s`, cBlue, key, cReset)
	}

	var coloredValue string
	if key == "level" {
		coloredValue = fmt.Sprintf(`%s"%s"%s`, levelColor(level), value, cReset)
	} else {
		coloredValue = colorizeValue(value)
	}

	return fmt.Sprintf("%s%s:%s%s", coloredKey, cDim, cReset, coloredValue)
}

func colorizeCompact(entries []kv, level Level) string {
	parts := make([]string, len(entries))
	for i, e := range entries {
		parts[i] = colorizeKeyValue(e.key, e.val, level)
	}
	return fmt.Sprintf("%s{%s%s%s}%s", cDim, cReset, strings.Join(parts, cDim+","+cReset), cDim, cReset)
}

func colorizeExpanded(entries []kv, level Level) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%s{%s", cDim, cReset))
	for i, e := range entries {
		comma := ""
		if i < len(entries)-1 {
			comma = fmt.Sprintf("%s,%s", cDim, cReset)
		}
		lines = append(lines, fmt.Sprintf("  %s%s", colorizeKeyValue(e.key, e.val, level), comma))
	}
	lines = append(lines, fmt.Sprintf("%s}%s", cDim, cReset))
	return strings.Join(lines, "\n")
}
