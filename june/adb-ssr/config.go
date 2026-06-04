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
		ADBHost:   normalizeADBHost(os.Getenv("ADB_HOST")),
		ADBSerial: strings.TrimSpace(os.Getenv("ADB_SERIAL")),
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
