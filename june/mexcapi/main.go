package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	for _, path := range []string{".env", "june/mexcapi/.env"} {
		loadDotEnv(path)
	}

	if err := initTemplates(); err != nil {
		log.Fatal(err)
	}

	apiKey := requireEnv("MEXC_API_KEY")
	secret := requireEnv("MEXC_SECRET_KEY")
	if apiKey == "" || secret == "" {
		log.Fatal("MEXC_API_KEY and MEXC_SECRET_KEY must be set in .env (see .env.example)")
	}

	symbol := os.Getenv("MEXC_SYMBOL")
	if symbol == "" {
		symbol = "HYPEUSDT"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8094"
	}

	tg := NewTelegram(os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID"))

	srv := &Server{
		client:        NewClient(apiKey, secret),
		futures:       NewFuturesClient(apiKey, secret),
		telegram:      tg,
		defaultSymbol: symbol,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /static/", handleStatic)
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /htmx/telegram/ping", srv.handleTelegramPing)
	mux.HandleFunc("GET /htmx/dashboard", srv.handleDashboardPartial)
	mux.HandleFunc("GET /", srv.handleIndex)

	NewPositionPoller(srv.futures, tg).Start()

	addr := "127.0.0.1:" + port
	log.Printf("mexcssr: http://%s (Go SSR + HTMX)", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
