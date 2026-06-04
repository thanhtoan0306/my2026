package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type apiResult struct {
	OK     bool   `json:"ok"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func okOutput(output string) apiResult {
	return apiResult{OK: true, Output: output}
}

func fail(err error) apiResult {
	return apiResult{OK: false, Error: err.Error()}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("ADB_SIDECAR_PORT")
	if port == "" {
		port = "19527"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/config", handleGetConfig)
	mux.HandleFunc("POST /api/config", handleSetConfig)
	mux.HandleFunc("GET /api/devices", handleDevices)
	mux.HandleFunc("POST /api/connect", handleConnect)
	mux.HandleFunc("POST /api/disconnect", handleDisconnect)
	mux.HandleFunc("POST /api/key", handleKey)
	mux.HandleFunc("POST /api/text", handleText)
	mux.HandleFunc("POST /api/tap", handleTap)
	mux.HandleFunc("POST /api/shell", handleShell)
	mux.HandleFunc("POST /api/terminal", handleTerminal)
	mux.HandleFunc("POST /api/reboot", handleReboot)
	mux.HandleFunc("GET /api/apps", handleApps)
	mux.HandleFunc("POST /api/apps/close-all", handleCloseAll)
	mux.HandleFunc("POST /api/apps/stop", handleStopApp)

	addr := "127.0.0.1:" + port
	log.Printf("adb-sidecar listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "adb": adbBin()})
}

func handleGetConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, getConfig())
}

func handleSetConfig(w http.ResponseWriter, r *http.Request) {
	var c Config
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeJSON(w, http.StatusBadRequest, fail(err))
		return
	}
	c.ADBHost = normalizeADBHost(c.ADBHost)
	setConfig(c)
	writeJSON(w, http.StatusOK, getConfig())
}

func applyConfig(c Config) Config {
	c.ADBHost = normalizeADBHost(c.ADBHost)
	setConfig(c)
	return getConfig()
}

func handleDevices(w http.ResponseWriter, r *http.Request) {
	cfg := getConfig()
	out, err := adbDevices(cfg)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config Config `json:"config"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbConnect(cfg)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleDisconnect(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config Config `json:"config"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbDisconnect(cfg)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config  Config `json:"config"`
		Keycode string `json:"keycode"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbKeyEvent(cfg, body.Keycode)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleText(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config Config `json:"config"`
		Text   string `json:"text"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbInputText(cfg, body.Text)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleTap(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config Config `json:"config"`
		X      string `json:"x"`
		Y      string `json:"y"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbTap(cfg, body.X, body.Y)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleShell(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config  Config `json:"config"`
		Command string `json:"command"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbShell(cfg, body.Command)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleTerminal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config  Config `json:"config"`
		Command string `json:"command"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := runLocalAdb(body.Command, cfg.ADBSerial)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleReboot(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config Config `json:"config"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbReboot(cfg)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleApps(w http.ResponseWriter, r *http.Request) {
	cfg := getConfig()
	apps, err := adbListApps(cfg)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error(), "apps": []AppInfo{}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "apps": apps})
}

func handleCloseAll(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config Config `json:"config"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbCloseAllApps(cfg)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}

func handleStopApp(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config  Config `json:"config"`
		Package string `json:"package"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := applyConfig(body.Config)
	out, err := adbForceStop(cfg, body.Package)
	if err != nil {
		writeJSON(w, http.StatusOK, fail(err))
		return
	}
	writeJSON(w, http.StatusOK, okOutput(out))
}
