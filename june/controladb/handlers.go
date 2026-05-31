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
	Config      Config
	Message     string
	Error       bool
	TerminalCmd string
	ShellCmd    string
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

	if r.Method == http.MethodPost && r.FormValue("action") == "save" {
		cfg = mergeConfig(r)
		setConfig(cfg)
		data.Config = cfg
		data.Message = "Device IP saved: " + cfg.ADBHost
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
	if cfg.ADBHost == "" {
		lines = append(lines, "Enter device IP and save, then check status.")
		data.Error = true
	} else {
		if msg, err := adbConnect(cfg); err != nil {
			lines = append(lines, "Connect: ERROR — "+err.Error())
			data.Error = true
		} else if msg != "" {
			lines = append(lines, "Connect: "+msg)
		} else {
			lines = append(lines, "Connect: OK ("+cfg.ADBHost+")")
		}
		if out, err := adbDevices(cfg); err != nil {
			lines = append(lines, "\nDevices: ERROR — "+err.Error())
			data.Error = true
		} else {
			lines = append(lines, "\n"+out)
		}
		if model, err := adbGetProp(cfg, "ro.product.model"); err == nil && model != "" {
			lines = append(lines, "\nModel: "+model)
		}
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

	if cfg.ADBHost == "" {
		data.Error = true
		data.Message = "Enter device IP in Connection first."
		render(w, data)
		return
	}

	keycode := r.FormValue("keyevent")
	if keycode == "" {
		keycode = r.FormValue("keyevent_custom")
	}

	var out string
	var err error

	switch r.FormValue("do") {
	case "connect":
		out, err = adbConnect(cfg)
	case "disconnect":
		out, err = adbDisconnect(cfg)
	case "text":
		text := r.FormValue("input_text")
		out, err = adbInputText(cfg, text)
		if err == nil {
			out = "Sent: " + text + "\n" + out
		}
	case "shell":
		cmd := r.FormValue("shell_cmd")
		data.ShellCmd = cmd
		out, err = adbShell(cfg, cmd)
		if err == nil && cmd != "" {
			out = "$ adb shell " + cmd + "\n" + out
		}
	case "terminal":
		cmd := r.FormValue("terminal_cmd")
		data.TerminalCmd = cmd
		out, err = runLocalShell(cmd)
		if err == nil && cmd != "" {
			out = "$ " + cmd + "\n" + out
		}
	case "reboot":
		out, err = adbShell(cfg, "reboot")
	default:
		if keycode != "" {
			out, err = adbKeyEvent(cfg, keycode)
			if err == nil {
				out = fmt.Sprintf("keyevent %s OK\n%s", keycode, out)
			}
		} else {
			err = fmt.Errorf("pick a remote button or enter a keycode")
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
