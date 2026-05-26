package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Todo struct {
	ID        int64
	Title     string
	Done      bool
	CreatedAt time.Time
}

type todoRow struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type PageData struct {
	Todos []Todo
	Error string
}

type App struct {
	client    *supabaseClient
	templates *template.Template
}

type supabaseClient struct {
	baseURL string
	key     string
	http    *http.Client
}

func main() {
	baseURL := strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")
	key := os.Getenv("SUPABASE_KEY")
	if baseURL == "" || key == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_KEY are required. Copy .env.example → .env (Supabase → Connect → API keys)")
	}

	client := &supabaseClient{
		baseURL: baseURL,
		key:     key,
		http:    &http.Client{Timeout: 15 * time.Second},
	}

	if err := client.ping(); err != nil {
		log.Fatalf("supabase: %v\n\nFix: Project Settings → API → copy URL + service_role key (server only)", err)
	}

	tmpl, err := template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	app := &App{client: client, templates: tmpl}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.index)
	mux.HandleFunc("POST /todos", app.create)
	mux.HandleFunc("POST /todos/{id}/toggle", app.toggle)
	mux.HandleFunc("POST /todos/{id}/delete", app.delete)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := envOr("PORT", "8080")
	log.Printf("Supabase todo list: http://127.0.0.1:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func (c *supabaseClient) ping() error {
	_, err := c.listTodos()
	return err
}

func (c *supabaseClient) listTodos() ([]Todo, error) {
	u := c.baseURL + "/rest/v1/items?select=id,title,done,created_at&order=created_at.desc"
	var rows []todoRow
	if err := c.doJSON(http.MethodGet, u, nil, &rows); err != nil {
		return nil, err
	}
	todos := make([]Todo, 0, len(rows))
	for _, r := range rows {
		todos = append(todos, Todo(r))
	}
	return todos, nil
}

func (c *supabaseClient) createTodo(title string) error {
	body, _ := json.Marshal(map[string]any{
		"title": title,
		"done":  false,
	})
	return c.doJSON(http.MethodPost, c.baseURL+"/rest/v1/items", body, nil)
}

func (c *supabaseClient) getTodo(id int64) (Todo, error) {
	u := fmt.Sprintf("%s/rest/v1/items?id=eq.%d&select=id,title,done,created_at", c.baseURL, id)
	var rows []todoRow
	if err := c.doJSON(http.MethodGet, u, nil, &rows); err != nil {
		return Todo{}, err
	}
	if len(rows) == 0 {
		return Todo{}, fmt.Errorf("not found")
	}
	return Todo(rows[0]), nil
}

func (c *supabaseClient) setDone(id int64, done bool) error {
	body, _ := json.Marshal(map[string]bool{"done": done})
	u := fmt.Sprintf("%s/rest/v1/items?id=eq.%d", c.baseURL, id)
	return c.doJSON(http.MethodPatch, u, body, nil)
}

func (c *supabaseClient) deleteTodo(id int64) error {
	u := fmt.Sprintf("%s/rest/v1/items?id=eq.%d", c.baseURL, id)
	return c.doJSON(http.MethodDelete, u, nil, nil)
}

func (c *supabaseClient) doJSON(method, u string, body []byte, out any) error {
	resp, err := c.do(method, u, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	if out != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *supabaseClient) do(method, u string, body []byte) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, u, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.key)
	req.Header.Set("Authorization", "Bearer "+c.key)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=minimal")
	}
	return c.http.Do(req)
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	todos, err := a.client.listTodos()
	if err != nil {
		log.Printf("list todos: %v", err)
		a.render(w, PageData{Error: "Could not load todos (run schema.sql in Supabase SQL Editor?)"})
		return
	}
	if todos == nil {
		todos = []Todo{}
	}
	a.render(w, PageData{Todos: todos})
}

func (a *App) create(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := a.client.createTodo(title); err != nil {
		log.Printf("insert todo: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) toggle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	todo, err := a.client.getTodo(id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := a.client.setDone(id, !todo.Done); err != nil {
		log.Printf("toggle todo: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := a.client.deleteTodo(id); err != nil {
		log.Printf("delete todo: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) render(w http.ResponseWriter, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
