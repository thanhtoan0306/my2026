package main

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates *template.Template

func initTemplates() error {
	funcs := template.FuncMap{
		"pct":      formatPct,
		"bytes":    formatBytes,
		"duration": formatDuration,
		"bar":      bar,
		"trim":     trimMiddle,
	}
	t, err := template.New("root").Funcs(funcs).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return err
	}
	templates = t
	return nil
}

