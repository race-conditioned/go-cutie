// showcase — run with `go run ./showcase` to see every log output cutie can produce.
package main

import (
	"fmt"
	"time"

	cutie "github.com/race-conditioned/go-cutie"
)

func main() {
	// ── PrettyHandler ──────────────────────────────────────────────────────────────

	fmt.Print("\n── PrettyHandler ─────────────────────────────────────────\n\n")

	pretty := cutie.New(&cutie.PrettyHandler{})

	pretty.Debug("cache miss", cutie.Attrs{"key": "user:42", "ttl": 300})
	pretty.Info("server started", cutie.Attrs{"port": 8080, "stage": "local"})
	pretty.Warn("pool exhausted", cutie.Attrs{"active": 50, "max": 50})
	pretty.Error("query failed", cutie.Attrs{"err": "connection refused", "db": "postgres"})

	// ── PrettyHandler with .With() ─────────────────────────────────────────────────

	fmt.Print("\n── PrettyHandler + With() ────────────────────────────────\n\n")

	svc := pretty.With(cutie.Attrs{"service": "billing", "requestId": "req_abc123"})

	svc.Info("charge created", cutie.Attrs{"amount": 4999, "currency": "usd"})
	svc.Error("refund failed", cutie.Attrs{"err": "insufficient funds", "chargeId": "ch_xyz"})

	// ── JSONHandler (compact, default) ─────────────────────────────────────────────

	fmt.Print("\n── JSONHandler (compact) ────────────────────────────────\n\n")

	colorTrue := true
	jsonLog := cutie.New(cutie.NewJSONHandler(&cutie.JSONHandlerOptions{Color: &colorTrue}))

	jsonLog.Debug("cache miss", cutie.Attrs{"key": "user:42", "ttl": 300})
	jsonLog.Info("server started", cutie.Attrs{"port": 8080, "stage": "local"})
	jsonLog.Warn("pool exhausted", cutie.Attrs{"active": 50, "max": 50})
	jsonLog.Error("query failed", cutie.Attrs{"err": "connection refused", "db": "postgres"})

	// ── JSONHandler (expanded) ─────────────────────────────────────────────────────

	fmt.Print("\n── JSONHandler (expanded default) ──────────────────────\n\n")

	jsonExpanded := cutie.New(cutie.NewJSONHandler(&cutie.JSONHandlerOptions{Expand: true, Color: &colorTrue}))

	jsonExpanded.Info("server started", cutie.Attrs{"port": 8080, "stage": "local"})
	jsonExpanded.Error("query failed", cutie.Attrs{"err": "connection refused", "db": "postgres"})

	// ── JSONHandler per-call override ──────────────────────────────────────────────

	fmt.Print("\n── JSONHandler per-call override ───────────────────────\n\n")

	fmt.Println("compact default → force expanded:")
	jsonLog.Expanded().Info("expanded override", cutie.Attrs{"port": 8080})

	fmt.Println("\nexpanded default → force compact:")
	jsonExpanded.Compact().Warn("compact override", cutie.Attrs{"active": 50, "max": 50})

	// ── JSONHandler no-color (production) ──────────────────────────────────────────

	fmt.Print("\n── JSONHandler (no color, compact) ─────────────────────\n\n")

	colorFalse := false
	jsonPlain := cutie.New(cutie.NewJSONHandler(&cutie.JSONHandlerOptions{Color: &colorFalse}))

	jsonPlain.Info("server started", cutie.Attrs{"port": 8080, "stage": "production"})
	jsonPlain.Error("query failed", cutie.Attrs{"err": "connection refused"})

	fmt.Print("\n── JSONHandler (no color, expanded) ────────────────────\n\n")

	jsonPlainExpanded := cutie.New(cutie.NewJSONHandler(&cutie.JSONHandlerOptions{Expand: true, Color: &colorFalse}))

	jsonPlainExpanded.Info("server started", cutie.Attrs{"port": 8080, "stage": "production"})
	jsonPlainExpanded.Error("query failed", cutie.Attrs{"err": "connection refused"})

	// ── JSONHandler with .With() ───────────────────────────────────────────────────

	fmt.Print("\n── JSONHandler + With() ─────────────────────────────────\n\n")

	jsonSvc := jsonLog.With(cutie.Attrs{"service": "billing"})

	jsonSvc.Info("charge created", cutie.Attrs{"amount": 4999})
	jsonSvc.Error("refund failed", cutie.Attrs{"err": "insufficient funds"})

	// ── PrintBanner — basic ────────────────────────────────────────────────────────

	fmt.Print("\n── PrintBanner (all keys) ───────────────────────────────\n\n")

	cfg := cutie.Attrs{
		"port":        8080,
		"stage":       "local",
		"store":       "postgres",
		"cacheDriver": "redis",
		"logLevel":    "debug",
	}

	cutie.PrintBanner("my-app", cfg)

	// ── PrintBanner — pick (flat) ──────────────────────────────────────────────────

	fmt.Print("── PrintBanner (pick) ──────────────────────────────────\n\n")

	cutie.PrintBannerPick("my-app", cfg, []string{"stage", "port", "store"})

	// ── PrintBanner — pick (grouped) ───────────────────────────────────────────────

	fmt.Print("── PrintBanner (grouped) ───────────────────────────────\n\n")

	cutie.PrintBannerGrouped("my-app", cfg, [][]string{
		{"port", "stage"},
		{"store", "cacheDriver"},
		{"logLevel"},
	})

	// ── PrintBanner — long value truncation ────────────────────────────────────────

	fmt.Print("── PrintBanner (long value) ────────────────────────────\n\n")

	cutie.PrintBanner("my-app", cutie.Attrs{
		"dsn":   "postgres://user:password@very-long-hostname.us-east-1.rds.amazonaws.com:5432/mydb",
		"stage": "production",
	})

	// ── PrintListening ─────────────────────────────────────────────────────────────

	fmt.Print("── PrintListening ─────────────────────────────────────\n\n")

	cutie.PrintListening("http://localhost:8080", 11)
	cutie.PrintListening("http://localhost:3000")

	// ── PrintAccess — method + status matrix ───────────────────────────────────────

	fmt.Print("── PrintAccess ────────────────────────────────────────\n\n")

	cutie.PrintAccess(cutie.AccessRecord{Method: "GET", Path: "/users", Status: 200, Duration: 2 * time.Millisecond})
	cutie.PrintAccess(cutie.AccessRecord{Method: "GET", Path: "/users/42", Status: 304, Duration: 1 * time.Millisecond})
	cutie.PrintAccess(cutie.AccessRecord{Method: "POST", Path: "/users", Status: 201, Duration: 15 * time.Millisecond})
	cutie.PrintAccess(cutie.AccessRecord{Method: "PUT", Path: "/users/42", Status: 200, Duration: 8 * time.Millisecond})
	cutie.PrintAccess(cutie.AccessRecord{Method: "DELETE", Path: "/users/42/sessions", Status: 204, Duration: 3 * time.Millisecond})
	cutie.PrintAccess(cutie.AccessRecord{Method: "GET", Path: "/admin/secrets", Status: 403, Duration: 0})
	cutie.PrintAccess(cutie.AccessRecord{Method: "GET", Path: "/missing", Status: 404, Duration: 1 * time.Millisecond})
	cutie.PrintAccess(cutie.AccessRecord{Method: "POST", Path: "/webhooks/stripe", Status: 500, Duration: 230 * time.Millisecond})
	cutie.PrintAccess(cutie.AccessRecord{Method: "PATCH", Path: "/users/42", Status: 200, Duration: 5 * time.Millisecond})

	fmt.Println()
}
