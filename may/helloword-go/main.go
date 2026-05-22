package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	bind := os.Getenv("BIND")
	if bind == "" {
		bind = "0.0.0.0"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleHello)

	addr := bind + ":" + port
	log.Printf("Hello page: http://127.0.0.1:%s", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleHello(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Hello World</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 2rem; background: #0f172a; color: #e2e8f0; }
    h1 { color: #38bdf8; }
  </style>
</head>
<body>
  <h1>Hello World</h1>
  <p>Served by Go on Termux.</p>
</body>
</html>`))
}
