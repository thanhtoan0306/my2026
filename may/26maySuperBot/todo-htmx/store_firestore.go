package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

var appNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

type todoDoc struct {
	Title     string    `firestore:"title"`
	Done      bool      `firestore:"done"`
	CreatedAt time.Time `firestore:"createdAt"`
}

type appDoc struct {
	DisplayName string    `firestore:"displayName"`
	Description string    `firestore:"description"`
	CreatedAt   time.Time `firestore:"createdAt"`
}

type AppRecord struct {
	Name        string
	DisplayName string
	Description string
	CreatedAt   time.Time
}

type FirestoreDB struct {
	client    *firestore.Client
	projectID string
}

func NewFirestoreDB(ctx context.Context, projectID string) (*FirestoreDB, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}
	return &FirestoreDB{client: client, projectID: projectID}, nil
}

func (db *FirestoreDB) AppsCollection() *firestore.CollectionRef {
	return db.client.Collection("apps")
}

func (db *FirestoreDB) TodosCollection(appName string) *firestore.CollectionRef {
	return db.client.Collection("apps").Doc(appName).Collection("todos")
}

func (db *FirestoreDB) NotesCollection(appName string) *firestore.CollectionRef {
	return db.client.Collection("apps").Doc(appName).Collection("notes")
}

func NormalizeAppName(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	name = strings.ReplaceAll(name, " ", "-")
	if !appNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid app name: use lowercase letters, numbers, - or _ (3–63 chars)")
	}
	return name, nil
}

func (db *FirestoreDB) ListApps(ctx context.Context) ([]AppRecord, error) {
	iter := db.AppsCollection().OrderBy("createdAt", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var apps []AppRecord
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var d appDoc
		if err := doc.DataTo(&d); err != nil {
			return nil, err
		}
		apps = append(apps, AppRecord{
			Name:        doc.Ref.ID,
			DisplayName: d.DisplayName,
			Description: d.Description,
			CreatedAt:   d.CreatedAt,
		})
	}
	if apps == nil {
		apps = []AppRecord{}
	}
	return apps, nil
}

func (db *FirestoreDB) GetApp(ctx context.Context, name string) (AppRecord, error) {
	doc, err := db.AppsCollection().Doc(name).Get(ctx)
	if err != nil {
		return AppRecord{}, err
	}
	var d appDoc
	if err := doc.DataTo(&d); err != nil {
		return AppRecord{}, err
	}
	return AppRecord{
		Name:        doc.Ref.ID,
		DisplayName: d.DisplayName,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
	}, nil
}

func (db *FirestoreDB) CreateApp(ctx context.Context, name, displayName, description string) (AppRecord, error) {
	normalized, err := NormalizeAppName(name)
	if err != nil {
		return AppRecord{}, err
	}
	if displayName == "" {
		displayName = normalized
	}
	now := time.Now().UTC()
	record := AppRecord{
		Name:        normalized,
		DisplayName: displayName,
		Description: description,
		CreatedAt:   now,
	}
	_, err = db.AppsCollection().Doc(normalized).Create(ctx, appDoc{
		DisplayName: record.DisplayName,
		Description: record.Description,
		CreatedAt:   record.CreatedAt,
	})
	return record, err
}

func (db *FirestoreDB) ListTodos(ctx context.Context, appName string) ([]Todo, error) {
	iter := db.TodosCollection(appName).OrderBy("createdAt", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var todos []Todo
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var d todoDoc
		if err := doc.DataTo(&d); err != nil {
			return nil, err
		}
		todos = append(todos, Todo{
			ID:        doc.Ref.ID,
			Title:     d.Title,
			Done:      d.Done,
			CreatedAt: d.CreatedAt,
		})
	}
	if todos == nil {
		todos = []Todo{}
	}
	return todos, nil
}

func (db *FirestoreDB) CreateTodo(ctx context.Context, appName, title string) (Todo, error) {
	ref := db.TodosCollection(appName).NewDoc()
	t := Todo{
		ID:        ref.ID,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().UTC(),
	}
	_, err := ref.Set(ctx, todoDoc{Title: t.Title, Done: t.Done, CreatedAt: t.CreatedAt})
	return t, err
}

func (db *FirestoreDB) ToggleTodo(ctx context.Context, appName, id string) (Todo, error) {
	ref := db.TodosCollection(appName).Doc(id)
	doc, err := ref.Get(ctx)
	if err != nil {
		return Todo{}, err
	}
	var d todoDoc
	if err := doc.DataTo(&d); err != nil {
		return Todo{}, err
	}
	d.Done = !d.Done
	if _, err := ref.Set(ctx, d); err != nil {
		return Todo{}, err
	}
	return Todo{ID: id, Title: d.Title, Done: d.Done, CreatedAt: d.CreatedAt}, nil
}

func (db *FirestoreDB) DeleteTodo(ctx context.Context, appName, id string) error {
	_, err := db.TodosCollection(appName).Doc(id).Delete(ctx)
	return err
}

func (db *FirestoreDB) CountTodos(ctx context.Context, appName string) (int, error) {
	iter := db.TodosCollection(appName).Documents(ctx)
	defer iter.Stop()
	n := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			return n, nil
		}
		if err != nil {
			return 0, err
		}
		n++
	}
}
