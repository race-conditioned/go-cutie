package cutie

import (
	"fmt"
	"time"
)

// AccessRecord holds the data for an HTTP access log line.
type AccessRecord struct {
	Method   string
	Path     string
	Status   int
	Duration time.Duration
}

func methodColor(method string) string {
	switch method {
	case "GET":
		return cCyan
	case "POST":
		return cBlue
	case "PUT":
		return cYellow
	case "DELETE":
		return cMagenta
	default:
		return cWhite
	}
}

func statusColor(status int) string {
	if status >= 500 {
		return cRed
	}
	if status >= 400 {
		return cYellow
	}
	if status >= 300 {
		return cYellow
	}
	return cCyan
}

// PrintAccess writes a colored HTTP access log line to stdout.
func PrintAccess(record AccessRecord) {
	method := record.Method
	for len(method) < 7 {
		method += " "
	}
	m := fmt.Sprintf("%s%s%s", methodColor(record.Method), method, cReset)
	p := fmt.Sprintf("%s%s%s", cWhite, record.Path, cReset)
	s := fmt.Sprintf("%s%s%d%s", statusColor(record.Status), cBold, record.Status, cReset)

	ms := record.Duration.Milliseconds()
	dur := fmt.Sprintf("%s%dms%s", cDim, ms, cReset)

	fprintln(stdout, fmt.Sprintf("  %s→%s  %s  %s  %s  %s", cCyan, cReset, m, p, s, dur))
}
