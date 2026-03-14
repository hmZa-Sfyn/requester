package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// ─── ANSI colours ─────────────────────────────────────────────────────────────

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	blue   = "\033[34m"
	gray   = "\033[90m"
	purple = "\033[35m"
)

// ─── Browser fingerprint profiles ─────────────────────────────────────────────

type BrowserProfile struct {
	Name          string
	UserAgent     string
	HeaderOrder   []string
	StaticHeaders map[string]string
}

var browserProfiles = []BrowserProfile{
	{
		Name: "chrome-win",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		HeaderOrder: []string{
			"Host", "Connection", "Cache-Control", "Upgrade-Insecure-Requests",
			"User-Agent", "Accept", "Accept-Encoding", "Accept-Language", "Cookie",
		},
		StaticHeaders: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept-Language":           "en-US,en;q=0.9",
			"Cache-Control":             "max-age=0",
			"Connection":                "keep-alive",
			"Upgrade-Insecure-Requests": "1",
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
			"Sec-CH-UA":                 `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
			"Sec-CH-UA-Mobile":          "?0",
			"Sec-CH-UA-Platform":        `"Windows"`,
		},
	},
	{
		Name: "chrome-mac",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		HeaderOrder: []string{
			"Host", "Connection", "Cache-Control", "Upgrade-Insecure-Requests",
			"User-Agent", "Accept", "Accept-Encoding", "Accept-Language", "Cookie",
		},
		StaticHeaders: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept-Language":           "en-US,en;q=0.9",
			"Cache-Control":             "max-age=0",
			"Connection":                "keep-alive",
			"Upgrade-Insecure-Requests": "1",
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
			"Sec-CH-UA":                 `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
			"Sec-CH-UA-Mobile":          "?0",
			"Sec-CH-UA-Platform":        `"macOS"`,
		},
	},
	{
		Name: "firefox-win",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) " +
			"Gecko/20100101 Firefox/125.0",
		HeaderOrder: []string{
			"Host", "User-Agent", "Accept", "Accept-Language",
			"Accept-Encoding", "Connection", "Cookie",
		},
		StaticHeaders: map[string]string{
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept-Language": "en-US,en;q=0.5",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "document",
			"Sec-Fetch-Mode":  "navigate",
			"Sec-Fetch-Site":  "none",
			"Sec-Fetch-User":  "?1",
			"TE":              "trailers",
		},
	},
	{
		Name: "firefox-linux",
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) " +
			"Gecko/20100101 Firefox/125.0",
		HeaderOrder: []string{
			"Host", "User-Agent", "Accept", "Accept-Language",
			"Accept-Encoding", "Connection", "Cookie",
		},
		StaticHeaders: map[string]string{
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept-Language": "en-US,en;q=0.5",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "document",
			"Sec-Fetch-Mode":  "navigate",
			"Sec-Fetch-Site":  "none",
			"Sec-Fetch-User":  "?1",
		},
	},
	{
		Name: "safari-mac",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_1) " +
			"AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15",
		HeaderOrder: []string{
			"Host", "Accept", "Accept-Encoding", "Accept-Language",
			"Connection", "User-Agent", "Cookie",
		},
		StaticHeaders: map[string]string{
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept-Language": "en-US,en;q=0.9",
			"Connection":      "keep-alive",
		},
	},
	{
		Name: "edge-win",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
		HeaderOrder: []string{
			"Host", "Connection", "Cache-Control", "Upgrade-Insecure-Requests",
			"User-Agent", "Accept", "Accept-Encoding", "Accept-Language", "Cookie",
		},
		StaticHeaders: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept-Language":           "en-US,en;q=0.9",
			"Cache-Control":             "max-age=0",
			"Connection":                "keep-alive",
			"Upgrade-Insecure-Requests": "1",
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
			"Sec-CH-UA":                 `"Microsoft Edge";v="124", "Chromium";v="124", "Not-A.Brand";v="99"`,
			"Sec-CH-UA-Mobile":          "?0",
			"Sec-CH-UA-Platform":        `"Windows"`,
		},
	},
	{
		Name: "chrome-android",
		UserAgent: "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/124.0.6367.82 Mobile Safari/537.36",
		HeaderOrder: []string{
			"Host", "Connection", "User-Agent", "Accept",
			"Accept-Encoding", "Accept-Language", "Cookie",
		},
		StaticHeaders: map[string]string{
			"Accept":             "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":    "gzip, deflate, br",
			"Accept-Language":    "en-US,en;q=0.9",
			"Connection":         "keep-alive",
			"Sec-Fetch-Dest":     "document",
			"Sec-Fetch-Mode":     "navigate",
			"Sec-Fetch-Site":     "none",
			"Sec-Fetch-User":     "?1",
			"Sec-CH-UA":          `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
			"Sec-CH-UA-Mobile":   "?1",
			"Sec-CH-UA-Platform": `"Android"`,
		},
	},
}

var chromeVersions = []string{"120", "121", "122", "123", "124", "125"}
var firefoxVersions = []string{"122.0", "123.0", "124.0", "125.0"}

// ─── Accept-Language pools ─────────────────────────────────────────────────────

var acceptLanguages = []string{
	"en-US,en;q=0.9",
	"en-GB,en;q=0.9",
	"en-US,en;q=0.8,fr;q=0.6",
	"en-US,en;q=0.9,de;q=0.8",
	"en-US,en;q=0.5",
	"en-CA,en;q=0.9,fr-CA;q=0.8",
}

// ─── Spoof config ──────────────────────────────────────────────────────────────

type SpoofConfig struct {
	Enabled       bool
	Profile       *BrowserProfile
	RandomUA      bool
	RandomLang    bool
	FakeReferer   bool
	DelayMin      time.Duration
	DelayMax      time.Duration
	RotateProfile bool
}

func defaultSpoofConfig() SpoofConfig {
	return SpoofConfig{
		Enabled:    false,
		RandomUA:   true,
		RandomLang: true,
	}
}

// ─── State ─────────────────────────────────────────────────────────────────────

type Entry struct {
	Method   string
	URL      string
	Status   int
	Duration time.Duration
	At       time.Time
}

type State struct {
	Headers   map[string]string
	History   []Entry
	BaseURL   string
	Spoof     SpoofConfig
	CookieJar *cookiejar.Jar
	Referer   string
}

func NewState() *State {
	jar, _ := cookiejar.New(nil)
	return &State{
		Headers: map[string]string{
			"Content-Type": "application/json",
			"User-Agent":   "req/1.0",
		},
		Spoof:     defaultSpoofConfig(),
		CookieJar: jar,
		Referer:   "https://www.google.com/",
	}
}

// ─── Spoof engine ──────────────────────────────────────────────────────────────

func mutateUA(ua string) string {
	if strings.Contains(ua, "Chrome/") {
		ver := chromeVersions[rand.Intn(len(chromeVersions))]
		start := strings.Index(ua, "Chrome/")
		if start >= 0 {
			end := start + 7
			for end < len(ua) && ua[end] >= '0' && ua[end] <= '9' {
				end++
			}
			return ua[:start+7] + ver + ua[end:]
		}
	}
	if strings.Contains(ua, "Firefox/") {
		ver := firefoxVersions[rand.Intn(len(firefoxVersions))]
		start := strings.Index(ua, "Firefox/")
		if start >= 0 {
			end := start + 8
			for end < len(ua) && (ua[end] >= '0' && ua[end] <= '9' || ua[end] == '.') {
				end++
			}
			return ua[:start+8] + ver + ua[end:]
		}
	}
	return ua
}

func applySpoof(req *http.Request, s *State) {
	if !s.Spoof.Enabled {
		return
	}

	// choose profile
	var profile *BrowserProfile
	if s.Spoof.RotateProfile || s.Spoof.Profile == nil {
		p := browserProfiles[rand.Intn(len(browserProfiles))]
		profile = &p
	} else {
		profile = s.Spoof.Profile
	}

	setIfMissing := func(k, v string) {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	// User-Agent
	ua := profile.UserAgent
	if s.Spoof.RandomUA {
		ua = mutateUA(ua)
	}
	setIfMissing("User-Agent", ua)

	// all profile static headers
	for k, v := range profile.StaticHeaders {
		setIfMissing(k, v)
	}

	// random Accept-Language
	if s.Spoof.RandomLang {
		setIfMissing("Accept-Language", acceptLanguages[rand.Intn(len(acceptLanguages))])
	}

	// Referer chaining
	if s.Spoof.FakeReferer && s.Referer != "" {
		setIfMissing("Referer", s.Referer)
	}

	// random DNT (1-in-3 browsers send it)
	if rand.Intn(3) == 0 {
		setIfMissing("DNT", "1")
	}
}

func humanDelay(s *State) {
	if s.Spoof.DelayMax <= 0 {
		return
	}
	d := s.Spoof.DelayMin
	spread := s.Spoof.DelayMax - s.Spoof.DelayMin
	if spread > 0 {
		d += time.Duration(rand.Int63n(int64(spread)))
	}
	if d > 0 {
		fmt.Printf("%s  ⏱  waiting %s…%s\n", gray, d.Round(time.Millisecond), reset)
		time.Sleep(d)
	}
}

// ─── Box helpers ───────────────────────────────────────────────────────────────

func visLen(s string) int {
	inEsc := false
	n := 0
	for _, r := range s {
		switch {
		case r == '\033':
			inEsc = true
		case inEsc:
			if r == 'm' {
				inEsc = false
			}
		default:
			n++
		}
	}
	return n
}

func boxWidth(content []string) int {
	w := 32
	for _, line := range content {
		if l := visLen(line) + 4; l > w {
			w = l
		}
	}
	return w
}

func printBox(title string, lines []string) {
	inner := boxWidth(append(lines, title))
	pad := inner - utf8.RuneCountInString(title) - 3
	if pad < 1 {
		pad = 1
	}
	fmt.Println("╭─ " + bold + cyan + title + reset + " " + strings.Repeat("─", pad) + "╮")
	if len(lines) == 0 {
		fmt.Printf("│ %s%-*s%s │\n", dim, inner-2, "(empty)", reset)
	}
	for _, l := range lines {
		sp := inner - 2 - visLen(l)
		if sp < 0 {
			sp = 0
		}
		fmt.Printf("│ %s%s │\n", l, strings.Repeat(" ", sp))
	}
	fmt.Println("╰" + strings.Repeat("─", inner) + "╯")
}

func padRight(s string, n int) string {
	l := utf8.RuneCountInString(s)
	if l >= n {
		return s
	}
	return s + strings.Repeat(" ", n-l)
}

// ─── Status / display ──────────────────────────────────────────────────────────

func statusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return yellow
	case code >= 400:
		return red
	}
	return red + bold
}

func printHeaders(s *State) {
	lines := make([]string, 0, len(s.Headers))
	for k, v := range s.Headers {
		lines = append(lines, fmt.Sprintf("%s%s%s: %s%s%s", cyan, k, reset, dim, v, reset))
	}
	printBox("Headers", lines)
}

func printSpoofStatus(s *State) {
	onOff := func(b bool) string {
		if b {
			return green + "on" + reset
		}
		return red + "off" + reset
	}
	pName := "random"
	if s.Spoof.Profile != nil {
		pName = s.Spoof.Profile.Name
	}
	lines := []string{
		fmt.Sprintf("enabled      %s", onOff(s.Spoof.Enabled)),
		fmt.Sprintf("profile      %s%s%s", cyan, pName, reset),
		fmt.Sprintf("rotate       %s", onOff(s.Spoof.RotateProfile)),
		fmt.Sprintf("random UA    %s", onOff(s.Spoof.RandomUA)),
		fmt.Sprintf("random lang  %s", onOff(s.Spoof.RandomLang)),
		fmt.Sprintf("referer      %s", onOff(s.Spoof.FakeReferer)),
	}
	if s.Spoof.DelayMax > 0 {
		lines = append(lines, fmt.Sprintf("delay        %s%s – %s%s",
			yellow,
			s.Spoof.DelayMin.Round(time.Millisecond),
			s.Spoof.DelayMax.Round(time.Millisecond),
			reset))
	} else {
		lines = append(lines, fmt.Sprintf("delay        %soff%s", dim, reset))
	}
	printBox("Spoof config", lines)
}

func printHistory(s *State) {
	if len(s.History) == 0 {
		printBox("History", nil)
		return
	}
	lines := make([]string, 0, len(s.History))
	for i, e := range s.History {
		sc := statusColor(e.Status)
		lines = append(lines, fmt.Sprintf(
			"%s%2d%s  %s%-7s%s %s%-40s%s %s%d%s %s%s%s",
			dim, i+1, reset,
			yellow, e.Method, reset,
			dim, truncate(e.URL, 38), reset,
			sc, e.Status, reset,
			gray, e.Duration.Round(time.Millisecond), reset,
		))
	}
	printBox("History", lines)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func printHelp() {
	lines := []string{
		fmt.Sprintf("%sGET%s    <url> [body]          %sSend GET request%s", yellow, reset, dim, reset),
		fmt.Sprintf("%sPOST%s   <url> [body]          %sSend POST request%s", yellow, reset, dim, reset),
		fmt.Sprintf("%sPUT%s    <url> [body]          %sSend PUT request%s", yellow, reset, dim, reset),
		fmt.Sprintf("%sPATCH%s  <url> [body]          %sSend PATCH request%s", yellow, reset, dim, reset),
		fmt.Sprintf("%sDELETE%s <url>                 %sSend DELETE request%s", yellow, reset, dim, reset),
		"",
		fmt.Sprintf("%sset%s    <key> <value>         %sSet a header%s", cyan, reset, dim, reset),
		fmt.Sprintf("%sunset%s  <key>                 %sRemove a header%s", cyan, reset, dim, reset),
		fmt.Sprintf("%sbase%s   <url>                 %sSet base URL prefix%s", cyan, reset, dim, reset),
		fmt.Sprintf("%sheaders%s                      %sShow headers box%s", cyan, reset, dim, reset),
		fmt.Sprintf("%shistory%s                      %sShow request history%s", cyan, reset, dim, reset),
		fmt.Sprintf("%sreplay%s  <n>                  %sReplay history entry%s", cyan, reset, dim, reset),
		fmt.Sprintf("%scookies%s                      %sCookie jar info%s", cyan, reset, dim, reset),
		fmt.Sprintf("%sclear%s   [headers|history|cookies]", cyan, reset),
		"",
		fmt.Sprintf("%s── Browser spoof ─────────────────────────────────────%s", gray, reset),
		fmt.Sprintf("%sspoof on%s / %sspoof off%s            %sEnable/disable%s", purple, reset, purple, reset, dim, reset),
		fmt.Sprintf("%sspoof status%s                  %sShow spoof config%s", purple, reset, dim, reset),
		fmt.Sprintf("%sspoof profiles%s                %sList all profiles%s", purple, reset, dim, reset),
		fmt.Sprintf("%sspoof profile%s <name|random>   %sPin or randomise profile%s", purple, reset, dim, reset),
		fmt.Sprintf("%sspoof rotate%s  on|off          %sNew profile each request%s", purple, reset, dim, reset),
		fmt.Sprintf("%sspoof ua%s      on|off          %sRandomise version numbers%s", purple, reset, dim, reset),
		fmt.Sprintf("%sspoof lang%s    on|off          %sRandomise Accept-Language%s", purple, reset, dim, reset),
		fmt.Sprintf("%sspoof referer%s on|off          %sSend chained Referer%s", purple, reset, dim, reset),
		fmt.Sprintf("%sspoof delay%s   <min> <max> ms  %sHuman-like delay%s", purple, reset, dim, reset),
		fmt.Sprintf("%sspoof delay off%s               %sDisable delay%s", purple, reset, dim, reset),
		"",
		fmt.Sprintf("%sProfiles:%s chrome-win · chrome-mac · chrome-android", dim, reset),
		fmt.Sprintf("          firefox-win · firefox-linux · safari-mac · edge-win"),
		"",
		fmt.Sprintf("%shelp%s / %squit%s", cyan, reset, cyan, reset),
	}
	printBox("Commands", lines)
}

// ─── HTTP engine ───────────────────────────────────────────────────────────────

func doRequest(s *State, method, rawURL, body string) {
	targetURL := rawURL
	if s.BaseURL != "" && !strings.HasPrefix(rawURL, "http") {
		targetURL = strings.TrimRight(s.BaseURL, "/") + "/" + strings.TrimLeft(rawURL, "/")
	}

	humanDelay(s)

	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		fmt.Printf("%s✗ Bad request: %v%s\n", red, err, reset)
		return
	}

	// user headers first
	for k, v := range s.Headers {
		req.Header.Set(k, v)
	}
	// spoof layer fills the gaps
	applySpoof(req, s)

	if s.Spoof.Enabled {
		ua := req.Header.Get("User-Agent")
		fmt.Printf("%s  ⌨  fingerprint: %s%s%s\n", gray, dim, ua, reset)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     s.CookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			// preserve spoofed headers through redirects
			for k, v := range via[0].Header {
				req.Header[k] = v
			}
			return nil
		},
	}

	start := time.Now()
	resp, err := client.Do(req)
	dur := time.Since(start)

	if err != nil {
		fmt.Printf("%s✗ Request failed: %v%s\n", red, err, reset)
		return
	}
	defer resp.Body.Close()

	if s.Spoof.FakeReferer {
		s.Referer = targetURL
	}

	s.History = append(s.History, Entry{
		Method: method, URL: targetURL,
		Status: resp.StatusCode, Duration: dur, At: time.Now(),
	})
	if len(s.History) > 20 {
		s.History = s.History[len(s.History)-20:]
	}

	printResponse(resp, dur)
}

func printResponse(resp *http.Response, dur time.Duration) {
	sc := statusColor(resp.StatusCode)
	fmt.Printf("\n%s%s%d %s%s  %s%s%s\n",
		sc, bold, resp.StatusCode, http.StatusText(resp.StatusCode), reset,
		gray, dur.Round(time.Millisecond), reset,
	)

	var hlines []string
	for k, vv := range resp.Header {
		hlines = append(hlines, fmt.Sprintf("%s%s%s: %s%s%s", cyan, k, reset, dim, strings.Join(vv, ", "), reset))
	}
	printBox("Response Headers", hlines)

	raw, _ := io.ReadAll(resp.Body)
	if len(raw) == 0 {
		fmt.Printf("%s(empty body)%s\n\n", dim, reset)
		return
	}

	var pretty bytes.Buffer
	if json.Indent(&pretty, raw, "", "  ") == nil {
		colorJSON(pretty.String())
	} else {
		fmt.Println(string(raw))
	}
	fmt.Println()
}

// ─── JSON syntax highlighting ──────────────────────────────────────────────────

func colorJSON(s string) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]

		switch {
		case strings.HasPrefix(trimmed, `"`) && strings.Contains(trimmed, ":"):
			if idx := strings.Index(trimmed, `":`); idx >= 0 {
				key := trimmed[:idx+2]
				val := strings.TrimSpace(trimmed[idx+2:])
				fmt.Printf("%s%s%s%s%s\n", indent, cyan+key+reset, colorValue(val), reset, "")
				continue
			}
		case strings.HasPrefix(trimmed, `"`):
			fmt.Printf("%s%s%s%s\n", indent, green, trimmed, reset)
			continue
		case trimmed == "{" || trimmed == "}" || trimmed == "[" || trimmed == "]" ||
			strings.HasSuffix(trimmed, "{") || strings.HasSuffix(trimmed, "["):
			fmt.Printf("%s%s%s%s\n", indent, gray, trimmed, reset)
			continue
		}
		fmt.Println(line)
	}
}

func colorValue(v string) string {
	v2 := strings.TrimRight(v, ",")
	switch {
	case v2 == "true" || v2 == "false":
		return yellow + v + reset
	case v2 == "null":
		return red + v + reset
	case len(v2) > 0 && v2[0] == '"':
		return green + v + reset
	default:
		if _, err := strconv.ParseFloat(v2, 64); err == nil {
			return blue + v + reset
		}
	}
	return v
}

// ─── REPL ──────────────────────────────────────────────────────────────────────

func repl(s *State) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("\n%s%s req%s — HTTP request engine  %stype %shelp%s%s for commands%s\n\n",
		bold, cyan, reset, dim, cyan, reset, dim, reset)

	for {
		spoofTag := ""
		if s.Spoof.Enabled {
			pName := "rnd"
			if s.Spoof.Profile != nil {
				pName = s.Spoof.Profile.Name
			}
			spoofTag = fmt.Sprintf(" %s[%s]%s", purple, pName, reset)
		}
		if s.BaseURL != "" {
			fmt.Printf("%s[%s]%s%s %s›%s ", gray, s.BaseURL, reset, spoofTag, cyan, reset)
		} else {
			fmt.Printf("%sreq%s%s %s›%s ", cyan, reset, spoofTag, dim, reset)
		}

		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !handleLine(s, line) {
			break
		}
	}
	fmt.Printf("\n%sBye.%s\n", dim, reset)
}

// ─── Command handler ───────────────────────────────────────────────────────────

func handleLine(s *State, line string) bool {
	parts := splitLine(line)
	if len(parts) == 0 {
		return true
	}
	cmd := strings.ToUpper(parts[0])

	switch cmd {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		if len(parts) < 2 {
			fmt.Printf("%s✗ Usage: %s <url> [body]%s\n", red, cmd, reset)
			return true
		}
		body := ""
		if len(parts) >= 3 {
			body = strings.Join(parts[2:], " ")
		}
		doRequest(s, cmd, parts[1], body)

	case "SET":
		if len(parts) < 3 {
			fmt.Printf("%s✗ Usage: set <key> <value>%s\n", red, reset)
			return true
		}
		s.Headers[parts[1]] = strings.Join(parts[2:], " ")
		fmt.Printf("%s✓ Header set%s\n", green, reset)
		printHeaders(s)

	case "UNSET":
		if len(parts) < 2 {
			fmt.Printf("%s✗ Usage: unset <key>%s\n", red, reset)
			return true
		}
		delete(s.Headers, parts[1])
		fmt.Printf("%s✓ Header removed%s\n", green, reset)
		printHeaders(s)

	case "BASE":
		if len(parts) < 2 {
			s.BaseURL = ""
			fmt.Printf("%s✓ Base URL cleared%s\n", green, reset)
			return true
		}
		s.BaseURL = parts[1]
		fmt.Printf("%s✓ Base URL: %s%s%s\n", green, cyan, s.BaseURL, reset)

	case "HEADERS":
		printHeaders(s)

	case "HISTORY":
		printHistory(s)

	case "REPLAY":
		if len(parts) < 2 {
			fmt.Printf("%s✗ Usage: replay <n>%s\n", red, reset)
			return true
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n < 1 || n > len(s.History) {
			fmt.Printf("%s✗ Invalid index (1–%d)%s\n", red, len(s.History), reset)
			return true
		}
		e := s.History[n-1]
		fmt.Printf("%sReplaying %s%s %s%s%s\n", dim, yellow, e.Method, cyan, e.URL, reset)
		doRequest(s, e.Method, e.URL, "")

	case "CLEAR":
		sub := ""
		if len(parts) >= 2 {
			sub = strings.ToLower(parts[1])
		}
		switch sub {
		case "headers":
			s.Headers = map[string]string{}
			fmt.Printf("%s✓ Headers cleared%s\n", green, reset)
		case "history":
			s.History = nil
			fmt.Printf("%s✓ History cleared%s\n", green, reset)
		case "cookies":
			s.CookieJar, _ = cookiejar.New(nil)
			fmt.Printf("%s✓ Cookie jar cleared%s\n", green, reset)
		case "":
			s.Headers = map[string]string{}
			s.History = nil
			s.CookieJar, _ = cookiejar.New(nil)
			fmt.Printf("%s✓ Cleared headers, history and cookies%s\n", green, reset)
		default:
			fmt.Printf("%s✗ clear [headers|history|cookies]%s\n", red, reset)
		}

	case "SPOOF":
		handleSpoof(s, parts[1:])

	case "COOKIES":
		lines := []string{
			fmt.Sprintf("%sCookies auto-sent per domain via session jar.%s", dim, reset),
			fmt.Sprintf("Run %sclear cookies%s to wipe.", cyan+bold, reset),
		}
		printBox("Cookie Jar", lines)

	case "HELP", "?":
		printHelp()

	case "QUIT", "EXIT", "Q":
		return false

	default:
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
			fmt.Printf("%s✗ Unknown command: %s  (type help)%s\n", red, parts[0], reset)
		}
	}
	return true
}

// ─── Spoof sub-commands ────────────────────────────────────────────────────────

func handleSpoof(s *State, args []string) {
	if len(args) == 0 {
		printSpoofStatus(s)
		return
	}
	sub := strings.ToLower(args[0])

	switch sub {
	case "on":
		s.Spoof.Enabled = true
		fmt.Printf("%s✓ Spoof enabled%s\n", green, reset)
		printSpoofStatus(s)

	case "off":
		s.Spoof.Enabled = false
		s.Spoof.Profile = nil
		fmt.Printf("%s✓ Spoof disabled%s\n", green, reset)

	case "status":
		printSpoofStatus(s)

	case "profiles":
		lines := make([]string, 0, len(browserProfiles))
		for _, p := range browserProfiles {
			lines = append(lines, fmt.Sprintf("%s%-18s%s %s%s%s",
				cyan, p.Name, reset,
				dim, truncate(p.UserAgent, 55), reset))
		}
		printBox("Profiles", lines)

	case "profile":
		if len(args) < 2 {
			fmt.Printf("%s✗ Usage: spoof profile <name|random>%s\n", red, reset)
			return
		}
		name := strings.ToLower(args[1])
		if name == "random" || name == "rand" {
			s.Spoof.Profile = nil
			fmt.Printf("%s✓ Profile: random%s\n", green, reset)
			return
		}
		for i, p := range browserProfiles {
			if p.Name == name {
				s.Spoof.Profile = &browserProfiles[i]
				fmt.Printf("%s✓ Profile pinned: %s%s%s\n", green, cyan, name, reset)
				fmt.Printf("%s  %s%s\n", gray, dim+s.Spoof.Profile.UserAgent, reset)
				return
			}
		}
		fmt.Printf("%s✗ Unknown profile '%s'. Run: spoof profiles%s\n", red, name, reset)

	case "rotate":
		if len(args) < 2 {
			fmt.Printf("%s✗ Usage: spoof rotate on|off%s\n", red, reset)
			return
		}
		s.Spoof.RotateProfile = strings.ToLower(args[1]) == "on"
		fmt.Printf("%s✓ Rotate profile: %v%s\n", green, s.Spoof.RotateProfile, reset)

	case "ua":
		if len(args) < 2 {
			fmt.Printf("%s✗ Usage: spoof ua on|off%s\n", red, reset)
			return
		}
		s.Spoof.RandomUA = strings.ToLower(args[1]) == "on"
		fmt.Printf("%s✓ Random UA version: %v%s\n", green, s.Spoof.RandomUA, reset)

	case "lang":
		if len(args) < 2 {
			fmt.Printf("%s✗ Usage: spoof lang on|off%s\n", red, reset)
			return
		}
		s.Spoof.RandomLang = strings.ToLower(args[1]) == "on"
		fmt.Printf("%s✓ Random Accept-Language: %v%s\n", green, s.Spoof.RandomLang, reset)

	case "referer":
		if len(args) < 2 {
			fmt.Printf("%s✗ Usage: spoof referer on|off%s\n", red, reset)
			return
		}
		s.Spoof.FakeReferer = strings.ToLower(args[1]) == "on"
		fmt.Printf("%s✓ Referer chaining: %v%s\n", green, s.Spoof.FakeReferer, reset)

	case "delay":
		if len(args) < 2 || strings.ToLower(args[1]) == "off" {
			s.Spoof.DelayMin = 0
			s.Spoof.DelayMax = 0
			fmt.Printf("%s✓ Delay off%s\n", green, reset)
			return
		}
		if len(args) < 3 {
			fmt.Printf("%s✗ Usage: spoof delay <min_ms> <max_ms>%s\n", red, reset)
			return
		}
		minMs, e1 := strconv.Atoi(args[1])
		maxMs, e2 := strconv.Atoi(args[2])
		if e1 != nil || e2 != nil || minMs < 0 || maxMs < minMs {
			fmt.Printf("%s✗ Invalid values (e.g. spoof delay 200 800)%s\n", red, reset)
			return
		}
		s.Spoof.DelayMin = time.Duration(minMs) * time.Millisecond
		s.Spoof.DelayMax = time.Duration(maxMs) * time.Millisecond
		fmt.Printf("%s✓ Delay: %d–%d ms%s\n", green, minMs, maxMs, reset)

	default:
		fmt.Printf("%s✗ Unknown spoof sub-command '%s'%s\n", red, sub, reset)
	}
}

// ─── Line parser ───────────────────────────────────────────────────────────────

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

// ─── Usage ─────────────────────────────────────────────────────────────────────

func usage() {
	fmt.Printf(`%sreq%s — HTTP request engine

%sUsage:%s
  req                              interactive REPL
  req <METHOD> <url> [body]        one-shot request
  req <url>                        GET shorthand
  req --spoof <METHOD> <url>       one-shot with all browser spoof on

%sExamples:%s
  req GET https://httpbin.org/get
  req POST https://httpbin.org/post '{"name":"alice"}'
  req --spoof GET https://httpbin.org/headers
  req https://example.com

%sREPL:%s  type %shelp%s inside the REPL
`, bold+cyan, reset, bold, reset, bold, reset, bold, reset, cyan, reset)
}

// ─── main ──────────────────────────────────────────────────────────────────────

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
			fmt.Fprintf(os.Stderr, "%s✗ Unknown: %s%s\n", red, args[0], reset)
			usage()
			os.Exit(1)
		}
	}

	_ = url.QueryEscape // keep import used
}
