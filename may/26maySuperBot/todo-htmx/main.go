package main

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Todo struct {
	ID        string
	Title     string
	Done      bool
	CreatedAt time.Time
}

type PageData struct {
	Todos               []Todo
	Error               string
	ProjectID           string
	AppID               string
	FirestorePath       string
	ConsoleProjectURL   string
	ConsoleFirestoreURL string
}

type TodoStore interface {
	List(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, title string) (Todo, error)
	Toggle(ctx context.Context, id string) (Todo, error)
	Delete(ctx context.Context, id string) error
}

type App struct {
	store     TodoStore
	templates *template.Template
	projectID string
	appID     string
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	store, projectID, appID, err := initFirestoreStore(ctx, root)
	if err != nil {
		log.Fatalf("firebase/firestore: %v", err)
	}

	tmpl, err := template.ParseGlob(filepath.Join(root, "templates", "*.html"))
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	app := &App{store: store, templates: tmpl, projectID: projectID, appID: appID}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.index)
	mux.HandleFunc("POST /todos", app.create)
	mux.HandleFunc("POST /todos/{id}/toggle", app.toggle)
	mux.HandleFunc("DELETE /todos/{id}", app.delete)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(root, "static")))))

	port := envOr("PORT", "8080")
	log.Printf("Todo HTMX + Firestore: http://127.0.0.1:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func (a *App) basePageData() PageData {
	return PageData{
		ProjectID:           a.projectID,
		AppID:               a.appID,
		FirestorePath:       "artifacts/" + a.appID + "/public/data/todos",
		ConsoleProjectURL:   "https://console.firebase.google.com/project/" + a.projectID,
		ConsoleFirestoreURL: "https://console.firebase.google.com/project/" + a.projectID + "/firestore/databases/-default-/data",
	}
}

func initFirestoreStore(ctx context.Context, root string) (*FirestoreStore, string, string, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	appID := os.Getenv("FIREBASE_APP_ID")

	if projectID == "" {
		if id, err := projectIDFromConfig(filepath.Join(root, "config.json")); err == nil {
			projectID = id
		}
	}
	if projectID == "" {
		if id, err := projectIDFromConfig(filepath.Join(root, "..", "firebase-sticky-notes", "config.json")); err == nil {
			projectID = id
		}
	}
	if projectID == "" {
		return nil, "", "", errors.New("set FIREBASE_PROJECT_ID or add todo-htmx/config.json with projectId")
	}
	if appID == "" {
		appID = projectID
	}
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil, "", "", errors.New("set GOOGLE_APPLICATION_CREDENTIALS to your Firebase service account JSON path")
	}

	log.Printf("Firestore: project=%s path=artifacts/%s/public/data/todos", projectID, appID)
	store, err := NewFirestoreStore(ctx, projectID, appID)
	return store, projectID, appID, err
}

type firebaseWebConfig struct {
	ProjectID string `json:"projectId"`
}

func projectIDFromConfig(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var cfg firebaseWebConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return "", err
	}
	if cfg.ProjectID == "" {
		return "", errors.New("projectId missing in config")
	}
	return cfg.ProjectID, nil
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	todos, err := a.store.List(ctx)
	data := a.basePageData()
	if err != nil {
		log.Printf("list todos: %v", err)
		data.Error = "Could not load todos from Firestore"
		a.render(w, "index.html", data)
		return
	}
	data.Todos = todos
	a.render(w, "index.html", data)
}

func (a *App) create(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := r.ParseForm(); err != nil {
		a.respondCreate(w, r, ctx, PageData{Error: "Invalid form"})
		return
	}
	title := r.FormValue("title")
	if title == "" {
		todos, _ := a.store.List(ctx)
		a.respondCreate(w, r, ctx, PageData{Error: "Title is required", Todos: todos})
		return
	}

	if _, err := a.store.Create(ctx, title); err != nil {
		log.Printf("create todo: %v", err)
		todos, _ := a.store.List(ctx)
		a.respondCreate(w, r, ctx, PageData{Error: "Could not save to Firestore", Todos: todos})
		return
	}

	todos, err := a.store.List(ctx)
	if err != nil {
		a.respondCreate(w, r, ctx, PageData{Error: "Saved but could not reload list"})
		return
	}
	a.respondCreate(w, r, ctx, PageData{Todos: todos})
}

func (a *App) respondCreate(w http.ResponseWriter, r *http.Request, ctx context.Context, data PageData) {
	if isHTMX(r) {
		if data.Error != "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
		a.render(w, "todo_list.html", data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) toggle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id := r.PathValue("id")
	todo, err := a.store.Toggle(ctx, id)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.NotFound(w, r)
			return
		}
		log.Printf("toggle todo: %v", err)
		http.Error(w, "Firestore error", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		a.render(w, "todo_item.html", todo)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) delete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id := r.PathValue("id")
	if err := a.store.Delete(ctx, id); err != nil {
		if status.Code(err) == codes.NotFound {
			http.NotFound(w, r)
			return
		}
		log.Printf("delete todo: %v", err)
		http.Error(w, "Firestore error", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
