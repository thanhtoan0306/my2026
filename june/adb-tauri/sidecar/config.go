package main

import (
	"strings"
	"sync"
)

type Config struct {
	ADBHost   string `json:"adb_host"`
	ADBSerial string `json:"adb_serial"`
}

var (
	cfgMu  sync.RWMutex
	appCfg = Config{}
)

func getConfig() Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return appCfg
}

func setConfig(c Config) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	appCfg = c
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
