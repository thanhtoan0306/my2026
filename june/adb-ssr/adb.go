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
	cmd := exec.Command(adbBin(), base...)
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

func maybeConnect(cfg Config) error {
	if cfg.ADBHost == "" {
		return nil
	}
	_, err := runADB(Config{}, "connect", cfg.ADBHost)
	return err
}

func adbConnect(cfg Config) (string, error) {
	if cfg.ADBHost == "" {
		return "", fmt.Errorf("enter device IP (host:port) for network ADB")
	}
	return runADB(Config{}, "connect", cfg.ADBHost)
}

func adbDisconnect(cfg Config) (string, error) {
	if cfg.ADBHost == "" {
		return "", fmt.Errorf("enter device IP first")
	}
	return runADB(Config{}, "disconnect", cfg.ADBHost)
}

func adbDevices(cfg Config) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "devices", "-l")
}

func adbKeyEvent(cfg Config, keycode string) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "shell", "input", "keyevent", keycode)
}

func adbInputText(cfg Config, text string) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	escaped := strings.ReplaceAll(text, " ", "%s")
	return runADB(cfg, "shell", "input", "text", escaped)
}

func adbTap(cfg Config, x, y string) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "shell", "input", "tap", x, y)
}

// adbShell runs a device shell command. Pipes/redirection use sh -c; simple commands use argv form.
func adbShell(cfg Config, command string) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("empty shell command")
	}
	if strings.ContainsAny(command, "|&;<>") {
		return runADB(cfg, "shell", "sh", "-c", command)
	}
	parts := strings.Fields(command)
	return runADB(cfg, append([]string{"shell"}, parts...)...)
}

func adbShellOneLine(cfg Config, command string) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("empty shell command")
	}
	return runADB(cfg, "shell", command)
}

func adbGetProp(cfg Config, prop string) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "shell", "getprop", prop)
}

func adbReboot(cfg Config) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "reboot")
}

func runLocalAdb(command string, serial string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("enter a command")
	}
	if strings.HasPrefix(command, "adb ") {
		command = strings.TrimSpace(command[4:])
	}
	args := strings.Fields(command)
	if len(args) == 0 {
		return "", fmt.Errorf("empty adb command")
	}
	if serial != "" {
		args = append([]string{"-s", serial}, args...)
	}
	cmd := exec.Command(adbBin(), args...)
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
