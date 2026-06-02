package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	okxWSURL  = "wss://ws.okx.com:8443/ws/v5/public"
	instID    = "BTC-USDT"
	reconnect = 3 * time.Second
)

type Ticker struct {
	InstID    string
	Last      string
	Bid       string
	Ask       string
	Open24h   string
	High24h   string
	Low24h    string
	Vol24h    string
	ChangePct string
	UpdatedAt time.Time
	Connected bool
	Error     string
}

type tickerStore struct {
	mu     sync.RWMutex
	ticker Ticker
}

func (s *tickerStore) get() Ticker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ticker
}

func (s *tickerStore) setConnected(connected bool, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ticker.Connected = connected
	s.ticker.Error = errMsg
}

func (s *tickerStore) update(data okxTickerData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ticker.InstID = data.InstID
	s.ticker.Last = data.Last
	s.ticker.Bid = data.BidPx
	s.ticker.Ask = data.AskPx
	s.ticker.Open24h = data.Open24h
	s.ticker.High24h = data.High24h
	s.ticker.Low24h = data.Low24h
	s.ticker.Vol24h = data.VolCcy24h
	s.ticker.ChangePct = pctChange(data.Open24h, data.Last)
	s.ticker.UpdatedAt = time.Now()
	s.ticker.Connected = true
	s.ticker.Error = ""
}

type okxMessage struct {
	Event string          `json:"event"`
	Data  []okxTickerData `json:"data"`
	Msg   string          `json:"msg"`
}

type okxTickerData struct {
	InstID    string `json:"instId"`
	Last      string `json:"last"`
	BidPx     string `json:"bidPx"`
	AskPx     string `json:"askPx"`
	Open24h   string `json:"open24h"`
	High24h   string `json:"high24h"`
	Low24h    string `json:"low24h"`
	VolCcy24h string `json:"volCcy24h"`
}

func pctChange(open, last string) string {
	o, err1 := strconv.ParseFloat(open, 64)
	l, err2 := strconv.ParseFloat(last, 64)
	if err1 != nil || err2 != nil || o == 0 {
		return "—"
	}
	pct := (l - o) / o * 100
	return fmt.Sprintf("%+.2f%%", pct)
}

func runOKXFeed(store *tickerStore) {
	for {
		if err := connectAndListen(store); err != nil {
			log.Printf("okx ws: %v — reconnect in %s", err, reconnect)
			store.setConnected(false, err.Error())
		}
		time.Sleep(reconnect)
	}
}

func connectAndListen(store *tickerStore) error {
	conn, _, err := websocket.DefaultDialer.Dial(okxWSURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	sub := map[string]any{
		"op": "subscribe",
		"args": []map[string]string{
			{"channel": "tickers", "instId": instID},
		},
	}
	if err := conn.WriteJSON(sub); err != nil {
		return err
	}

	store.setConnected(true, "")
	log.Printf("okx ws: subscribed to %s tickers", instID)

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg okxMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Event == "error" {
			if msg.Msg != "" {
				return fmt.Errorf("%s", msg.Msg)
			}
			return fmt.Errorf("websocket error")
		}
		if len(msg.Data) > 0 {
			store.update(msg.Data[0])
		}
	}
}
