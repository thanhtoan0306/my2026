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

	appConfig = configFromEnv()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("POST /", handleIndex)
	mux.HandleFunc("POST /status", handleStatus)
	mux.HandleFunc("POST /adb", handleADB)

	addr := "127.0.0.1:" + port
	log.Printf("Control ADB: http://%s", addr)
	log.Printf("Requires adb in PATH. Enable wireless debugging on the device.")
	log.Fatal(http.ListenAndServe(addr, mux))
}
