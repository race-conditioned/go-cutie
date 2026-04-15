package cutie

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

// --- helpers ---

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func strip(s string) string { return ansiRe.ReplaceAllString(s, "") }

func captureOutput(fn func()) (out, err string) {
	var outBuf, errBuf bytes.Buffer
	oldOut, oldErr := stdout, stderr
	stdout, stderr = &outBuf, &errBuf
	defer func() { stdout, stderr = oldOut, oldErr }()
	fn()
	return outBuf.String(), errBuf.String()
}

type recordingHandler struct {
	records []LogRecord
}

func (h *recordingHandler) Handle(r LogRecord) {
	h.records = append(h.records, r)
}

// --- Logger tests ---

func TestLoggerDispatchesRecord(t *testing.T) {
	h := &recordingHandler{}
	log := New(h)
	log.Info("hello", Attrs{"port": 8080})

	if len(h.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(h.records))
	}
	r := h.records[0]
	if r.Level != LevelInfo {
		t.Errorf("level = %q, want %q", r.Level, LevelInfo)
	}
	if r.Msg != "hello" {
		t.Errorf("msg = %q, want %q", r.Msg, "hello")
	}
	if r.Attrs["port"] != 8080 {
		t.Errorf("attrs[port] = %v, want 8080", r.Attrs["port"])
	}
}

func TestLoggerAllLevels(t *testing.T) {
	h := &recordingHandler{}
	log := New(h)
	log.Debug("d")
	log.Info("i")
	log.Warn("w")
	log.Error("e")

	if len(h.records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(h.records))
	}
	want := []Level{LevelDebug, LevelInfo, LevelWarn, LevelError}
	for i, w := range want {
		if h.records[i].Level != w {
			t.Errorf("record[%d].Level = %q, want %q", i, h.records[i].Level, w)
		}
	}
}

func TestLoggerWithMergesAttrs(t *testing.T) {
	h := &recordingHandler{}
	log := New(h).With(Attrs{"service": "billing"})
	log.Info("charge", Attrs{"amount": 42})

	r := h.records[0]
	if r.Attrs["service"] != "billing" {
		t.Errorf("missing base attr service")
	}
	if r.Attrs["amount"] != 42 {
		t.Errorf("missing call attr amount")
	}
}

func TestLoggerWithDoesNotMutateParent(t *testing.T) {
	h := &recordingHandler{}
	parent := New(h).With(Attrs{"a": 1})
	_ = parent.With(Attrs{"b": 2})
	parent.Info("test")

	r := h.records[0]
	if _, ok := r.Attrs["b"]; ok {
		t.Error("parent was mutated by With()")
	}
}

func TestLoggerExpandedCompact(t *testing.T) {
	h := &recordingHandler{}
	log := New(h)

	log.Expanded().Info("exp")
	log.Compact().Info("cmp")
	log.Info("default")

	if h.records[0].Expand == nil || !*h.records[0].Expand {
		t.Error("Expanded() should set Expand=true")
	}
	if h.records[1].Expand == nil || *h.records[1].Expand {
		t.Error("Compact() should set Expand=false")
	}
	if h.records[2].Expand != nil {
		t.Error("default should have Expand=nil")
	}
}

// --- PrettyHandler tests ---

func TestPrettyHandlerOutput(t *testing.T) {
	log := New(&PrettyHandler{})
	out, _ := captureOutput(func() {
		log.Info("server started", Attrs{"port": 8080, "stage": "local"})
	})
	plain := strip(out)
	if !strings.Contains(plain, "INFO") {
		t.Error("missing INFO level")
	}
	if !strings.Contains(plain, "server started") {
		t.Error("missing message")
	}
	if !strings.Contains(plain, "port=8080") {
		t.Error("missing port attr")
	}
	if !strings.Contains(plain, "stage=local") {
		t.Error("missing stage attr")
	}
}

func TestPrettyHandlerQuotesSpaces(t *testing.T) {
	log := New(&PrettyHandler{})
	out, _ := captureOutput(func() {
		log.Info("test", Attrs{"err": "connection refused"})
	})
	plain := strip(out)
	if !strings.Contains(plain, `err="connection refused"`) {
		t.Errorf("expected quoted value, got: %s", plain)
	}
}

func TestPrettyHandlerStderr(t *testing.T) {
	log := New(&PrettyHandler{})
	out, errOut := captureOutput(func() {
		log.Warn("warning")
		log.Error("failure")
	})
	if out != "" {
		t.Error("warn/error should not go to stdout")
	}
	plain := strip(errOut)
	if !strings.Contains(plain, "WARN") {
		t.Error("missing WARN on stderr")
	}
	if !strings.Contains(plain, "ERROR") {
		t.Error("missing ERROR on stderr")
	}
}

// --- JSONHandler tests ---

func TestJSONHandlerCompact(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Color: ptrBool(false)}))
	out, _ := captureOutput(func() {
		log.Info("hello", Attrs{"port": 8080})
	})
	out = strings.TrimSpace(out)
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if m["level"] != "info" {
		t.Errorf("level = %v", m["level"])
	}
	if m["msg"] != "hello" {
		t.Errorf("msg = %v", m["msg"])
	}
	if m["port"] != float64(8080) {
		t.Errorf("port = %v", m["port"])
	}
}

func TestJSONHandlerExpanded(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Expand: true, Color: ptrBool(false)}))
	out, _ := captureOutput(func() {
		log.Info("hello")
	})
	if !strings.Contains(out, "\n") {
		t.Error("expanded output should be multi-line")
	}
	// Should still be valid JSON
	out = strings.TrimSpace(out)
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
}

func TestJSONHandlerStderr(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Color: ptrBool(false)}))
	out, errOut := captureOutput(func() {
		log.Warn("w")
		log.Error("e")
	})
	if out != "" {
		t.Error("warn/error should not go to stdout")
	}
	if !strings.Contains(errOut, `"warn"`) {
		t.Error("missing warn on stderr")
	}
	if !strings.Contains(errOut, `"error"`) {
		t.Error("missing error on stderr")
	}
}

func TestJSONHandlerColorFalse(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Color: ptrBool(false)}))
	out, _ := captureOutput(func() {
		log.Info("test")
	})
	if strings.Contains(out, "\x1b[") {
		t.Error("color=false should produce no ANSI escapes")
	}
}

func TestJSONHandlerPerCallExpand(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Color: ptrBool(false)}))

	out1, _ := captureOutput(func() {
		log.Expanded().Info("expanded")
	})
	if !strings.Contains(out1, "\n  ") {
		t.Error("Expanded() should produce multi-line")
	}

	out2, _ := captureOutput(func() {
		log.Compact().Info("compact")
	})
	lines := strings.Split(strings.TrimSpace(out2), "\n")
	if len(lines) != 1 {
		t.Error("Compact() should produce single line")
	}
}

func TestJSONHandlerWithAttrs(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Color: ptrBool(false)})).
		With(Attrs{"service": "api"})
	out, _ := captureOutput(func() {
		log.Info("req", Attrs{"path": "/users"})
	})
	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatal(err)
	}
	if m["service"] != "api" {
		t.Error("missing With attr")
	}
	if m["path"] != "/users" {
		t.Error("missing call attr")
	}
}

// --- Banner tests ---

func TestPrintBannerAllKeys(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintBanner("test-app", Attrs{"port": 8080, "stage": "local"})
	})
	plain := strip(out)
	if !strings.Contains(plain, "test-app") {
		t.Error("missing title")
	}
	if !strings.Contains(plain, "port") {
		t.Error("missing port key")
	}
	if !strings.Contains(plain, "8080") {
		t.Error("missing port value")
	}
	if !strings.Contains(plain, "╭") || !strings.Contains(plain, "╰") {
		t.Error("missing box borders")
	}
}

func TestPrintBannerPick(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintBannerPick("app", Attrs{"a": 1, "b": 2, "c": 3}, []string{"a", "c"})
	})
	plain := strip(out)
	if !strings.Contains(plain, "1") || !strings.Contains(plain, "3") {
		t.Error("missing picked values")
	}
}

func TestPrintBannerPickOrder(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintBannerPick("app", Attrs{"z": "last", "a": "first"}, []string{"z", "a"})
	})
	plain := strip(out)
	zIdx := strings.Index(plain, "last")
	aIdx := strings.Index(plain, "first")
	if zIdx > aIdx {
		t.Error("pick order should be preserved (z before a)")
	}
}

func TestPrintBannerGrouped(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintBannerGrouped("app", Attrs{"a": 1, "b": 2, "c": 3}, [][]string{{"a"}, {"b", "c"}})
	})
	plain := strip(out)
	// Should have a divider between groups (├...┤ appears more than just header)
	count := strings.Count(plain, "├")
	if count < 2 {
		t.Errorf("expected at least 2 dividers (header + group), got %d", count)
	}
}

func TestPrintBannerAlignment(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintBannerPick("app", Attrs{"a": 1, "longkey": 2}, []string{"a", "longkey"})
	})
	plain := strip(out)
	lines := strings.Split(plain, "\n")
	// Find rows with values and check the value column starts at the same position
	var valueStarts []int
	for _, l := range lines {
		if strings.Contains(l, "│") && (strings.Contains(l, "1") || strings.Contains(l, "2")) {
			// Find position of the value
			idx := strings.Index(l, "1")
			if idx < 0 {
				idx = strings.Index(l, "2")
			}
			if idx > 0 {
				valueStarts = append(valueStarts, idx)
			}
		}
	}
	if len(valueStarts) == 2 && valueStarts[0] != valueStarts[1] {
		t.Errorf("values not aligned: positions %v", valueStarts)
	}
}

// --- Access tests ---

func TestPrintAccessFields(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintAccess(AccessRecord{
			Method:   "GET",
			Path:     "/users",
			Status:   200,
			Duration: 2 * time.Millisecond,
		})
	})
	plain := strip(out)
	if !strings.Contains(plain, "GET") {
		t.Error("missing method")
	}
	if !strings.Contains(plain, "/users") {
		t.Error("missing path")
	}
	if !strings.Contains(plain, "200") {
		t.Error("missing status")
	}
	if !strings.Contains(plain, "2ms") {
		t.Error("missing duration")
	}
}

func TestPrintAccessZeroDuration(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintAccess(AccessRecord{
			Method: "DELETE",
			Path:   "/x",
			Status: 204,
		})
	})
	plain := strip(out)
	if !strings.Contains(plain, "0ms") {
		t.Error("missing 0ms for zero duration")
	}
}

// --- Listening tests ---

func TestPrintListeningAddr(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintListening("http://localhost:8080")
	})
	plain := strip(out)
	if !strings.Contains(plain, "http://localhost:8080") {
		t.Error("missing address")
	}
	if !strings.Contains(plain, "▶") {
		t.Error("missing arrow")
	}
}

func TestPrintListeningHandlers(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintListening("http://localhost:8080", 11)
	})
	plain := strip(out)
	if !strings.Contains(plain, "11 handlers") {
		t.Error("missing handler count")
	}
}

func TestPrintListeningNoHandlers(t *testing.T) {
	out, _ := captureOutput(func() {
		PrintListening("http://localhost:8080")
	})
	plain := strip(out)
	if strings.Contains(plain, "handlers") {
		t.Error("should not show handler count when not provided")
	}
}

// --- JSON colorized output tests ---

func TestJSONHandlerColorized(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Color: ptrBool(true)}))
	out, _ := captureOutput(func() {
		log.Info("test", Attrs{"port": 8080})
	})
	if !strings.Contains(out, "\x1b[") {
		t.Error("color=true should produce ANSI escapes")
	}
	plain := strip(out)
	if !strings.Contains(plain, `"level"`) {
		t.Error("missing level key in colorized output")
	}
	if !strings.Contains(plain, `"info"`) {
		t.Error("missing level value in colorized output")
	}
}

func TestJSONHandlerColorizedExpanded(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Expand: true, Color: ptrBool(true)}))
	out, _ := captureOutput(func() {
		log.Info("test")
	})
	if !strings.Contains(out, "\n") {
		t.Error("expanded colorized should be multi-line")
	}
	plain := strip(out)
	if !strings.Contains(plain, "{") || !strings.Contains(plain, "}") {
		t.Error("missing JSON braces")
	}
}

func TestLoggerNoAttrs(t *testing.T) {
	h := &recordingHandler{}
	log := New(h)
	log.Info("bare")

	if len(h.records[0].Attrs) != 0 {
		t.Errorf("expected empty attrs, got %v", h.records[0].Attrs)
	}
}

// Verify the JSON field ordering: level, msg, time, then attrs
func TestJSONFieldOrder(t *testing.T) {
	log := New(NewJSONHandler(&JSONHandlerOptions{Color: ptrBool(false)}))
	out, _ := captureOutput(func() {
		log.Info("hi", Attrs{"z": 1})
	})
	out = strings.TrimSpace(out)
	levelIdx := strings.Index(out, `"level"`)
	msgIdx := strings.Index(out, `"msg"`)
	timeIdx := strings.Index(out, `"time"`)
	zIdx := strings.Index(out, `"z"`)

	if levelIdx > msgIdx || msgIdx > timeIdx || timeIdx > zIdx {
		t.Errorf("wrong field order: level@%d msg@%d time@%d z@%d\noutput: %s",
			levelIdx, msgIdx, timeIdx, zIdx, out)
	}
}

func init() {
	// Silence the unused import warning for fmt in tests
	_ = fmt.Sprintf
}
