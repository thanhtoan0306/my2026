package main

import (
	"strconv"
	"strings"
	"time"
)

func formatTime(ms int64) string {
	if ms == 0 {
		return "—"
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04:05")
}

func formatNum(v string, maxDecimals int) string {
	if v == "" {
		return "—"
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return v
	}
	return strconv.FormatFloat(n, 'f', maxDecimals, 64)
}

func shortID(id string) string {
	if id == "" {
		return "—"
	}
	if len(id) <= 14 {
		return id
	}
	return id[:8] + "…" + id[len(id)-4:]
}

func tradeSideLabel(buyer bool) string {
	if buyer {
		return "BUY"
	}
	return "SELL"
}

func tradeSideClass(buyer bool) string {
	if buyer {
		return "buy"
	}
	return "sell"
}

func orderSideClass(side string) string {
	if strings.EqualFold(side, "BUY") {
		return "buy"
	}
	return "sell"
}

func sumQuoteVolume(trades []Trade, buyer bool) float64 {
	var total float64
	for _, t := range trades {
		if t.IsBuyer != buyer {
			continue
		}
		q, _ := strconv.ParseFloat(t.QuoteQty, 64)
		total += q
	}
	return total
}

func formatFloat(v float64, maxDecimals int) string {
	return strconv.FormatFloat(v, 'f', maxDecimals, 64)
}

func pnlClass(v float64) string {
	if v > 0 {
		return "buy"
	}
	if v < 0 {
		return "sell"
	}
	return "muted"
}

func formatPNL(v float64) string {
	if v > 0 {
		return "+" + formatFloat(v, 4)
	}
	return formatFloat(v, 4)
}

func usdtFree(balances []Balance) string {
	for _, b := range balances {
		if b.Asset == "USDT" {
			return formatNum(b.Free, 4)
		}
	}
	return "—"
}
