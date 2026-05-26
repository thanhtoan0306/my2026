package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

type PageData struct {
	Title        string
	Message      string
	RenderedAt   string
	DockerImage  string
	ProjectDir   string
	OfflineImage string
	OfflineTar   string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dockerImage := os.Getenv("DOCKER_IMAGE")
	if dockerImage == "" {
		dockerImage = "your-dockerhub-user/docker-ssr-hello:latest"
	}

	projectDir := os.Getenv("PROJECT_DIR")
	if projectDir == "" {
		projectDir = "may/26mayDockerSSR"
	}

	offlineImage := os.Getenv("OFFLINE_IMAGE")
	if offlineImage == "" {
		offlineImage = "docker-ssr-hello:offline"
	}

	offlineTar := os.Getenv("OFFLINE_TAR")
	if offlineTar == "" {
		offlineTar = "output/docker-ssr-hello.tar"
	}

	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:        "Docker SSR Hello",
			Message:      "Hello, World!",
			RenderedAt:   time.Now().UTC().Format(time.RFC3339),
			DockerImage:  dockerImage,
			ProjectDir:   projectDir,
			OfflineImage: offlineImage,
			OfflineTar:   offlineTar,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	addr := "0.0.0.0:" + port
	log.Printf("listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}
