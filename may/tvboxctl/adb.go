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
		return "", nil
	}
	return runADB(Config{}, "connect", cfg.ADBHost)
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

// Common TV / Android browser packages — force-stop closes all tabs.
var browserPackages = []string{
	"com.android.chrome",
	"com.google.android.apps.chrome",
	"com.android.browser",
	"com.google.android.tv",
	"org.mozilla.firefox",
	"com.microsoft.emmx",
	"com.opera.browser",
	"com.brave.browser",
	"com.kiwibrowser.browser",
}

func adbKillAllTabs(cfg Config) (string, error) {
	if _, err := adbConnect(cfg); err != nil {
		return "", err
	}
	var lines []string
	for _, pkg := range browserPackages {
		out, err := runADB(cfg, "shell", "am", "force-stop", pkg)
		line := pkg + ": stopped"
		if err != nil {
			if out != "" {
				line = pkg + ": " + out
			} else {
				continue
			}
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return "No known browser packages found (or none were running).", nil
	}
	return strings.Join(lines, "\n"), nil
}

func adbDevices(cfg Config) (string, error) {
	if cfg.ADBHost != "" {
		if msg, err := adbConnect(cfg); err != nil {
			return "", err
		} else if msg != "" {
			_ = msg
		}
	}
	return runADB(cfg, "devices", "-l")
}
