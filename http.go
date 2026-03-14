package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

	for k, v := range s.Headers {
		req.Header.Set(k, v)
	}
	applySpoof(req, s)

	if s.Spoof.Enabled {
		fmt.Printf("%s  ⌨  UA: %s%s%s\n", gray, dim, req.Header.Get("User-Agent"), reset)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     s.CookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
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
		fmt.Printf("%s✗ %v%s\n", red, err, reset)
		return
	}
	defer resp.Body.Close()

	if s.Spoof.FakeReferer {
		s.Referer = targetURL
	}
	s.addHistory(method, targetURL, resp.StatusCode, dur)
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
		hlines = append(hlines, fmt.Sprintf("%s%s%s: %s%s%s",
			cyan, k, reset, dim, strings.Join(vv, ", "), reset))
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

func colorJSON(src string) {
	scanner := bufio.NewScanner(strings.NewReader(src))
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
