package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	credentialsFile  = "credentials.json"
	tokenFile        = "token.json"
	folderIDFile     = "folder_id.txt"
	maxTextPreview   = 64 << 10
)

var driveScopes = []string{drive.DriveScope}

func driveService(ctx context.Context) (*drive.Service, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w (download OAuth client JSON from Google Cloud Console)", credentialsFile, err)
	}
	cfg, err := google.ConfigFromJSON(b, driveScopes...)
	if err != nil {
		return nil, err
	}
	tok, err := loadOrFetchToken(ctx, cfg)
	if err != nil {
		return nil, err
	}
	client := cfg.Client(ctx, tok)
	return drive.NewService(ctx, option.WithHTTPClient(client))
}

func loadOrFetchToken(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	if b, err := os.ReadFile(tokenFile); err == nil {
		var tok oauth2.Token
		if err := json.Unmarshal(b, &tok); err == nil && tok.Valid() {
			return &tok, nil
		}
	}
	tok, err := tokenFromWeb(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := saveToken(tokenFile, tok); err != nil {
		return nil, err
	}
	return tok, nil
}

func tokenFromWeb(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	state := fmt.Sprintf("st-%d", time.Now().UnixNano())
	redirect := "http://localhost:8093/oauth/callback"
	cfg.RedirectURL = redirect
	url := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	mux := http.NewServeMux()
	var code string
	var authErr error
	done := make(chan struct{})

	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			authErr = fmt.Errorf("oauth state mismatch")
			close(done)
			return
		}
		if e := r.URL.Query().Get("error"); e != "" {
			authErr = fmt.Errorf("oauth error: %s", e)
			close(done)
			return
		}
		code = r.URL.Query().Get("code")
		fmt.Fprint(w, "<html><body><h1>Authorized</h1><p>You can close this tab and return to the terminal.</p></body></html>")
		close(done)
	})

	srv := &http.Server{Addr: "localhost:8093", Handler: mux}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()

	log.Printf("Open this URL in your browser to authorize Google Drive:\n%s", url)
	select {
	case <-done:
	case err := <-serveErr:
		if err != nil && err != http.ErrServerClosed {
			return nil, fmt.Errorf("oauth callback server: %w", err)
		}
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)

	if authErr != nil {
		return nil, authErr
	}
	if code == "" {
		return nil, fmt.Errorf("no authorization code received")
	}
	return cfg.Exchange(ctx, code)
}

func saveToken(path string, tok *oauth2.Token) error {
	b, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func loadFolderID() (string, error) {
	b, err := os.ReadFile(folderIDFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func saveFolderID(id string) error {
	return os.WriteFile(folderIDFile, []byte(id+"\n"), 0600)
}

// initPublicFolder creates (or reuses) a Drive folder and grants anyone read access.
func initPublicFolder(ctx context.Context, name string) (string, error) {
	if id, err := loadFolderID(); err == nil && id != "" {
		if err := ensurePublic(ctx, id); err != nil {
			return "", err
		}
		log.Printf("Using existing folder id=%s", id)
		return id, nil
	}

	svc, err := driveService(ctx)
	if err != nil {
		return "", err
	}

	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}
	created, err := svc.Files.Create(folder).Fields("id").Do()
	if err != nil {
		return "", fmt.Errorf("create folder: %w", err)
	}

	if err := ensurePublic(ctx, created.Id); err != nil {
		return "", err
	}
	if err := saveFolderID(created.Id); err != nil {
		return "", err
	}

	log.Printf("Created public folder %q id=%s", name, created.Id)
	log.Printf("Browse: https://drive.google.com/drive/folders/%s", created.Id)
	return created.Id, nil
}

func ensurePublic(ctx context.Context, folderID string) error {
	svc, err := driveService(ctx)
	if err != nil {
		return err
	}
	perms, err := svc.Permissions.List(folderID).Fields("permissions(id,type,role)").Do()
	if err != nil {
		return err
	}
	for _, p := range perms.Permissions {
		if p.Type == "anyone" {
			return nil
		}
	}
	_, err = svc.Permissions.Create(folderID, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Do()
	if err != nil {
		return fmt.Errorf("set public permission: %w", err)
	}
	log.Printf("Folder %s is now visible to anyone with the link", folderID)
	return nil
}

func findSubfolder(ctx context.Context, parentID, name string) (*drive.File, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf("'%s' in parents and name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false",
		parentID, strings.ReplaceAll(name, "'", "\\'"))
	res, err := svc.Files.List().Q(q).Fields("files(id,name)").PageSize(1).Do()
	if err != nil {
		return nil, err
	}
	if len(res.Files) == 0 {
		return nil, nil
	}
	return res.Files[0], nil
}

func createSubfolder(ctx context.Context, parentID, name string) (*drive.File, error) {
	if existing, err := findSubfolder(ctx, parentID, name); err != nil {
		return nil, err
	} else if existing != nil {
		log.Printf("Folder %q already exists id=%s", name, existing.Id)
		return existing, nil
	}

	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	created, err := svc.Files.Create(&drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}).Fields("id,name").Do()
	if err != nil {
		return nil, fmt.Errorf("create subfolder: %w", err)
	}
	if err := ensurePublic(ctx, created.Id); err != nil {
		return nil, err
	}
	log.Printf("Created folder %q id=%s", name, created.Id)
	log.Printf("Open: https://drive.google.com/drive/folders/%s", created.Id)
	return created, nil
}

func uploadFile(ctx context.Context, folderID string, filename string, r io.Reader) (*drive.File, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	f := &drive.File{
		Name:    filepath.Base(filename),
		Parents: []string{folderID},
	}
	created, err := svc.Files.Create(f).Media(r).Fields("id,name,mimeType,size,webViewLink,webContentLink").Do()
	if err != nil {
		return nil, err
	}
	return created, nil
}

func getFileInFolder(ctx context.Context, folderID, fileID string) (*drive.File, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	f, err := svc.Files.Get(fileID).Fields("id,name,mimeType,size,createdTime,parents").Do()
	if err != nil {
		return nil, fmt.Errorf("file not found")
	}
	for _, p := range f.Parents {
		if p == folderID {
			return f, nil
		}
	}
	return nil, fmt.Errorf("file not in storage folder")
}

func downloadFileContent(ctx context.Context, fileID string) ([]byte, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	res, err := svc.Files.Get(fileID).Download()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(io.LimitReader(res.Body, maxTextPreview+1))
}

func listFiles(ctx context.Context, folderID string) ([]*drive.File, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	var out []*drive.File
	pageToken := ""
	for {
		call := svc.Files.List().
			Q(q).
			Fields("nextPageToken, files(id,name,mimeType,size,createdTime,webViewLink)").
			PageSize(100).
			OrderBy("createdTime desc")
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Do()
		if err != nil {
			return nil, err
		}
		out = append(out, res.Files...)
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	return out, nil
}

func publicViewURL(fileID string) string {
	return "https://drive.google.com/file/d/" + fileID + "/view"
}

func publicDirectURL(fileID string) string {
	return "https://drive.google.com/uc?export=view&id=" + fileID
}

func publicDownloadURL(fileID string) string {
	return "https://drive.google.com/uc?export=download&id=" + fileID
}

func isImageMime(m string) bool {
	return strings.HasPrefix(m, "image/")
}
