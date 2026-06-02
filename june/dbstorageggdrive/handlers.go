package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
)

//go:embed templates/*
var templateFS embed.FS

var tmpl *template.Template

const maxUpload = 32 << 20 // 32 MiB

type FileRow struct {
	ID            string
	Name          string
	MimeType      string
	SizeHuman     string
	CreatedHuman  string
	ViewURL       string
	DirectURL     string
	DownloadURL   string
	PublicURL     string
	IsImage       bool
}

type IndexData struct {
	FolderID   string
	FolderName string
	FolderURL  string
	Message    string
	Error      bool
	Files      []FileRow
}

type FileData struct {
	ID           string
	Name         string
	MimeType     string
	SizeHuman    string
	CreatedHuman string
	ViewURL      string
	DirectURL    string
	DownloadURL  string
	IsImage      bool
	TextContent  string
}

func initTemplates() error {
	funcMap := template.FuncMap{}
	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	return err
}

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fileToRow(f *drive.File) FileRow {
	row := FileRow{
		ID:           f.Id,
		Name:         f.Name,
		MimeType:     f.MimeType,
		SizeHuman:    formatSize(f.Size),
		CreatedHuman: formatTime(f.CreatedTime),
		ViewURL:      publicViewURL(f.Id),
		DownloadURL:  publicDownloadURL(f.Id),
	}
	if isImageMime(f.MimeType) {
		row.IsImage = true
		row.DirectURL = publicDirectURL(f.Id)
		row.PublicURL = row.DirectURL
	} else {
		row.PublicURL = row.DownloadURL
	}
	return row
}

func filesToRows(files []*drive.File) []FileRow {
	rows := make([]FileRow, 0, len(files))
	for _, f := range files {
		rows = append(rows, fileToRow(f))
	}
	return rows
}

func formatSize(n int64) string {
	if n <= 0 {
		return "—"
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func formatTime(iso string) string {
	if iso == "" {
		return "—"
	}
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Local().Format("2006-01-02 15:04")
}

func folderURL(id string) string {
	return "https://drive.google.com/drive/folders/" + id
}

func handleIndex(folderID, folderName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := IndexData{
			FolderID:   folderID,
			FolderName: folderName,
			FolderURL:  folderURL(folderID),
		}

		if r.Method == http.MethodPost && r.FormValue("action") == "upload" {
			handleUploadPOST(w, r, folderID, folderName, &data)
			return
		}

		files, err := listFiles(r.Context(), folderID)
		if err != nil {
			data.Message = err.Error()
			data.Error = true
			render(w, "index", data)
			return
		}
		data.Files = filesToRows(files)
		render(w, "index", data)
	}
}

func handleUploadPOST(w http.ResponseWriter, r *http.Request, folderID, folderName string, data *IndexData) {
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		data.Message = err.Error()
		data.Error = true
		fillFiles(r.Context(), folderID, data)
		render(w, "index", data)
		return
	}

	fh, hdr, err := r.FormFile("file")
	if err != nil {
		data.Message = "Choose a file to upload."
		data.Error = true
		fillFiles(r.Context(), folderID, data)
		render(w, "index", data)
		return
	}
	defer fh.Close()

	created, err := uploadFile(r.Context(), folderID, hdr.Filename, fh)
	if err != nil {
		log.Printf("upload: %v", err)
		data.Message = err.Error()
		data.Error = true
		fillFiles(r.Context(), folderID, data)
		render(w, "index", data)
		return
	}

	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		row := fileToRow(created)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":          created.Id,
			"name":        created.Name,
			"mimeType":    created.MimeType,
			"viewUrl":     row.ViewURL,
			"directUrl":   row.DirectURL,
			"downloadUrl": row.DownloadURL,
		})
		return
	}

	row := fileToRow(created)
	data.Message = fmt.Sprintf("Uploaded %q — %s", created.Name, row.ViewURL)
	fillFiles(r.Context(), folderID, data)
	render(w, "index", data)
}

func fillFiles(ctx context.Context, folderID string, data *IndexData) {
	files, err := listFiles(ctx, folderID)
	if err != nil {
		return
	}
	data.Files = filesToRows(files)
}

func handleFileView(folderID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/file/")
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		f, err := getFileInFolder(r.Context(), folderID, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		data := FileData{
			ID:           f.Id,
			Name:         f.Name,
			MimeType:     f.MimeType,
			SizeHuman:    formatSize(f.Size),
			CreatedHuman: formatTime(f.CreatedTime),
			ViewURL:      publicViewURL(f.Id),
			DownloadURL:  publicDownloadURL(f.Id),
			IsImage:      isImageMime(f.MimeType),
		}
		if data.IsImage {
			data.DirectURL = publicDirectURL(f.Id)
		}

		if isTextMime(f.MimeType) && f.Size > 0 && f.Size <= maxTextPreview {
			body, err := downloadFileContent(r.Context(), id)
			if err == nil {
				data.TextContent = string(body)
			}
		}

		render(w, "file", data)
	}
}

func isTextMime(m string) bool {
	return strings.HasPrefix(m, "text/") ||
		m == "application/json" ||
		m == "application/xml" ||
		strings.HasSuffix(m, "+json") ||
		strings.HasSuffix(m, "+xml")
}

func handleCLIUpload(ctx context.Context, folderID, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	created, err := uploadFile(ctx, folderID, filepath.Base(path), f)
	if err != nil {
		return err
	}
	log.Printf("Uploaded %s", created.Name)
	log.Printf("  view:     %s", publicViewURL(created.Id))
	if isImageMime(created.MimeType) {
		log.Printf("  direct:   %s", publicDirectURL(created.Id))
	}
	log.Printf("  download: %s", publicDownloadURL(created.Id))
	return nil
}
