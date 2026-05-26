# Todo List (Go + MongoDB)

A minimal web todo list backed by MongoDB.

## Requirements

- Go 1.22+
- MongoDB (local install or Docker)

## Quick start

Start MongoDB:

```bash
docker compose up -d
```

Run the app (from this directory):

```bash
go mod tidy
go run .
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080).

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGODB_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGODB_DB` | `todos` | Database name |
| `PORT` | `8080` | HTTP port |

## Features

- Add a todo
- Mark done / undone
- Delete a todo
- Todos stored in MongoDB collection `items`
