package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// loadEnvFile sets OKX_* from ~/okx-ssr/.env (does not override existing env).
func loadEnvFile() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path := filepath.Join(home, "okx-ssr", ".env")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, val)
		n++
	}
	if n > 0 {
		log.Printf("Loaded %d vars from %s", n, path)
	}
}

func hasEnvCreds() bool {
	return os.Getenv("OKX_API_KEY") != "" &&
		os.Getenv("OKX_SECRET_KEY") != "" &&
		os.Getenv("OKX_PASSPHRASE") != ""
}
