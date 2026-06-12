package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type PositionPoller struct {
	futures   *FuturesClient
	telegram  *Telegram
	interval  time.Duration
	minDelta  float64
	lastPNL   map[int64]float64
	seen      map[int64]bool
	mu        sync.Mutex
}

func NewPositionPoller(futures *FuturesClient, telegram *Telegram) *PositionPoller {
	interval := 20 * time.Second
	if v := os.Getenv("PNL_POLL_INTERVAL_SEC"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			interval = time.Duration(sec) * time.Second
		}
	}

	minDelta := 0.1
	if v := os.Getenv("PNL_ALERT_DELTA_USDT"); v != "" {
		if d, err := strconv.ParseFloat(v, 64); err == nil && d > 0 {
			minDelta = d
		}
	}

	return &PositionPoller{
		futures:  futures,
		telegram: telegram,
		interval: interval,
		minDelta: minDelta,
		lastPNL:  make(map[int64]float64),
		seen:     make(map[int64]bool),
	}
}

func (p *PositionPoller) Start() {
	if p.telegram == nil || !p.telegram.Enabled() {
		log.Printf("pnl poller: disabled (telegram not configured)")
		return
	}

	log.Printf("pnl poller: every %s, alert when unrealized PNL rises +%.2f USDT vs previous poll", p.interval, p.minDelta)
	go func() {
		// First tick establishes baseline without alerts.
		p.poll()
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		for range ticker.C {
			p.poll()
		}
	}()
}

func (p *PositionPoller) poll() {
	views, err := p.futures.loadPositionViews()
	if err != nil {
		log.Printf("pnl poller: fetch positions: %v", err)
		return
	}

	active := make(map[int64]bool, len(views))
	for _, v := range views {
		id := v.Position.PositionID
		active[id] = true

		if !v.HasPNL {
			continue
		}

		cur := v.UnrealizedPNL

		p.mu.Lock()
		prev, hadPrev := p.lastPNL[id]
		firstSeen := !p.seen[id]
		p.seen[id] = true
		p.lastPNL[id] = cur
		p.mu.Unlock()

		if firstSeen || !hadPrev {
			continue
		}

		delta := cur - prev
		if delta < p.minDelta {
			continue
		}

		msg := formatPNLAlert(v, prev, cur, delta)
		if err := p.telegram.Send(msg); err != nil {
			log.Printf("pnl poller: telegram %s: %v", v.Position.Symbol, err)
			continue
		}
		log.Printf("pnl poller: alert sent %s +%.4f USDT", v.DisplaySymbol, delta)
	}

	p.mu.Lock()
	for id := range p.lastPNL {
		if !active[id] {
			delete(p.lastPNL, id)
			delete(p.seen, id)
		}
	}
	p.mu.Unlock()
}

func formatPNLAlert(v PositionView, prev, cur, delta float64) string {
	return fmt.Sprintf(
		"📈 Unrealized PNL up +%.4f USDT\n%s · %s\nPNL: %s USDT (was %s)\nChange: +%.2f%% margin · Mark %s",
		delta,
		v.DisplaySymbol,
		v.SideLabel,
		formatPNL(cur),
		formatPNL(prev),
		v.PNLPercent,
		formatFloat(v.FairPrice, 8),
	)
}
