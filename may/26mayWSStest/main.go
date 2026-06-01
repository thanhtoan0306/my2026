package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed templates/*
var templateFS embed.FS

type PageData struct {
	Title            string
	RenderedAt       string
	DefaultWSURL     string
	BinanceTickerURL string
}

func binanceTickerURL() string {
	if u := os.Getenv("BINANCE_TICKER_WS_URL"); u != "" {
		return u
	}
	return "wss://stream.binance.com:9443/ws/btcusdt@ticker"
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func defaultWSURL(r *http.Request) string {
	if u := os.Getenv("DEFAULT_WS_URL"); u != "" {
		return u
	}
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		scheme = "wss"
	}
	return scheme + "://" + r.Host + "/ws"
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:            "WebSocket Test",
			RenderedAt:       time.Now().UTC().Format(time.RFC3339),
			DefaultWSURL:     defaultWSURL(r),
			BinanceTickerURL: binanceTickerURL(),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /ws", handleEchoWS)

	addr := "0.0.0.0:" + port
	log.Printf("listening on http://localhost:%s (echo ws at /ws)", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleEchoWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	defer conn.Close()

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		text := strings.TrimSpace(string(msg))
		if text == "" {
			text = "(empty)"
		}
		reply := "echo: " + text
		if err := conn.WriteMessage(mt, []byte(reply)); err != nil {
			break
		}
	}
}
