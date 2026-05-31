package main

import (
	"os"
	"strings"
)

type Config struct {
	ADBHost   string
	ADBSerial string
}

func configFromEnv() Config {
	return Config{
		ADBHost:   os.Getenv("ADB_HOST"),
		ADBSerial: os.Getenv("ADB_SERIAL"),
	}
}

func configFromForm(r map[string]string, fallback Config) Config {
	c := fallback
	if v := strings.TrimSpace(r["device_ip"]); v != "" {
		c.ADBHost = normalizeADBHost(v)
	}
	if v := strings.TrimSpace(r["adb_serial"]); v != "" {
		c.ADBSerial = v
	}
	return c
}

// normalizeADBHost accepts "192.168.1.50" or "192.168.1.50:5555".
func normalizeADBHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		return host
	}
	return host + ":5555"
}
