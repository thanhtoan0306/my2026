package main

import (
	"context"
	"log"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()
	folderName := os.Getenv("DRIVE_FOLDER_NAME")
	if folderName == "" {
		folderName = "PublicStorage"
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if _, err := initPublicFolder(ctx, folderName); err != nil {
				log.Fatal(err)
			}
			return
		case "upload":
			if len(os.Args) < 3 {
				log.Fatal("usage: go run . upload <file-path>")
			}
			folderID, err := loadFolderID()
			if err != nil || folderID == "" {
				folderID, err = initPublicFolder(ctx, folderName)
				if err != nil {
					log.Fatal(err)
				}
			}
			if err := handleCLIUpload(ctx, folderID, os.Args[2]); err != nil {
				log.Fatal(err)
			}
			return
		case "mkdir":
			if len(os.Args) < 3 {
				log.Fatal("usage: go run . mkdir <folder-name>")
			}
			parentID, err := loadFolderID()
			if err != nil || parentID == "" {
				log.Fatal("run init first or ensure folder_id.txt exists")
			}
			if _, err := createSubfolder(ctx, parentID, os.Args[2]); err != nil {
				log.Fatal(err)
			}
			return
		}
	}

	if err := initTemplates(); err != nil {
		log.Fatal(err)
	}

	folderID, err := loadFolderID()
	if err != nil || folderID == "" {
		folderID, err = initPublicFolder(ctx, folderName)
		if err != nil {
			log.Fatal(err)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8094"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex(folderID, folderName))
	mux.HandleFunc("POST /", handleIndex(folderID, folderName))
	mux.HandleFunc("GET /file/{id}", handleFileView(folderID))

	addr := "127.0.0.1:" + port
	log.Printf("Drive storage SSR: http://%s", addr)
	log.Printf("Public folder: %s", folderURL(folderID))
	log.Fatal(http.ListenAndServe(addr, mux))
}
