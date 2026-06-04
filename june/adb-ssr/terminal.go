package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func runLocalShell(command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("enter a terminal command")
	}
	cmd := exec.Command("sh", "-c", command)
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
