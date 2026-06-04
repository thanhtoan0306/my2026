package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
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
	DevicesRaw  string
	Apps        []AppInfo
	TerminalCmd string
	ShellCmd    string
	TapX        string
	TapY        string
	InputText   string
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

func loadDevices(cfg Config, data *PageData) {
	out, err := adbDevices(cfg)
	if err != nil {
		data.DevicesRaw = "Error: " + err.Error()
		return
	}
	data.DevicesRaw = out
	if out == "" {
		data.DevicesRaw = "(no devices)"
	}
}

func loadApps(cfg Config, data *PageData) {
	apps, err := adbListApps(cfg)
	if err != nil {
		data.Error = true
		data.Message = err.Error()
		data.Apps = nil
		return
	}
	data.Apps = apps
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	cfg := getConfig()
	data := PageData{Config: cfg}

	if r.Method == http.MethodPost {
		cfg = mergeConfig(r)
		setConfig(cfg)
		data.Config = cfg
		switch r.FormValue("action") {
		case "save":
			data.Message = "Settings saved."
			loadDevices(cfg, &data)
		case "refresh":
			loadDevices(cfg, &data)
			data.Message = "Device list refreshed."
		default:
			loadDevices(cfg, &data)
		}
	} else {
		loadDevices(cfg, &data)
		loadApps(cfg, &data)
	}

	if isHTMX(r) && r.FormValue("action") == "refresh" {
		_ = renderTemplate(w, "devices", data)
		return
	}
	respond(w, r, data)
}

func handleApps(w http.ResponseWriter, r *http.Request) {
	cfg := mergeConfig(r)
	setConfig(cfg)
	data := PageData{Config: cfg}
	loadApps(cfg, &data)
	if data.Message == "" && len(data.Apps) == 0 && !data.Error {
		data.Message = "No third-party apps (pm list packages -3)."
	}

	if isHTMX(r) {
		_ = renderTemplate(w, "apps", data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleADB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cfg := mergeConfig(r)
	setConfig(cfg)
	data := PageData{Config: cfg}
	data.ShellCmd = r.FormValue("shell_cmd")
	data.TerminalCmd = r.FormValue("terminal_cmd")
	data.InputText = r.FormValue("input_text")
	data.TapX = r.FormValue("tap_x")
	data.TapY = r.FormValue("tap_y")

	keycode := strings.TrimSpace(r.FormValue("keyevent"))
	pkg := strings.TrimSpace(r.FormValue("package"))

	var out string
	var err error
	dualOOB := false

	switch r.FormValue("do") {
	case "connect":
		out, err = adbConnect(cfg)
	case "disconnect":
		out, err = adbDisconnect(cfg)
	case "text":
		out, err = adbInputText(cfg, data.InputText)
		if err == nil && data.InputText != "" {
			out = "Sent: " + data.InputText + "\n" + out
		}
	case "tap":
		if data.TapX == "" || data.TapY == "" {
			err = fmt.Errorf("enter tap X and Y")
		} else {
			out, err = adbTap(cfg, data.TapX, data.TapY)
		}
	case "shell":
		out, err = adbShell(cfg, data.ShellCmd)
		if err == nil && data.ShellCmd != "" {
			out = "$ adb shell " + data.ShellCmd + "\n" + out
		}
	case "terminal":
		cmd := data.TerminalCmd
		if strings.HasPrefix(strings.ToLower(cmd), "adb ") || !strings.Contains(cmd, " ") {
			out, err = runLocalAdb(cmd, cfg.ADBSerial)
		} else {
			out, err = runLocalShell(cmd)
		}
		if err == nil && cmd != "" {
			out = "$ " + cmd + "\n" + out
		}
	case "reboot":
		out, err = adbReboot(cfg)
	case "list_apps":
		loadApps(cfg, &data)
		if err == nil && len(data.Apps) > 0 {
			out = fmt.Sprintf("Found %d apps.", len(data.Apps))
		}
	case "close_all":
		out, err = adbCloseAllApps(cfg)
		loadApps(cfg, &data)
		dualOOB = isHTMX(r)
	case "stop_app":
		if pkg == "" {
			err = fmt.Errorf("missing package name")
		} else {
			out, err = adbForceStop(cfg, pkg)
			loadApps(cfg, &data)
			if err == nil {
				out = "Stopped: " + pkg + "\n" + out
			}
		}
	case "refresh":
		loadDevices(cfg, &data)
		data.Message = data.DevicesRaw
		respond(w, r, data)
		return
	default:
		if keycode != "" {
			out, err = adbKeyEvent(cfg, keycode)
			if err == nil {
				out = fmt.Sprintf("keyevent %s\n%s", keycode, out)
			}
		} else {
			err = fmt.Errorf("pick a remote button or action")
		}
	}

	loadDevices(cfg, &data)

	if err != nil {
		data.Error = true
		data.Message = err.Error()
	} else {
		data.Message = out
		if data.Message == "" {
			data.Message = "OK"
		}
	}

	if dualOOB {
		if err := renderHTMXDual(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	respond(w, r, data)
}
