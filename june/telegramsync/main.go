package main

import (
	"context"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "login":
			if err := runLogin(context.Background()); err != nil {
				log.Fatal(err)
			}
			log.Println("Telegram session saved.")
			return
		case "sync":
			ctx := context.Background()
			payload, err := syncTodayMessages(ctx)
			if err != nil {
				log.Fatal(err)
			}
			raw, err := marshalPayload(payload)
			if err != nil {
				log.Fatal(err)
			}
			name := "telegram-sync.json"
			_ = os.WriteFile(name, raw, 0644)
			url, err := uploadJSON(ctx, name, raw)
			if err != nil {
				log.Printf("local file %s (%d bytes); drive: %v", name, len(raw), err)
			} else {
				log.Printf("uploaded: %s", url)
			}
			return
		}
	}

	if err := initTemplates(); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8095"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex())
	mux.HandleFunc("POST /", handleIndex())

	addr := "127.0.0.1:" + port
	log.Printf("Telegram sync SSR: http://%s", addr)
	log.Printf("Chat: https://web.telegram.org/k/#%d", chatIDFromEnv())
	log.Fatal(http.ListenAndServe(addr, mux))
}
