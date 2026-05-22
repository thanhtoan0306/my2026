package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

//go:embed templates/*
var templateFS embed.FS

var (
	appConfig Config
	configMu  sync.RWMutex
	tmpl      *template.Template
)

type PageData struct {
	Config  Config
	Message string
	Error   bool
	SSHCmd  string
}

func initTemplates() error {
	var err error
	tmpl, err = template.ParseFS(templateFS, "templates/*.html")
	return err
}

func getConfig() Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return appConfig
}

func setConfig(c Config) {
	configMu.Lock()
	defer configMu.Unlock()
	appConfig = c
}

func formMap(r *http.Request) map[string]string {
	if err := r.ParseForm(); err != nil {
		return map[string]string{}
	}
	m := make(map[string]string, len(r.Form))
	for k, v := range r.Form {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

func mergeConfig(r *http.Request) Config {
	return configFromForm(formMap(r), getConfig())
}

func render(w http.ResponseWriter, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "dashboard", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	cfg := getConfig()
	data := PageData{Config: cfg}

	if r.Method == http.MethodPost {
		cfg = mergeConfig(r)
		if r.FormValue("action") == "save" {
			setConfig(cfg)
			data.Config = cfg
			data.Message = "Connection settings saved."
			render(w, data)
			return
		}
	}

	if r.Method == http.MethodGet {
		render(w, data)
		return
	}

	render(w, data)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	cfg := mergeConfig(r)
	setConfig(cfg)
	data := PageData{Config: cfg}

	var lines []string
	if out, err := adbDevices(cfg); err != nil {
		lines = append(lines, "ADB: ERROR — "+err.Error())
		data.Error = true
	} else {
		lines = append(lines, "ADB devices:\n"+out)
	}

	if cfg.SSHHost != "" && cfg.SSHUser != "" {
		if out, err := runSSH(cfg, "echo ok && uname -a"); err != nil {
			lines = append(lines, "\nSSH: ERROR — "+err.Error())
			data.Error = true
		} else {
			lines = append(lines, "\nSSH:\n"+out)
		}
	} else {
		lines = append(lines, "\nSSH: configure host and user first.")
	}

	data.Message = joinLines(lines)
	render(w, data)
}

func handleADB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cfg := mergeConfig(r)
	setConfig(cfg)
	data := PageData{Config: cfg}

	keycode := r.FormValue("keyevent")
	if keycode == "" {
		keycode = r.FormValue("keyevent_custom")
	}

	var out string
	var err error

	switch r.FormValue("do") {
	case "text":
		text := r.FormValue("input_text")
		out, err = adbInputText(cfg, text)
		if err == nil {
			out = "Sent text: " + text + "\n" + out
		}
	case "shell":
		cmd := r.FormValue("shell_cmd")
		out, err = adbShell(cfg, cmd)
	case "reboot":
		out, err = adbShell(cfg, "reboot")
	case "kill_tabs":
		out, err = adbKillAllTabs(cfg)
	default:
		if keycode != "" {
			out, err = adbKeyEvent(cfg, keycode)
			if err == nil {
				out = fmt.Sprintf("keyevent %s OK\n%s", keycode, out)
			}
		} else {
			err = fmt.Errorf("choose a button or enter a keycode")
		}
	}

	if err != nil {
		data.Error = true
		data.Message = err.Error()
	} else {
		data.Message = out
		if data.Message == "" {
			data.Message = "OK"
		}
	}
	render(w, data)
}

func handleSSH(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cfg := mergeConfig(r)
	setConfig(cfg)
	data := PageData{Config: cfg}

	cmd := r.FormValue("ssh_cmd")
	switch r.FormValue("quick") {
	case "uptime":
		cmd = "uptime"
	case "reboot":
		cmd = "reboot"
	}
	data.SSHCmd = cmd

	if cmd == "" {
		data.Error = true
		data.Message = "Enter an SSH command."
		render(w, data)
		return
	}

	out, err := runSSH(cfg, cmd)
	if err != nil {
		data.Error = true
		data.Message = err.Error()
	} else {
		data.Message = out
		if data.Message == "" {
			data.Message = "(no output)"
		}
	}
	render(w, data)
}

func joinLines(lines []string) string {
	s := ""
	for i, l := range lines {
		if i > 0 {
			s += "\n"
		}
		s += l
	}
	return s
}
