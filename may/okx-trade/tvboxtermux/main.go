package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	loadEnvFile()
	if err := initTemplates(); err != nil {
		log.Fatal(err)
	}
	if hasEnvCreds() {
		log.Print("OKX credentials found (env or ~/okx-ssr/.env)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("POST /close", handleClose)
	mux.HandleFunc("GET /fragment/live", handleFragmentLive)

	bind := os.Getenv("BIND")
	if bind == "" {
		bind = "127.0.0.1"
	}
	addr := bind + ":" + port
	log.Printf("OKX positions (SSR): http://%s", addr)
	log.Printf("Credentials: ~/okx-ssr/.env or OKX_API_KEY / OKX_SECRET_KEY / OKX_PASSPHRASE env vars.")
	log.Fatal(http.ListenAndServe(addr, mux))
}
