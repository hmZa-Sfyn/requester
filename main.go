package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

// ─── Tab completer ────────────────────────────────────────────────────────────

var topLevelCmds = []string{
	"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
	"set", "unset", "base", "headers", "history", "replay", "cookies",
	"clear", "spoof", "help", "quit", "exit",
}

var spoofSubCmds = []string{
	"on", "off", "status", "profiles", "profile",
	"rotate", "ua", "lang", "referer", "delay",
}

var clearSubCmds = []string{"headers", "history", "cookies"}

var profileNames = []string{
	"chrome-win", "chrome-mac", "chrome-android",
	"firefox-win", "firefox-linux",
	"safari-mac", "edge-win", "random",
}

func buildCompleter() *readline.PrefixCompleter {
	profileItems := make([]readline.PrefixCompleterInterface, len(profileNames))
	for i, n := range profileNames {
		profileItems[i] = readline.PcItem(n)
	}

	spoofItems := []readline.PrefixCompleterInterface{
		readline.PcItem("on"),
		readline.PcItem("off"),
		readline.PcItem("status"),
		readline.PcItem("profiles"),
		readline.PcItem("profile", profileItems...),
		readline.PcItem("rotate",
			readline.PcItem("on"),
			readline.PcItem("off"),
		),
		readline.PcItem("ua",
			readline.PcItem("on"),
			readline.PcItem("off"),
		),
		readline.PcItem("lang",
			readline.PcItem("on"),
			readline.PcItem("off"),
		),
		readline.PcItem("referer",
			readline.PcItem("on"),
			readline.PcItem("off"),
		),
		readline.PcItem("delay",
			readline.PcItem("off"),
		),
	}

	return readline.NewPrefixCompleter(
		readline.PcItem("GET"),
		readline.PcItem("POST"),
		readline.PcItem("PUT"),
		readline.PcItem("PATCH"),
		readline.PcItem("DELETE"),
		readline.PcItem("HEAD"),
		readline.PcItem("OPTIONS"),
		readline.PcItem("set"),
		readline.PcItem("unset"),
		readline.PcItem("base"),
		readline.PcItem("headers"),
		readline.PcItem("history"),
		readline.PcItem("replay"),
		readline.PcItem("cookies"),
		readline.PcItem("clear",
			readline.PcItem("headers"),
			readline.PcItem("history"),
			readline.PcItem("cookies"),
		),
		readline.PcItem("spoof", spoofItems...),
		readline.PcItem("help"),
		readline.PcItem("quit"),
		readline.PcItem("exit"),
	)
}

// ─── Prompt builder ───────────────────────────────────────────────────────────

func buildPrompt(s *State) string {
	// readline renders the prompt itself — we must NOT include ANSI colour codes
	// directly, but we can use \001 / \002 wrappers so readline measures width
	// correctly.
	esc := func(code string) string {
		return "\001" + code + "\002"
	}

	tag := ""
	if s.Spoof.Enabled {
		pName := "rnd"
		if s.Spoof.Profile != nil {
			pName = s.Spoof.Profile.Name
		}
		tag = esc(purple) + "[" + pName + "]" + esc(reset) + " "
	}

	if s.BaseURL != "" {
		return esc(gray) + s.BaseURL + esc(reset) + " " + tag + esc(cyan) + "›" + esc(reset) + " "
	}
	return esc(cyan) + "req" + esc(reset) + " " + tag + esc(dim) + "›" + esc(reset) + " "
}

// ─── REPL ─────────────────────────────────────────────────────────────────────

func repl(s *State) {
	// History file in home dir
	histFile := ""
	if home, err := os.UserHomeDir(); err == nil {
		histFile = home + "/.req_history"
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 buildPrompt(s),
		HistoryFile:            histFile,
		HistoryLimit:           500,
		AutoComplete:           buildCompleter(),
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		HistorySearchFold:      true, // case-insensitive Ctrl-R
		DisableAutoSaveHistory: false,
	})
	if err != nil {
		// fallback to plain bufio if readline init fails
		fmt.Fprintf(os.Stderr, "%swarn: readline unavailable (%v), using plain stdin%s\n", yellow, err, reset)
		replPlain(s)
		return
	}
	defer rl.Close()

	fmt.Printf("\n%s%sreq%s  HTTP engine  %s↑↓ history · Tab complete · Ctrl-R search%s\n\n",
		bold, cyan, reset, dim, reset)

	for {
		// refresh prompt (spoof tag / base URL may have changed)
		rl.SetPrompt(buildPrompt(s))

		line, err := rl.Readline()
		if err != nil {
			// Ctrl-D or EOF
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !handleLine(s, line) {
			break
		}
	}
	fmt.Printf("\n%sBye.%s\n", dim, reset)
}

// replPlain is a simple fallback when readline isn't available (e.g. piped stdin).
func replPlain(s *State) {
	fmt.Printf("\n%s%sreq%s  HTTP engine  (plain mode)%s\n\n", bold, cyan, reset, reset)
	rl, _ := readline.NewEx(&readline.Config{
		Prompt: "> ",
	})
	if rl == nil {
		return
	}
	defer rl.Close()
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !handleLine(s, line) {
			break
		}
	}
}

// ─── One-shot CLI ─────────────────────────────────────────────────────────────

func usage() {
	fmt.Printf(`%sreq%s — curl-like HTTP engine

%sUsage:%s
  req                                 interactive REPL (arrow keys, history, tab complete)
  req <METHOD> <url> [body]           one-shot request
  req <url>                           GET shorthand
  req --spoof <METHOD> <url> [body]   one-shot with full browser spoof

%sExamples:%s
  req GET  https://httpbin.org/get
  req POST https://httpbin.org/post '{"name":"alice"}'
  req --spoof GET https://httpbin.org/headers
  req https://example.com

%sKey bindings (REPL):%s
  ↑ / ↓      scroll command history
  ← / →      move cursor left / right
  Ctrl-A     jump to start of line
  Ctrl-E     jump to end of line
  Ctrl-R     incremental history search
  Ctrl-W     delete word backwards
  Ctrl-L     clear screen
  Ctrl-D     quit
  Tab        autocomplete

`, bold+cyan, reset, bold, reset, bold, reset, bold, reset)
}

// ─── Entry point ──────────────────────────────────────────────────────────────

func main() {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck
	s := NewState()
	args := os.Args[1:]

	if len(args) == 0 {
		repl(s)
		return
	}
	if args[0] == "--help" || args[0] == "-h" {
		usage()
		return
	}

	if args[0] == "--spoof" {
		s.Spoof.Enabled = true
		s.Spoof.RandomUA = true
		s.Spoof.RandomLang = true
		s.Spoof.FakeReferer = true
		args = args[1:]
	}

	if len(args) == 0 {
		usage()
		return
	}

	first := strings.ToUpper(args[0])
	switch first {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "%s✗ URL required%s\n", red, reset)
			os.Exit(1)
		}
		body := ""
		if len(args) >= 3 {
			body = strings.Join(args[2:], " ")
		}
		doRequest(s, first, args[1], body)
	default:
		if strings.HasPrefix(args[0], "http") {
			doRequest(s, "GET", args[0], "")
		} else {
			fmt.Fprintf(os.Stderr, "%s✗ unknown: %s%s\n", red, args[0], reset)
			usage()
			os.Exit(1)
		}
	}
}
