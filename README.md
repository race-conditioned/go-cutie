# cutie

A minimal, pretty structured logger for Go. Zero dependencies.

```
go get github.com/race-conditioned/go-cutie@latest
```

Two handlers, one logger, one banner. Pick `PrettyHandler` for local dev, `JSONHandler` for production.

## Usage

```go
import cutie "github.com/race-conditioned/go-cutie"

// Pretty output for dev
log := cutie.New(&cutie.PrettyHandler{})

// JSON output for prod
log := cutie.New(cutie.NewJSONHandler(nil))

log.Info("server started", cutie.Attrs{"port": 8080})
log.Error("query failed", cutie.Attrs{"err": "connection refused"})
```

### PrettyHandler

Colored, columnar output. No configuration needed.

```
  DEBUG  cache miss       key=user:42  ttl=300
  INFO   server started   port=8080  stage=local
  WARN   pool exhausted   active=50  max=50
  ERROR  query failed     err="connection refused"
```

### JSONHandler

Syntax-highlighted JSON in the terminal, plain JSON when piped.

```go
// Defaults: compact, auto-detect color from TTY
log := cutie.New(cutie.NewJSONHandler(nil))

// Expanded, color forced off
noColor := false
log := cutie.New(cutie.NewJSONHandler(&cutie.JSONHandlerOptions{
    Expand: true,
    Color:  &noColor,
}))
```

Compact:
```json
{"level":"info","msg":"server started","time":"2026-04-15T10:00:00.000Z","port":8080}
```

Expanded:
```json
{
  "level": "info",
  "msg": "server started",
  "time": "2026-04-15T10:00:00.000Z",
  "port": 8080
}
```

### Enriching logs with With()

`With()` returns a new logger with base attributes merged into every record. The parent logger is never mutated.

```go
svc := log.With(cutie.Attrs{"service": "billing", "requestId": "req_abc"})
svc.Info("charge created", cutie.Attrs{"amount": 4999})
// INFO  charge created  amount=4999  requestId=req_abc  service=billing
```

### Per-call format override

```go
log.Expanded().Info("this one is multi-line", cutie.Attrs{"x": 1})
log.Compact().Info("this one is single-line", cutie.Attrs{"x": 1})
```

## Banners

Styled startup boxes for printing your app config at boot.

```go
cfg := cutie.Attrs{
    "port":  8080,
    "stage": "local",
    "store": "postgres",
}

// All keys (sorted)
cutie.PrintBanner("my-app", cfg)

// Pick specific keys, preserve order
cutie.PrintBannerPick("my-app", cfg, []string{"port", "stage"})

// Grouped with dividers between sections
cutie.PrintBannerGrouped("my-app", cfg, [][]string{
    {"port", "stage"},
    {"store"},
})
```

```
╭──────────────────────────────────────────────╮
│                    my-app                    │
├──────────────────────────────────────────────┤
│  port              8080                      │
│  stage             local                     │
├──────────────────────────────────────────────┤
│  store             postgres                  │
╰──────────────────────────────────────────────╯
```

Long values are automatically truncated with `…`.

## Listening

```go
cutie.PrintListening("http://localhost:8080", 11)
// ▶  http://localhost:8080  ·  11 handlers

cutie.PrintListening("http://localhost:3000")
// ▶  http://localhost:3000
```

## Access logging

Colored HTTP access lines for dev.

```go
cutie.PrintAccess(cutie.AccessRecord{
    Method:   "GET",
    Path:     "/users",
    Status:   200,
    Duration: 2 * time.Millisecond,
})
```

```
→  GET     /users          200  2ms
→  POST    /users          201  15ms
→  DELETE  /users/42       204  3ms
→  GET     /missing        404  1ms
→  POST    /webhooks       500  230ms
```

Methods and status codes are color-coded: GET cyan, POST blue, PUT yellow, DELETE magenta. Status 2xx cyan, 3xx/4xx yellow, 5xx red.

## Output routing

`debug` and `info` write to stdout. `warn` and `error` write to stderr.

## License

MIT
