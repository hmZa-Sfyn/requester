package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// visLen returns the printable rune width of s, stripping ANSI escape codes.
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

func boxWidth(lines []string) int {
	w := 32
	for _, l := range lines {
		if v := visLen(l) + 4; v > w {
			w = v
		}
	}
	return w
}

// printBox draws a rounded box with a coloured title.
func printBox(title string, lines []string) {
	all := append([]string{title}, lines...)
	inner := boxWidth(all)

	pad := inner - utf8.RuneCountInString(title) - 3
	if pad < 1 {
		pad = 1
	}
	fmt.Println("╭─ " + bold + cyan + title + reset + " " + strings.Repeat("─", pad) + "╮")

	if len(lines) == 0 {
		sp := inner - 2 - 7
		if sp < 0 {
			sp = 0
		}
		fmt.Printf("│ %s(empty)%s%s │\n", dim, reset, strings.Repeat(" ", sp))
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

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
