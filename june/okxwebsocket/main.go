package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	if err := initTemplates(); err != nil {
		log.Fatal(err)
	}

	store = &tickerStore{}
	go runOKXFeed(store)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8092"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /fragment/price", handlePriceFragment)
	mux.HandleFunc("GET /fragment/status", handleStatusFragment)

	addr := "127.0.0.1:" + port
	log.Printf("OKX BTC ticker: http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
