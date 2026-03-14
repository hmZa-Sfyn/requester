package main

import (
	"fmt"
	"net/http/cookiejar"
	"time"
)

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

func (s *State) addHistory(method, url string, status int, dur time.Duration) {
	s.History = append(s.History, Entry{
		Method: method, URL: url,
		Status: status, Duration: dur, At: time.Now(),
	})
	if len(s.History) > 50 {
		s.History = s.History[len(s.History)-50:]
	}
}

// ─── Display ──────────────────────────────────────────────────────────────────

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
	printBox("Spoof", lines)
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
			"%s%2d%s  %s%-7s%s %s%-40s%s %s%3d%s %s%s%s",
			dim, i+1, reset,
			yellow, e.Method, reset,
			dim, truncate(e.URL, 38), reset,
			sc, e.Status, reset,
			gray, e.Duration.Round(time.Millisecond), reset,
		))
	}
	printBox("History", lines)
}
