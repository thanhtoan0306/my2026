package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	if err := initTemplates(); err != nil {
		log.Fatal(err)
	}

	st := &store{}
	startSampler(st, 2*time.Second)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8095"
	}

	monitorToken = os.Getenv("MONITOR_TOKEN")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", withAuth(func(w http.ResponseWriter, r *http.Request) { handleIndex(st, w, r) }))
	mux.HandleFunc("GET /fragment/summary", withAuth(func(w http.ResponseWriter, r *http.Request) { handleSummaryFragment(st, w, r) }))
	mux.HandleFunc("GET /fragment/procs", withAuth(func(w http.ResponseWriter, r *http.Request) { handleProcsFragment(st, w, r) }))

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	addr := host + ":" + port
	log.Printf("PC monitor: http://%s", addr)
	if monitorToken != "" {
		log.Printf("Auth enabled (MONITOR_TOKEN set)")
	}
	log.Fatal(http.ListenAndServe(addr, mux))
}

