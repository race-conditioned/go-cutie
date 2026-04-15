package cutie

import (
	"fmt"
	"sort"
	"strings"
)

const levelWidth = 5

func formatAttr(key string, value any) string {
	str := fmt.Sprintf("%v", value)
	if strings.Contains(str, " ") {
		str = `"` + str + `"`
	}
	return fmt.Sprintf("%s%s=%s%s", cDim, key, str, cReset)
}

// PrettyHandler writes colored, columnar log lines to the terminal.
// Suitable for local development.
//
// Output format:
//
//	INFO   server started    port=8080  stage=local
//	ERROR  db failed         err="connection refused"
type PrettyHandler struct{}

func (h *PrettyHandler) Handle(record LogRecord) {
	label := strings.ToUpper(string(record.Level))
	for len(label) < levelWidth {
		label += " "
	}

	coloredLevel := levelColor(record.Level) + label + cReset
	coloredMsg := cWhite + record.Msg + cReset

	attrStr := ""
	if len(record.Attrs) > 0 {
		keys := make([]string, 0, len(record.Attrs))
		for k := range record.Attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, formatAttr(k, record.Attrs[k]))
		}
		attrStr = "  " + strings.Join(parts, "  ")
	}

	line := fmt.Sprintf("  %s  %s%s", coloredLevel, coloredMsg, attrStr)
	fprintln(writerForLevel(record.Level), line)
}
