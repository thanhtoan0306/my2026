package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`
	Done      bool               `bson:"done"`
	CreatedAt time.Time          `bson:"created_at"`
}

type PageData struct {
	Todos []Todo
	Error string
}

type App struct {
	collection *mongo.Collection
	templates  *template.Template
}

func main() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect mongodb: %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("ping mongodb: %v (is MongoDB running?)", err)
	}

	dbName := envOr("MONGODB_DB", "todos")
	coll := client.Database(dbName).Collection("items")

	tmpl, err := template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	app := &App{collection: coll, templates: tmpl}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.index)
	mux.HandleFunc("POST /todos", app.create)
	mux.HandleFunc("POST /todos/{id}/toggle", app.toggle)
	mux.HandleFunc("POST /todos/{id}/delete", app.delete)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := envOr("PORT", "8080")
	log.Printf("Todo list: http://127.0.0.1:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cursor, err := a.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		a.render(w, PageData{Error: "Could not load todos"})
		return
	}
	defer cursor.Close(ctx)

	var todos []Todo
	if err := cursor.All(ctx, &todos); err != nil {
		a.render(w, PageData{Error: "Could not load todos"})
		return
	}
	if todos == nil {
		todos = []Todo{}
	}

	a.render(w, PageData{Todos: todos})
}

func (a *App) create(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := a.collection.InsertOne(ctx, Todo{
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		log.Printf("insert todo: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) toggle(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var todo Todo
	if err := a.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&todo); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	_, err = a.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"done": !todo.Done}})
	if err != nil {
		log.Printf("toggle todo: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) delete(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if _, err := a.collection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
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
