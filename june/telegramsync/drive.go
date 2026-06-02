package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func driveConfigDir() string {
	if d := os.Getenv("DRIVE_CONFIG_DIR"); d != "" {
		return d
	}
	return filepath.Join("..", "dbstorageggdrive")
}

func driveService(ctx context.Context) (*drive.Service, error) {
	dir := driveConfigDir()
	b, err := os.ReadFile(filepath.Join(dir, "credentials.json"))
	if err != nil {
		return nil, fmt.Errorf("read credentials in %s: %w", dir, err)
	}
	tokBytes, err := os.ReadFile(filepath.Join(dir, "token.json"))
	if err != nil {
		return nil, fmt.Errorf("read token in %s: %w", dir, err)
	}
	var tok oauth2.Token
	if err := json.Unmarshal(tokBytes, &tok); err != nil {
		return nil, err
	}
	cfg, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, err
	}
	client := cfg.Client(ctx, &tok)
	return drive.NewService(ctx, option.WithHTTPClient(client))
}

func backupFolderID() (string, error) {
	if id := strings.TrimSpace(os.Getenv("DRIVE_TELEGRAM_FOLDER_ID")); id != "" {
		return id, nil
	}
	dir := driveConfigDir()
	b, err := os.ReadFile(filepath.Join(dir, "telegram_backup_folder_id.txt"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func uploadJSON(ctx context.Context, name string, data []byte) (string, error) {
	folderID, err := backupFolderID()
	if err != nil {
		return "", err
	}
	svc, err := driveService(ctx)
	if err != nil {
		return "", err
	}
	f := &drive.File{
		Name:    name,
		Parents: []string{folderID},
	}
	created, err := svc.Files.Create(f).Media(bytes.NewReader(data)).Fields("id,webViewLink").Do()
	if err != nil {
		return "", err
	}
	return "https://drive.google.com/file/d/" + created.Id + "/view", nil
}
