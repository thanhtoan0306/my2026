package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

var pageTpl *template.Template

func initTemplates() error {
	var err error
	pageTpl, err = template.ParseFS(templateFS, "templates/*.html")
	return err
}

type pageData struct {
	ChatID      int64
	ChatURL     string
	Message     string
	Error       bool
	LastJSON    string
	DriveURL    string
	HasAPI      bool
	HasSession  bool
}

func hasSession() bool {
	_, err := os.Stat(sessionDir())
	return err == nil
}

func handleIndex() http.HandlerFunc {
	chatID := chatIDFromEnv()
	_, _, apiErr := apiCredentials()

	return func(w http.ResponseWriter, r *http.Request) {
		data := pageData{
			ChatID:     chatID,
			ChatURL:    fmt.Sprintf("https://web.telegram.org/k/#%d", chatID),
			HasAPI:     apiErr == nil,
			HasSession: hasSession(),
		}

		if r.Method == http.MethodPost && r.FormValue("action") == "sync" {
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
			defer cancel()

			payload, err := syncTodayMessages(ctx)
			if err != nil {
				data.Message = err.Error()
				data.Error = true
				render(w, data)
				return
			}

			raw, err := marshalPayload(payload)
			if err != nil {
				data.Message = err.Error()
				data.Error = true
				render(w, data)
				return
			}

			filename := fmt.Sprintf("telegram-%d-%s.json", chatID, payload.Date)
			if err := os.WriteFile(filename, raw, 0644); err != nil {
				data.Message = "saved sync failed locally: " + err.Error()
				data.Error = true
				render(w, data)
				return
			}

			driveURL, err := uploadJSON(ctx, filename, raw)
			if err != nil {
				data.Message = fmt.Sprintf("Synced %d messages locally as %s; Drive upload failed: %v", payload.Count, filename, err)
				data.LastJSON = filename
			} else {
				data.Message = fmt.Sprintf("Synced %d messages for today → %s (also on Drive)", payload.Count, filename)
				data.DriveURL = driveURL
				data.LastJSON = filename
			}
			render(w, data)
			return
		}

		render(w, data)
	}
}

func render(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTpl.ExecuteTemplate(w, "index", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
