package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

var (
	tmpl     *template.Template
	sessions = NewSessionStore()
)

const sessionCookie = "okx_sess"

type PageData struct {
	Env          string
	EnvLabel     string
	Connected    bool
	StatusText   string
	PosCount     int
	Positions    []Position
	Logs         []string
	Message      string
	Error        bool
	ShowEmptyRow bool
}

func initTemplates() error {
	funcs := template.FuncMap{
		"pnlClass": func(pos Position) string {
			if pos.IsPositive() {
				return "pos-positive"
			}
			return "pos-negative"
		},
	}
	var err error
	tmpl, err = template.New("").Funcs(funcs).ParseFS(templateFS, "templates/*.html")
	return err
}

func sessionFromRequest(r *http.Request) (*Session, string) {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return nil, ""
	}
	sess, ok := sessions.Get(c.Value)
	if !ok {
		return nil, ""
	}
	return sess, c.Value
}

func setSessionCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})
}

func demoFromEnv() bool {
	return os.Getenv("OKX_DEMO") == "1"
}

func envLabel(demo bool) string {
	if demo {
		return "Demo"
	}
	return "Real"
}

func pageFromSession(sess *Session) PageData {
	positions, connected, status, logs, demo := sess.snapshot()
	data := PageData{
		Env:         envLabel(demo),
		EnvLabel:    envLabel(demo),
		Connected:   connected,
		StatusText:  status,
		PosCount:    len(positions),
		Positions:   positions,
		Logs:        logs,
	}
	data.ShowEmptyRow = len(positions) == 0
	return data
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func render(w http.ResponseWriter, name string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderFragment(w http.ResponseWriter, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "fragment_live", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderPositionsOnly(w http.ResponseWriter, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "positions_tbody", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sessionFromEnv() (*Session, error) {
	if !hasEnvCreds() {
		return nil, fmt.Errorf("thiếu OKX_API_KEY, OKX_SECRET_KEY hoặc OKX_PASSPHRASE trong env / ~/okx-ssr/.env")
	}
	sess := &Session{
		APIKey:     os.Getenv("OKX_API_KEY"),
		SecretKey:  os.Getenv("OKX_SECRET_KEY"),
		Passphrase: os.Getenv("OKX_PASSPHRASE"),
		Demo:       demoFromEnv(),
		StatusText: "Đang đồng bộ...",
	}
	sess.addLog("REST: bootstrap positions...")
	if err := sess.refreshPositions(); err != nil {
		sess.setDisconnected("Lỗi kết nối")
		sess.addLog("REST lỗi: " + err.Error())
		return sess, err
	}
	positions, _, _, _, _ := sess.snapshot()
	sess.addLog(fmtLog("REST OK — %d vị thế, bật WS...", len(positions)))
	return sess, nil
}

func ensureSession(w http.ResponseWriter, r *http.Request) (*Session, string, PageData) {
	if sess, id := sessionFromRequest(r); sess != nil {
		sess.startWS(id)
		return sess, id, pageFromSession(sess)
	}
	if !hasEnvCreds() {
		return nil, "", PageData{
			Env:          envLabel(demoFromEnv()),
			EnvLabel:     envLabel(demoFromEnv()),
			StatusText:   "Chưa cấu hình",
			ShowEmptyRow: true,
			Logs:         []string{"Tạo ~/okx-ssr/.env với OKX_API_KEY, OKX_SECRET_KEY, OKX_PASSPHRASE"},
			Message:      "Thiếu credentials trong env / .env",
			Error:        true,
		}
	}
	sess, err := sessionFromEnv()
	data := pageFromSession(sess)
	if err != nil {
		data.Error = true
		data.Message = err.Error()
		return sess, "", data
	}
	id := newSessionID()
	sessions.Set(id, sess)
	setSessionCookie(w, id)
	sess.startWS(id)
	return sess, id, data
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, _, data := ensureSession(w, r)
	render(w, "index", data)
}

func handleFragmentLive(w http.ResponseWriter, r *http.Request) {
	sess, _ := sessionFromRequest(r)
	if sess == nil {
		if _, _, data := ensureSession(w, r); !data.Connected {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		sess, _ = sessionFromRequest(r)
	}
	if sess == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// HTMX reads in-memory state updated by OKX private WebSocket.
	renderFragment(w, pageFromSession(sess))
}

func handleClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	sess, id := sessionFromRequest(r)
	if sess == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	instID := r.FormValue("instId")
	mgnMode := r.FormValue("mgnMode")
	sess.addLog("Đang đóng vị thế: " + instID)
	if err := sess.client().ClosePosition(instID, mgnMode); err != nil {
		sess.addLog("Lỗi: " + err.Error())
		sessions.Set(id, sess)
		if isHTMX(r) {
			renderPositionsOnly(w, pageFromSession(sess))
		} else {
			data := pageFromSession(sess)
			data.Message = err.Error()
			render(w, "index", data)
		}
		return
	}
	sess.addLog("Đóng thành công: " + instID)
	_ = sess.refreshPositions()
	sessions.Set(id, sess)
	if isHTMX(r) {
		renderPositionsOnly(w, pageFromSession(sess))
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func fmtLog(format string, args ...any) string {
	return fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}
