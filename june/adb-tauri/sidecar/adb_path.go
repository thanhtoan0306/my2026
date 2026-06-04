package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func adbBin() string {
	if p := os.Getenv("ADB_PATH"); p != "" {
		return p
	}
	if p, err := exec.LookPath("adb"); err == nil {
		return p
	}
	home, err := os.UserHomeDir()
	if err == nil {
		candidates := []string{
			filepath.Join(home, "Library/Android/sdk/platform-tools/adb"),
			filepath.Join(home, "Android/Sdk/platform-tools/adb"),
			"/opt/homebrew/bin/adb",
			"/usr/local/bin/adb",
		}
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				return c
			}
		}
	}
	return "adb"
}
