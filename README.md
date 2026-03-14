# req — HTTP request engine (Go CLI, browser-spoof edition)

A curl-like HTTP client with a functional REPL, request history, and a rendered
headers box — all in a single Go file with **zero external dependencies**.

## Build

```bash
cd req
go build -o req .
```

Or install globally:

```bash
go install .
# then use `req` from anywhere
```

## Usage

### One-shot (like curl)

```bash
req GET  https://jsonplaceholder.typicode.com/posts/1
req POST https://jsonplaceholder.typicode.com/posts '{"title":"hi","body":"yo","userId":1}'
req https://httpbin.org/get          # shorthand — defaults to GET
```

### REPL mode

```bash
req
```

```
req › GET https://jsonplaceholder.typicode.com/posts/1

req › set Authorization "Bearer my-token"
req › set X-App-Version 2.1

req › headers              # shows header box

req › base https://api.example.com
[https://api.example.com] › GET /users
[https://api.example.com] › POST /users {"name":"alice"}

req › history              # shows history box (last 20 requests)
req › replay 2             # replay entry #2

req › clear headers        # wipe headers
req › clear history        # wipe history
req › clear                # wipe both

req › help
req › quit
```

## Browser spoofing

```bash
req --spoof GET https://httpbin.org/headers   # one-shot with full spoof
```

Inside the REPL:

```
req › spoof on                    # enable (random profile each request)
req › spoof profile chrome-win    # pin to a specific browser
req › spoof profile random        # back to random
req › spoof rotate on             # change profile on every request
req › spoof ua on                 # randomise Chrome/Firefox minor version
req › spoof lang on               # rotate Accept-Language locale strings
req › spoof referer on            # chain Referer from previous URL
req › spoof delay 200 800         # wait 200–800 ms before each request
req › spoof delay off
req › spoof profiles              # list all available profiles
req › spoof status                # show current config box
```

### What gets injected

| Header | Detail |
|--------|--------|
| `User-Agent` | Real Chrome/Firefox/Safari/Edge UA with optional version randomisation |
| `Accept` | Exact browser accept string (differs per browser) |
| `Accept-Encoding` | `gzip, deflate, br` |
| `Accept-Language` | Rotated from pool of real locale strings |
| `Cache-Control` | `max-age=0` (Chrome/Edge) |
| `Sec-Fetch-*` | Full `Dest`, `Mode`, `Site`, `User` set |
| `Sec-CH-UA` | Client hints with correct brand/version |
| `Sec-CH-UA-Mobile` | `?0` (desktop) or `?1` (Android) |
| `Sec-CH-UA-Platform` | `"Windows"`, `"macOS"`, `"Android"` etc. |
| `Referer` | Chained from previous request (optional) |
| `DNT` | Randomly included (1-in-3 requests) |
| `TE` | `trailers` for Firefox profiles |
| Cookies | Automatically stored & sent via session jar |
| Redirects | Spoofed headers are preserved through redirect chains |

### Available profiles

`chrome-win` · `chrome-mac` · `chrome-android` · `firefox-win` · `firefox-linux` · `safari-mac` · `edge-win`

## Features

| Feature | Detail |
|---------|--------|
| Methods | GET POST PUT PATCH DELETE HEAD OPTIONS |
| Headers | per-session, persistent, box display |
| Base URL | set once, prefix all relative paths |
| History | last 20 entries, replay by index |
| JSON body | pass inline as last arg |
| Pretty JSON | syntax-highlighted response body |
| Response headers | shown in box after each request |
| ANSI colours | status codes, keys, values, timing |
| Zero deps | stdlib only, single binary |

## Response display

```
200 OK  43ms

╭─ Response Headers ──────────────────────╮
│ Content-Type: application/json          │
│ Cache-Control: max-age=43200            │
╰─────────────────────────────────────────╯

{
  "id": 1,
  "title": "sunt aut facere ...",
  "body": "quia et suscipit ..."
}
```
