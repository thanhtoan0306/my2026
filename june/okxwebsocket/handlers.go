package main

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

var (
	store *tickerStore
	tmpl  *template.Template
)

type PageData struct {
	Ticker Ticker
}

func initTemplates() error {
	funcs := template.FuncMap{
		"changeClass": func(pct string) string {
			if strings.HasPrefix(pct, "+") {
				return "up"
			}
			if strings.HasPrefix(pct, "-") {
				return "down"
			}
			return "neutral"
		},
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Format("15:04:05")
		},
	}
	var err error
	tmpl, err = template.New("").Funcs(funcs).ParseFS(templateFS, "templates/*.html")
	return err
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := PageData{Ticker: store.get()}
	if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePriceFragment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := PageData{Ticker: store.get()}
	if err := tmpl.ExecuteTemplate(w, "price_fragment", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleStatusFragment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := PageData{Ticker: store.get()}
	if err := tmpl.ExecuteTemplate(w, "status_fragment", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
