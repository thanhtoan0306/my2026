package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func runSSH(cfg Config, command string) (string, error) {
	if cfg.SSHHost == "" || cfg.SSHUser == "" {
		return "", fmt.Errorf("ssh_host and ssh_user are required")
	}

	auth, err := sshAuth(cfg)
	if err != nil {
		return "", err
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(cfg.SSHHost, cfg.SSHPort), &ssh.ClientConfig{
		User:            cfg.SSHUser,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("ssh dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh session: %w", err)
	}
	defer session.Close()

	var buf bytes.Buffer
	session.Stdout = &buf
	session.Stderr = &buf
	if err := session.Run(command); err != nil {
		out := strings.TrimSpace(buf.String())
		if out != "" {
			return out, fmt.Errorf("%w", err)
		}
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func sshAuth(cfg Config) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	if cfg.SSHKey != "" {
		key, err := os.ReadFile(cfg.SSHKey)
		if err != nil {
			return nil, fmt.Errorf("read ssh key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parse ssh key: %w", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if cfg.SSHPassword != "" {
		methods = append(methods, ssh.Password(cfg.SSHPassword))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("set ssh_key path or ssh_password for authentication")
	}
	return methods, nil
}
