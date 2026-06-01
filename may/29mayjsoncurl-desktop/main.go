package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/webview/webview_go"
)

//go:embed templates/*
var templateFS embed.FS

const curlTimeout = 30 * time.Second

var (
	curlStart = regexp.MustCompile(`(?i)^\s*curl\b`)
	unsafePat = regexp.MustCompile(`(?i)(?:^|[\s;])(?:rm\s|sudo\s|curl\s+.*\|\s*sh|>\s*/|/dev/tcp)`)
)

type RunResult struct {
	OK     bool
	Status string
	Body   string
	Error  string
}

type PageData struct {
	CurlText        string
	ValidationError string
	Result          *RunResult
}

func main() {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d/", port)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		render(w, tmpl, PageData{})
	})
	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		curlText := r.FormValue("curl")
		if errMsg := validateCurl(curlText); errMsg != "" {
			render(w, tmpl, PageData{CurlText: curlText, ValidationError: errMsg})
			return
		}

		result := runCurl(r.Context(), strings.TrimSpace(curlText))
		render(w, tmpl, PageData{CurlText: curlText, Result: &result})
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("curl → JSON")
	w.SetSize(980, 760, webview.HintNone)
	w.Navigate(url)
	w.Run()
}

func render(w http.ResponseWriter, tmpl *template.Template, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func validateCurl(cmd string) string {
	text := strings.TrimSpace(cmd)
	if text == "" {
		return "Paste a curl command first."
	}
	if !curlStart.MatchString(text) {
		return "Command must start with curl."
	}
	if unsafePat.MatchString(text) {
		return "Command blocked for safety."
	}
	if hasShellChaining(text) {
		return "Shell chaining (&&, ||, ;, $(), etc.) is not allowed."
	}
	return ""
}

func hasShellChaining(text string) bool {
	if strings.Contains(text, "&&") ||
		strings.Contains(text, "||") ||
		strings.Contains(text, "`") ||
		strings.Contains(text, "$(") ||
		strings.Contains(text, "${") ||
		strings.Contains(text, "<(") ||
		strings.Contains(text, ">(") {
		return true
	}
	if idx := strings.LastIndex(text, ";"); idx >= 0 {
		return strings.TrimSpace(text[idx+1:]) != ""
	}
	return false
}

func sortJSONValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sorted := make(map[string]any, len(val))
		for _, k := range keys {
			sorted[k] = sortJSONValue(val[k])
		}
		return sorted
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = sortJSONValue(item)
		}
		return out
	default:
		return val
	}
}

func beautifyBody(raw string) string {
	stripped := strings.TrimSpace(raw)
	if stripped == "" {
		return ""
	}
	var parsed any
	if err := json.Unmarshal([]byte(stripped), &parsed); err != nil {
		return raw
	}
	sorted := sortJSONValue(parsed)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(sorted); err != nil {
		return raw
	}
	return buf.String()
}

func runCurl(ctx context.Context, cmd string) RunResult {
	ctx, cancel := context.WithTimeout(ctx, curlTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, "/bin/bash", "-lc", cmd)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return RunResult{
			OK:     false,
			Status: "timeout",
			Error:  fmt.Sprintf("Timed out after %ds", int(curlTimeout.Seconds())),
		}
	}

	outStr := stdout.String()
	errStr := stderr.String()
	combined := outStr
	if strings.TrimSpace(combined) == "" {
		combined = errStr
	}
	pretty := beautifyBody(combined)

	if err != nil {
		code := -1
		if command.ProcessState != nil {
			code = command.ProcessState.ExitCode()
		}
		if strings.TrimSpace(pretty) == "" {
			msg := strings.TrimSpace(errStr)
			if msg == "" {
				msg = fmt.Sprintf("curl exited with code %d", code)
			}
			return RunResult{
				OK:     false,
				Status: fmt.Sprintf("exit %d", code),
				Body:   pretty,
				Error:  msg,
			}
		}
		res := RunResult{
			OK:     false,
			Status: fmt.Sprintf("exit %d", code),
			Body:   pretty,
		}
		if strings.TrimSpace(errStr) != "" {
			res.Error = strings.TrimSpace(errStr)
		}
		return res
	}

	return RunResult{
		OK:     true,
		Status: "exit 0",
		Body:   pretty,
	}
}
