# req — curl-like HTTP engine (Go CLI)

Interactive HTTP client with arrow-key history, tab completion, browser spoofing,
and zero bloat. Single binary, two dependencies.

## Build

```bash
cd req
go mod tidy        # downloads readline + golang.org/x/sys
go build -o req .
```

Install globally:

```bash
go install .
```

## Usage

### Interactive REPL

```bash
req
```

Full readline experience — same feel as bash/zsh:

| Key | Action |
|-----|--------|
| `↑` / `↓` | Scroll command history |
| `←` / `→` | Move cursor inside the line |
| `Ctrl-A` | Jump to start of line |
| `Ctrl-E` | Jump to end of line |
| `Ctrl-R` | Incremental history search |
| `Ctrl-W` | Delete word backwards |
| `Ctrl-L` | Clear screen |
| `Ctrl-D` | Quit |
| `Tab` | Autocomplete commands, sub-commands, profile names |

History is persisted to `~/.req_history` (up to 500 entries) across sessions.

### One-shot CLI

```bash
req GET  https://httpbin.org/get
req POST https://httpbin.org/post '{"name":"alice"}'
req --spoof GET https://httpbin.org/headers   # full browser spoof
req https://example.com                        # GET shorthand
```

## REPL commands

### Requests
```
GET    <url> [body]
POST   <url> [body]
PUT    <url> [body]
PATCH  <url> [body]
DELETE <url>
```

### Session
```
set    <key> <value>       set a request header
unset  <key>               remove a header
base   <url>               set URL prefix (then use /path shorthand)
headers                    show headers box
history                    show last 50 requests
replay <n>                 re-fire history entry n
cookies                    cookie jar info
clear  [headers|history|cookies]
```

### Browser spoof
```
spoof on / off
spoof status
spoof profiles                     list all profiles
spoof profile <name|random>        pin or randomise profile
spoof rotate  on|off               new profile every request
spoof ua      on|off               randomise Chrome/Firefox version
spoof lang    on|off               rotate Accept-Language
spoof referer on|off               chain Referer header
spoof delay   <min_ms> <max_ms>    human-like random delay
spoof delay   off
```

**Profiles:** `chrome-win` · `chrome-mac` · `chrome-android` · `firefox-win` · `firefox-linux` · `safari-mac` · `edge-win`

### What spoof injects

| Header | Detail |
|--------|--------|
| `User-Agent` | Real browser string with optional version jitter |
| `Accept` | Per-browser exact value |
| `Accept-Encoding` | `gzip, deflate, br` |
| `Accept-Language` | Rotated from locale pool |
| `Cache-Control` | `max-age=0` (Chrome/Edge) |
| `Sec-Fetch-*` | Full set (`Dest`, `Mode`, `Site`, `User`) |
| `Sec-CH-UA` + `Sec-CH-UA-*` | Client hints |
| `Referer` | Chained from previous URL |
| `DNT` | Randomly on 1-in-3 requests |
| `TE: trailers` | Firefox only |
| Cookies | Persisted per-domain in session jar |
| Redirects | Spoof headers survive all redirects |

## Dependencies

| Package | Why |
|---------|-----|
| `github.com/chzyer/readline` | Arrow keys, Ctrl-R, Tab completion, history file |
| `golang.org/x/sys` | Required by readline for raw terminal mode |
