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

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))

	addr := ":" + port
	log.Printf("Wallet app: http://127.0.0.1%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
