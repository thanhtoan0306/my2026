package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type todoDoc struct {
	Title     string    `firestore:"title"`
	Done      bool      `firestore:"done"`
	CreatedAt time.Time `firestore:"createdAt"`
}

type FirestoreStore struct {
	col *firestore.CollectionRef
}

func NewFirestoreStore(ctx context.Context, projectID, appID string) (*FirestoreStore, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}

	col := client.Collection("artifacts").Doc(appID).Collection("public").Doc("data").Collection("todos")
	return &FirestoreStore{col: col}, nil
}

func (s *FirestoreStore) List(ctx context.Context) ([]Todo, error) {
	iter := s.col.OrderBy("createdAt", firestore.Desc).Documents(ctx)
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

func (s *FirestoreStore) Create(ctx context.Context, title string) (Todo, error) {
	ref := s.col.NewDoc()
	t := Todo{
		ID:        ref.ID,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().UTC(),
	}
	_, err := ref.Set(ctx, todoDoc{
		Title:     t.Title,
		Done:      t.Done,
		CreatedAt: t.CreatedAt,
	})
	return t, err
}

func (s *FirestoreStore) Toggle(ctx context.Context, id string) (Todo, error) {
	ref := s.col.Doc(id)
	doc, err := ref.Get(ctx)
	if err != nil {
		return Todo{}, err
	}
	var d todoDoc
	if err := doc.DataTo(&d); err != nil {
		return Todo{}, err
	}
	d.Done = !d.Done
	_, err = ref.Set(ctx, d)
	if err != nil {
		return Todo{}, err
	}
	return Todo{
		ID:        id,
		Title:     d.Title,
		Done:      d.Done,
		CreatedAt: d.CreatedAt,
	}, nil
}

func (s *FirestoreStore) Delete(ctx context.Context, id string) error {
	_, err := s.col.Doc(id).Delete(ctx)
	return err
}
