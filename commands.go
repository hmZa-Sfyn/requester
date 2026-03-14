package main

import (
	"fmt"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

// handleLine dispatches a REPL line. Returns false to quit.
func handleLine(s *State, line string) bool {
	parts := splitLine(line)
	if len(parts) == 0 {
		return true
	}
	cmd := strings.ToUpper(parts[0])

	switch cmd {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		if len(parts) < 2 {
			fmt.Printf("%s✗ usage: %s <url> [body]%s\n", red, cmd, reset)
			return true
		}
		body := ""
		if len(parts) >= 3 {
			body = strings.Join(parts[2:], " ")
		}
		doRequest(s, cmd, parts[1], body)

	case "SET":
		if len(parts) < 3 {
			fmt.Printf("%s✗ usage: set <key> <value>%s\n", red, reset)
			return true
		}
		s.Headers[parts[1]] = strings.Join(parts[2:], " ")
		fmt.Printf("%s✓ header set%s\n", green, reset)
		printHeaders(s)

	case "UNSET":
		if len(parts) < 2 {
			fmt.Printf("%s✗ usage: unset <key>%s\n", red, reset)
			return true
		}
		delete(s.Headers, parts[1])
		fmt.Printf("%s✓ header removed%s\n", green, reset)
		printHeaders(s)

	case "BASE":
		if len(parts) < 2 {
			s.BaseURL = ""
			fmt.Printf("%s✓ base URL cleared%s\n", green, reset)
			return true
		}
		s.BaseURL = parts[1]
		fmt.Printf("%s✓ base: %s%s%s\n", green, cyan, s.BaseURL, reset)

	case "HEADERS":
		printHeaders(s)

	case "HISTORY":
		printHistory(s)

	case "REPLAY":
		if len(parts) < 2 {
			fmt.Printf("%s✗ usage: replay <n>%s\n", red, reset)
			return true
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n < 1 || n > len(s.History) {
			fmt.Printf("%s✗ invalid index (1–%d)%s\n", red, len(s.History), reset)
			return true
		}
		e := s.History[n-1]
		fmt.Printf("%s↩ replaying %s%s %s%s%s\n", gray, yellow, e.Method, cyan, e.URL, reset)
		doRequest(s, e.Method, e.URL, "")

	case "CLEAR":
		sub := ""
		if len(parts) >= 2 {
			sub = strings.ToLower(parts[1])
		}
		switch sub {
		case "headers":
			s.Headers = map[string]string{}
			fmt.Printf("%s✓ headers cleared%s\n", green, reset)
		case "history":
			s.History = nil
			fmt.Printf("%s✓ history cleared%s\n", green, reset)
		case "cookies":
			s.CookieJar, _ = cookiejar.New(nil)
			fmt.Printf("%s✓ cookies cleared%s\n", green, reset)
		case "":
			s.Headers = map[string]string{}
			s.History = nil
			s.CookieJar, _ = cookiejar.New(nil)
			fmt.Printf("%s✓ cleared headers, history and cookies%s\n", green, reset)
		default:
			fmt.Printf("%s✗ clear [headers|history|cookies]%s\n", red, reset)
		}

	case "SPOOF":
		handleSpoof(s, parts[1:])

	case "COOKIES":
		lines := []string{
			fmt.Sprintf("%sSession cookie jar — sent automatically per domain.%s", dim, reset),
			fmt.Sprintf("Run %sclear cookies%s to wipe.", cyan+bold, reset),
		}
		printBox("Cookies", lines)

	case "HELP", "?":
		printHelp()

	case "QUIT", "EXIT", "Q":
		return false

	default:
		// bare URL → GET (or POST if body follows)
		if strings.HasPrefix(parts[0], "http") || strings.HasPrefix(parts[0], "/") {
			body := ""
			if len(parts) >= 2 {
				body = strings.Join(parts[1:], " ")
			}
			method := "GET"
			if body != "" {
				method = "POST"
			}
			doRequest(s, method, parts[0], body)
		} else {
			fmt.Printf("%s✗ unknown command: %s  (type help)%s\n", red, parts[0], reset)
		}
	}
	return true
}

// ─── Spoof sub-commands ───────────────────────────────────────────────────────

func handleSpoof(s *State, args []string) {
	if len(args) == 0 {
		printSpoofStatus(s)
		return
	}
	sub := strings.ToLower(args[0])

	switch sub {
	case "on":
		s.Spoof.Enabled = true
		fmt.Printf("%s✓ spoof on%s\n", green, reset)
		printSpoofStatus(s)

	case "off":
		s.Spoof.Enabled = false
		s.Spoof.Profile = nil
		fmt.Printf("%s✓ spoof off%s\n", green, reset)

	case "status":
		printSpoofStatus(s)

	case "profiles":
		lines := make([]string, 0, len(browserProfiles))
		for _, p := range browserProfiles {
			lines = append(lines, fmt.Sprintf("%s%-18s%s %s%s%s",
				cyan, p.Name, reset,
				dim, truncate(p.UserAgent, 56), reset))
		}
		printBox("Profiles", lines)

	case "profile":
		if len(args) < 2 {
			fmt.Printf("%s✗ usage: spoof profile <name|random>%s\n", red, reset)
			return
		}
		name := strings.ToLower(args[1])
		if name == "random" || name == "rand" {
			s.Spoof.Profile = nil
			fmt.Printf("%s✓ profile: random%s\n", green, reset)
			return
		}
		for i, p := range browserProfiles {
			if p.Name == name {
				s.Spoof.Profile = &browserProfiles[i]
				fmt.Printf("%s✓ profile: %s%s%s\n", green, cyan, name, reset)
				return
			}
		}
		fmt.Printf("%s✗ unknown profile '%s'  (try: spoof profiles)%s\n", red, name, reset)

	case "rotate":
		if len(args) < 2 {
			fmt.Printf("%s✗ usage: spoof rotate on|off%s\n", red, reset)
			return
		}
		s.Spoof.RotateProfile = strings.ToLower(args[1]) == "on"
		fmt.Printf("%s✓ rotate: %v%s\n", green, s.Spoof.RotateProfile, reset)

	case "ua":
		if len(args) < 2 {
			fmt.Printf("%s✗ usage: spoof ua on|off%s\n", red, reset)
			return
		}
		s.Spoof.RandomUA = strings.ToLower(args[1]) == "on"
		fmt.Printf("%s✓ random UA version: %v%s\n", green, s.Spoof.RandomUA, reset)

	case "lang":
		if len(args) < 2 {
			fmt.Printf("%s✗ usage: spoof lang on|off%s\n", red, reset)
			return
		}
		s.Spoof.RandomLang = strings.ToLower(args[1]) == "on"
		fmt.Printf("%s✓ random accept-language: %v%s\n", green, s.Spoof.RandomLang, reset)

	case "referer":
		if len(args) < 2 {
			fmt.Printf("%s✗ usage: spoof referer on|off%s\n", red, reset)
			return
		}
		s.Spoof.FakeReferer = strings.ToLower(args[1]) == "on"
		fmt.Printf("%s✓ referer chaining: %v%s\n", green, s.Spoof.FakeReferer, reset)

	case "delay":
		if len(args) < 2 || strings.ToLower(args[1]) == "off" {
			s.Spoof.DelayMin, s.Spoof.DelayMax = 0, 0
			fmt.Printf("%s✓ delay off%s\n", green, reset)
			return
		}
		if len(args) < 3 {
			fmt.Printf("%s✗ usage: spoof delay <min_ms> <max_ms>%s\n", red, reset)
			return
		}
		minMs, e1 := strconv.Atoi(args[1])
		maxMs, e2 := strconv.Atoi(args[2])
		if e1 != nil || e2 != nil || minMs < 0 || maxMs < minMs {
			fmt.Printf("%s✗ invalid values (e.g. spoof delay 200 800)%s\n", red, reset)
			return
		}
		s.Spoof.DelayMin = time.Duration(minMs) * time.Millisecond
		s.Spoof.DelayMax = time.Duration(maxMs) * time.Millisecond
		fmt.Printf("%s✓ delay: %d–%d ms%s\n", green, minMs, maxMs, reset)

	default:
		fmt.Printf("%s✗ unknown spoof sub-command '%s'%s\n", red, sub, reset)
	}
}

// ─── Help ─────────────────────────────────────────────────────────────────────

func printHelp() {
	section := func(s string) string {
		return fmt.Sprintf("%s── %s %s", gray, s, reset)
	}
	lines := []string{
		section("Requests"),
		fmt.Sprintf("  %sGET%s    <url> [body]", yellow, reset),
		fmt.Sprintf("  %sPOST%s   <url> [body]", yellow, reset),
		fmt.Sprintf("  %sPUT%s    <url> [body]", yellow, reset),
		fmt.Sprintf("  %sPATCH%s  <url> [body]", yellow, reset),
		fmt.Sprintf("  %sDELETE%s <url>", yellow, reset),
		"",
		section("Session"),
		fmt.Sprintf("  %sset%s    <key> <value>      %sset header%s", cyan, reset, dim, reset),
		fmt.Sprintf("  %sunset%s  <key>              %sremove header%s", cyan, reset, dim, reset),
		fmt.Sprintf("  %sbase%s   <url>              %sset URL prefix%s", cyan, reset, dim, reset),
		fmt.Sprintf("  %sheaders%s                   %sshow headers box%s", cyan, reset, dim, reset),
		fmt.Sprintf("  %shistory%s                   %sshow history box%s", cyan, reset, dim, reset),
		fmt.Sprintf("  %sreplay%s  <n>               %sreplay entry n%s", cyan, reset, dim, reset),
		fmt.Sprintf("  %scookies%s                   %scookie jar info%s", cyan, reset, dim, reset),
		fmt.Sprintf("  %sclear%s   [headers|history|cookies]", cyan, reset),
		"",
		section("Browser spoof"),
		fmt.Sprintf("  %sspoof on%s / %sspoof off%s", purple, reset, purple, reset),
		fmt.Sprintf("  %sspoof status%s", purple, reset),
		fmt.Sprintf("  %sspoof profiles%s", purple, reset),
		fmt.Sprintf("  %sspoof profile%s <name|random>", purple, reset),
		fmt.Sprintf("  %sspoof rotate%s  on|off", purple, reset),
		fmt.Sprintf("  %sspoof ua%s      on|off       %srandomise version%s", purple, reset, dim, reset),
		fmt.Sprintf("  %sspoof lang%s    on|off       %srandomise locale%s", purple, reset, dim, reset),
		fmt.Sprintf("  %sspoof referer%s on|off       %schain Referer%s", purple, reset, dim, reset),
		fmt.Sprintf("  %sspoof delay%s   <min> <max>  %sms, human timing%s", purple, reset, dim, reset),
		"",
		section("Navigation"),
		fmt.Sprintf("  %s↑ / ↓%s  scroll command history", dim, reset),
		fmt.Sprintf("  %s← / →%s  move cursor in line", dim, reset),
		fmt.Sprintf("  %sCtrl-R%s  search history", dim, reset),
		fmt.Sprintf("  %sCtrl-L%s  clear screen", dim, reset),
		fmt.Sprintf("  %sTab%s     complete commands", dim, reset),
		"",
		fmt.Sprintf("  %squit%s / exit", cyan, reset),
	}
	printBox("req — commands", lines)
}

// ─── Utilities ────────────────────────────────────────────────────────────────

// splitLine tokenises a line, honouring double-quoted strings.
func splitLine(s string) []string {
	var parts []string
	var cur strings.Builder
	inQ := false
	for _, r := range s {
		switch {
		case r == '"' && !inQ:
			inQ = true
		case r == '"' && inQ:
			inQ = false
		case r == ' ' && !inQ:
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}
