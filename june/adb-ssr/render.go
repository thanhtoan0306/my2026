package main

import (
	"bytes"
	"net/http"
	"strings"
)

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func hxTarget(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("HX-Target"))
}

func renderTemplate(w http.ResponseWriter, name string, data PageData) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, name, data)
}

func renderPage(w http.ResponseWriter, data PageData) {
	if err := renderTemplate(w, "dashboard", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderHTMX(w http.ResponseWriter, r *http.Request, data PageData) {
	target := hxTarget(r)
	switch target {
	case "#devices-panel", "devices-panel":
		_ = renderTemplate(w, "devices", data)
	case "#apps-panel", "apps-panel":
		_ = renderTemplate(w, "apps", data)
	default:
		_ = renderTemplate(w, "message", data)
	}
}

func renderHTMXDual(w http.ResponseWriter, data PageData) error {
	var msgBuf, appsBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&msgBuf, "message_oob", data); err != nil {
		return err
	}
	if err := tmpl.ExecuteTemplate(&appsBuf, "apps_oob", data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(msgBuf.Bytes())
	w.Write(appsBuf.Bytes())
	return nil
}

func respond(w http.ResponseWriter, r *http.Request, data PageData) {
	if isHTMX(r) {
		renderHTMX(w, r, data)
		return
	}
	renderPage(w, data)
}
