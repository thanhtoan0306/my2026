package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func formatPct(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "—"
	}
	return fmt.Sprintf("%.1f%%", v)
}

func formatBytes(b uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case b >= TB:
		return fmt.Sprintf("%.2f TB", float64(b)/TB)
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/MB)
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/KB)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "—"
	}
	sec := int(d.Seconds())
	days := sec / 86400
	sec %= 86400
	h := sec / 3600
	sec %= 3600
	m := sec / 60
	s := sec % 60

	if days > 0 {
		return fmt.Sprintf("%dd %02dh %02dm", days, h, m)
	}
	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func bar(p float64, width int) string {
	if width <= 0 {
		width = 20
	}
	if math.IsNaN(p) || math.IsInf(p, 0) {
		p = 0
	}
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	filled := int(math.Round((p / 100) * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func trimMiddle(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:1]
	}
	keepL := max / 2
	keepR := max - keepL - 1
	if keepR < 0 {
		keepR = 0
	}
	return s[:keepL] + "…" + s[len(s)-keepR:]
}

