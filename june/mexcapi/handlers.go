package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"
)

type DashboardData struct {
	Title          string
	Symbol         string
	Symbols        []string
	Account        *Account
	Balances       []Balance
	Trades         []Trade
	Orders         []Order
	Positions      []PositionView
	PositionsError string
	Error          string
	Updated        string
	BuyVol         float64
	SellVol        float64
	HTMX           bool
	TradeCnt       int
	OrderCnt       int
	BalanceCnt     int
	PositionCnt    int
}

type Server struct {
	client        *Client
	futures       *FuturesClient
	telegram      *Telegram
	defaultSymbol string
}

func (s *Server) loadPositions(data *DashboardData) {
	views, err := s.futures.loadPositionViews()
	if err != nil {
		data.PositionsError = err.Error()
		return
	}
	data.Positions = views
	data.PositionCnt = len(views)
}

func (s *Server) loadDashboard(symbol string) DashboardData {
	data := DashboardData{
		Title:  "MEXC Account Trades",
		Symbol: symbol,
	}

	s.loadPositions(&data)

	acct, err := s.client.Account()
	if err != nil {
		data.Error = err.Error()
		return data
	}
	data.Account = acct
	data.Balances = nonZeroBalances(acct.Balances)
	data.BalanceCnt = len(data.Balances)
	data.Symbols = symbolsFromBalances(acct.Balances, symbol)

	trades, err := s.client.MyTrades(symbol, 100)
	if err != nil {
		data.Error = err.Error()
		return data
	}
	sort.Slice(trades, func(i, j int) bool { return trades[i].Time > trades[j].Time })
	data.Trades = trades
	data.TradeCnt = len(trades)
	data.BuyVol = sumQuoteVolume(trades, true)
	data.SellVol = sumQuoteVolume(trades, false)

	orders, err := s.client.AllOrders(symbol, 100)
	if err != nil {
		data.Error = err.Error()
		return data
	}
	sort.Slice(orders, func(i, j int) bool { return orders[i].Time > orders[j].Time })
	data.Orders = orders
	data.OrderCnt = len(orders)
	data.Updated = time.Now().Format("2006-01-02 15:04:05")

	return data
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = s.defaultSymbol
	}
	data := s.loadDashboard(strings.ToUpper(symbol))
	data.HTMX = false
	w.Header().Set("Cache-Control", "no-store")
	render(w, "index.html", data)
}

func (s *Server) handleDashboardPartial(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = s.defaultSymbol
	}
	data := s.loadDashboard(strings.ToUpper(symbol))
	data.HTMX = true
	w.Header().Set("Cache-Control", "no-store")
	render(w, "page_partial.html", data)
}

func (s *Server) handleTelegramPing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := s.telegram.Send("ping")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		render(w, "ping_status.html", map[string]string{
			"Kind":    "error",
			"Message": err.Error(),
		})
		return
	}

	render(w, "ping_status.html", map[string]string{
		"Kind":    "ok",
		"Message": "Sent “ping” to Telegram",
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":        true,
		"service":   "mexcssr",
		"mode":      "ssr+htmx",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	if path == "" || strings.Contains(path, "..") {
		http.NotFound(w, r)
		return
	}
	data, err := staticFS.ReadFile("static/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(data)
}
