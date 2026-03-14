package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// ─── Profiles ─────────────────────────────────────────────────────────────────

type BrowserProfile struct {
	Name          string
	UserAgent     string
	StaticHeaders map[string]string
}

var browserProfiles = []BrowserProfile{
	{
		Name: "chrome-win",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		StaticHeaders: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":           "gzip, deflate, br",
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
		StaticHeaders: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":           "gzip, deflate, br",
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
		Name: "chrome-android",
		UserAgent: "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/124.0.6367.82 Mobile Safari/537.36",
		StaticHeaders: map[string]string{
			"Accept":             "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":    "gzip, deflate, br",
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
	{
		Name: "firefox-win",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) " +
			"Gecko/20100101 Firefox/125.0",
		StaticHeaders: map[string]string{
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
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
		StaticHeaders: map[string]string{
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
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
		StaticHeaders: map[string]string{
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
			"Connection":      "keep-alive",
		},
	},
	{
		Name: "edge-win",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
		StaticHeaders: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			"Accept-Encoding":           "gzip, deflate, br",
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
}

var chromeVersions = []string{"120", "121", "122", "123", "124", "125"}
var firefoxVersions = []string{"122.0", "123.0", "124.0", "125.0"}

var acceptLanguages = []string{
	"en-US,en;q=0.9",
	"en-GB,en;q=0.9",
	"en-US,en;q=0.8,fr;q=0.6",
	"en-US,en;q=0.9,de;q=0.8",
	"en-US,en;q=0.5",
	"en-CA,en;q=0.9,fr-CA;q=0.8",
}

// ─── Config ───────────────────────────────────────────────────────────────────

type SpoofConfig struct {
	Enabled       bool
	Profile       *BrowserProfile // nil = random
	RandomUA      bool
	RandomLang    bool
	FakeReferer   bool
	DelayMin      time.Duration
	DelayMax      time.Duration
	RotateProfile bool
}

func defaultSpoofConfig() SpoofConfig {
	return SpoofConfig{
		RandomUA:   true,
		RandomLang: true,
	}
}

// ─── Engine ───────────────────────────────────────────────────────────────────

func mutateUA(ua string) string {
	bumpVersion := func(ua, marker string, pool []string) string {
		idx := strings.Index(ua, marker)
		if idx < 0 {
			return ua
		}
		start := idx + len(marker)
		end := start
		for end < len(ua) && (ua[end] >= '0' && ua[end] <= '9' || ua[end] == '.') {
			end++
		}
		return ua[:start] + pool[rand.Intn(len(pool))] + ua[end:]
	}
	if strings.Contains(ua, "Chrome/") {
		ua = bumpVersion(ua, "Chrome/", chromeVersions)
	}
	if strings.Contains(ua, "Firefox/") {
		ua = bumpVersion(ua, "Firefox/", firefoxVersions)
	}
	return ua
}

func applySpoof(req *http.Request, s *State) {
	if !s.Spoof.Enabled {
		return
	}

	var profile *BrowserProfile
	if s.Spoof.RotateProfile || s.Spoof.Profile == nil {
		p := browserProfiles[rand.Intn(len(browserProfiles))]
		profile = &p
	} else {
		profile = s.Spoof.Profile
	}

	set := func(k, v string) {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	ua := profile.UserAgent
	if s.Spoof.RandomUA {
		ua = mutateUA(ua)
	}
	set("User-Agent", ua)

	for k, v := range profile.StaticHeaders {
		set(k, v)
	}

	if s.Spoof.RandomLang {
		set("Accept-Language", acceptLanguages[rand.Intn(len(acceptLanguages))])
	}

	if s.Spoof.FakeReferer && s.Referer != "" {
		set("Referer", s.Referer)
	}

	if rand.Intn(3) == 0 {
		set("DNT", "1")
	}
}

func humanDelay(s *State) {
	if s.Spoof.DelayMax <= 0 {
		return
	}
	d := s.Spoof.DelayMin
	if spread := s.Spoof.DelayMax - s.Spoof.DelayMin; spread > 0 {
		d += time.Duration(rand.Int63n(int64(spread)))
	}
	if d > 0 {
		fmt.Printf("%s  ⏱  waiting %s…%s\n", gray, d.Round(time.Millisecond), reset)
		time.Sleep(d)
	}
}
