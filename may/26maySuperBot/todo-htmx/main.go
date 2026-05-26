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

type BaseData struct {
	ProjectID           string
	ConsoleProjectURL   string
	ConsoleFirestoreURL string
	ActiveApp           string
	Error               string
}

type TodoPageData struct {
	BaseData
	AppName string
	Todos   []Todo
}

type DashboardData struct {
	BaseData
	Apps []AppRecord
}

type AppPageData struct {
	BaseData
	App           AppRecord
	AppName       string
	Todos         []Todo
	TodoCount     int
	TodosPath     string
	NotesPath     string
	FirestorePath string
}

type TodoItemData struct {
	Todo
	AppName string
}

type Server struct {
	db        *FirestoreDB
	templates *template.Template
	projectID string
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	projectID, err := resolveProjectID(root)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Fatal("set GOOGLE_APPLICATION_CREDENTIALS to your Firebase service account JSON path")
	}

	db, err := NewFirestoreDB(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore: %v", err)
	}

	tmpl, err := template.ParseGlob(filepath.Join(root, "templates", "*.html"))
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	srv := &Server{db: db, templates: tmpl, projectID: projectID}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", srv.rootRedirect)
	mux.HandleFunc("GET /dashboard", srv.dashboard)
	mux.HandleFunc("POST /dashboard/apps", srv.dashboardCreateApp)
	mux.HandleFunc("GET /apps/{projectName}", srv.appPage)
	mux.HandleFunc("POST /apps/{projectName}/todos", srv.appCreateTodo)
	mux.HandleFunc("POST /apps/{projectName}/todos/{id}/toggle", srv.appToggleTodo)
	mux.HandleFunc("DELETE /apps/{projectName}/todos/{id}", srv.appDeleteTodo)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(root, "static")))))

	port := envOr("PORT", "8080")
	log.Printf("Dashboard: http://127.0.0.1:%s/dashboard", port)
	log.Printf("Apps path: apps/{projectName}/{{todos,notes,...}}")
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func resolveProjectID(root string) (string, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
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
		return "", errors.New("set FIREBASE_PROJECT_ID or add config.json with projectId")
	}
	return projectID, nil
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

func (s *Server) baseData() BaseData {
	return BaseData{
		ProjectID:           s.projectID,
		ConsoleProjectURL:   "https://console.firebase.google.com/project/" + s.projectID,
		ConsoleFirestoreURL: "https://console.firebase.google.com/project/" + s.projectID + "/firestore/databases/-default-/data",
	}
}

func (s *Server) rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	apps, err := s.db.ListApps(ctx)
	data := DashboardData{BaseData: s.baseData(), Apps: apps}
	if err != nil {
		log.Printf("list apps: %v", err)
		data.Error = "Could not load apps from Firestore"
	}
	s.render(w, "dashboard.html", data)
}

func (s *Server) dashboardCreateApp(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := r.ParseForm(); err != nil {
		s.dashboardError(w, r, "Invalid form")
		return
	}

	name := r.FormValue("name")
	displayName := r.FormValue("displayName")
	description := r.FormValue("description")

	if _, err := s.db.CreateApp(ctx, name, displayName, description); err != nil {
		log.Printf("create app: %v", err)
		s.dashboardError(w, r, err.Error())
		return
	}

	if isHTMX(r) {
		apps, _ := s.db.ListApps(ctx)
		s.render(w, "app_list.html", DashboardData{BaseData: s.baseData(), Apps: apps})
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (s *Server) dashboardError(w http.ResponseWriter, r *http.Request, msg string) {
	if isHTMX(r) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		apps, _ := s.db.ListApps(r.Context())
		base := s.baseData()
		base.Error = msg
		s.render(w, "app_list.html", DashboardData{BaseData: base, Apps: apps})
		return
	}
	http.Error(w, msg, http.StatusBadRequest)
}

func (s *Server) appPage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	projectName := r.PathValue("projectName")
	app, err := s.db.GetApp(ctx, projectName)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.NotFound(w, r)
			return
		}
		log.Printf("get app: %v", err)
		http.Error(w, "Firestore error", http.StatusInternalServerError)
		return
	}

	todos, err := s.db.ListTodos(ctx, projectName)
	data := s.appPageData(app)
	data.AppName = projectName
	if err != nil {
		log.Printf("list todos: %v", err)
		data.Error = "Could not load todos"
	} else {
		data.Todos = todos
		data.TodoCount = len(todos)
	}
	data.ActiveApp = projectName
	s.render(w, "app.html", data)
}

func (s *Server) appPageData(app AppRecord) AppPageData {
	return AppPageData{
		BaseData:      s.baseData(),
		App:           app,
		TodosPath:     "apps/" + app.Name + "/todos",
		NotesPath:     "apps/" + app.Name + "/notes",
		FirestorePath: "apps/" + app.Name,
	}
}

func (s *Server) appCreateTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	projectName := r.PathValue("projectName")
	if err := r.ParseForm(); err != nil {
		s.appTodoListError(w, r, projectName, "Invalid form")
		return
	}
	title := r.FormValue("title")
	if title == "" {
		s.appTodoListError(w, r, projectName, "Title is required")
		return
	}

	if _, err := s.db.CreateTodo(ctx, projectName, title); err != nil {
		log.Printf("create todo: %v", err)
		s.appTodoListError(w, r, projectName, "Could not save to Firestore")
		return
	}

	todos, err := s.db.ListTodos(ctx, projectName)
	if err != nil {
		s.appTodoListError(w, r, projectName, "Saved but could not reload")
		return
	}
	s.respondAppTodos(w, r, projectName, todos, "")
}

func (s *Server) appToggleTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	projectName := r.PathValue("projectName")
	id := r.PathValue("id")

	todo, err := s.db.ToggleTodo(ctx, projectName, id)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.NotFound(w, r)
			return
		}
		log.Printf("toggle: %v", err)
		http.Error(w, "Firestore error", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		s.render(w, "todo_item.html", TodoItemData{Todo: todo, AppName: projectName})
		return
	}
	http.Redirect(w, r, "/apps/"+projectName, http.StatusSeeOther)
}

func (s *Server) appDeleteTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	projectName := r.PathValue("projectName")
	id := r.PathValue("id")

	if err := s.db.DeleteTodo(ctx, projectName, id); err != nil {
		if status.Code(err) == codes.NotFound {
			http.NotFound(w, r)
			return
		}
		log.Printf("delete: %v", err)
		http.Error(w, "Firestore error", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/apps/"+projectName, http.StatusSeeOther)
}

func (s *Server) appTodoListError(w http.ResponseWriter, r *http.Request, projectName, msg string) {
	todos, _ := s.db.ListTodos(r.Context(), projectName)
	s.respondAppTodos(w, r, projectName, todos, msg)
}

func (s *Server) respondAppTodos(w http.ResponseWriter, r *http.Request, projectName string, todos []Todo, errMsg string) {
	base := s.baseData()
	base.Error = errMsg
	data := TodoPageData{BaseData: base, AppName: projectName, Todos: todos}
	if isHTMX(r) {
		if errMsg != "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
		s.render(w, "todo_list.html", data)
		return
	}
	http.Redirect(w, r, "/apps/"+projectName, http.StatusSeeOther)
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
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
