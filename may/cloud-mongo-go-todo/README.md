# Cloud Todo (Go + MongoDB)

Todo list web app that connects to a **remote** MongoDB (e.g. [MongoDB Atlas](https://www.mongodb.com/cloud/atlas)). No Docker.

## Setup

1. Create a free cluster on MongoDB Atlas.
2. Create a database user and allow your IP in **Network Access**.
3. Copy the connection string (Connect → Drivers → Go).

```bash
cp .env.example .env
# Edit .env and set MONGODB_URI
```

## Run

```bash
export $(grep -v '^#' .env | xargs)   # load .env
go mod tidy
go run .
```

Or:

```bash
MONGODB_URI='mongodb+srv://user:pass@cluster.mongodb.net/?retryWrites=true&w=majority' go run .
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080).

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MONGODB_URI` | yes | — | Atlas or any MongoDB connection string |
| `MONGODB_DB` | no | `todos` | Database name |
| `PORT` | no | `8080` | HTTP port |

## Data location

Todos are stored in the cloud MongoDB cluster:

- **Database:** `todos` (or `MONGODB_DB`)
- **Collection:** `items`

Data lives on Atlas (or your provider), not on local disk.
