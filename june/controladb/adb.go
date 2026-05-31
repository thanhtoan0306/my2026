package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func runADB(cfg Config, args ...string) (string, error) {
	base := []string{}
	if cfg.ADBSerial != "" {
		base = append(base, "-s", cfg.ADBSerial)
	}
	base = append(base, args...)
	cmd := exec.Command("adb", base...)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text != "" {
			return text, fmt.Errorf("%w: %s", err, text)
		}
		return "", err
	}
	return text, nil
}

func adbConnect(cfg Config) (string, error) {
	if cfg.ADBHost == "" {
		return "", fmt.Errorf("enter device IP (network debugging must be enabled)")
	}
	return runADB(Config{}, "connect", cfg.ADBHost)
}

func adbDisconnect(cfg Config) (string, error) {
	if cfg.ADBHost == "" {
		return "", fmt.Errorf("enter device IP first")
	}
	return runADB(Config{}, "disconnect", cfg.ADBHost)
}

func adbKeyEvent(cfg Config, keycode string) (string, error) {
	if _, err := adbConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "shell", "input", "keyevent", keycode)
}

func adbInputText(cfg Config, text string) (string, error) {
	if _, err := adbConnect(cfg); err != nil {
		return "", err
	}
	escaped := strings.ReplaceAll(text, " ", "%s")
	return runADB(cfg, "shell", "input", "text", escaped)
}

func adbShell(cfg Config, command string) (string, error) {
	if _, err := adbConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "shell", command)
}

func adbDevices(cfg Config) (string, error) {
	if cfg.ADBHost != "" {
		if _, err := adbConnect(cfg); err != nil {
			return "", err
		}
	}
	return runADB(cfg, "devices", "-l")
}

func adbGetProp(cfg Config, prop string) (string, error) {
	if _, err := adbConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "shell", "getprop", prop)
}
