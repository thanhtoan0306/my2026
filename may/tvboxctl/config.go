package main

import "os"

type Config struct {
	ADBSerial   string
	ADBHost     string
	SSHHost     string
	SSHPort     string
	SSHUser     string
	SSHKey      string
	SSHPassword string
}

func configFromEnv() Config {
	port := os.Getenv("SSH_PORT")
	if port == "" {
		port = "22"
	}
	return Config{
		ADBSerial:   os.Getenv("ADB_SERIAL"),
		ADBHost:     os.Getenv("ADB_HOST"),
		SSHHost:     os.Getenv("SSH_HOST"),
		SSHPort:     port,
		SSHUser:     os.Getenv("SSH_USER"),
		SSHKey:      os.Getenv("SSH_KEY"),
		SSHPassword: os.Getenv("SSH_PASSWORD"),
	}
}

func configFromForm(r map[string]string, fallback Config) Config {
	c := fallback
	if v := r["adb_serial"]; v != "" {
		c.ADBSerial = v
	}
	if v := r["adb_host"]; v != "" {
		c.ADBHost = v
	}
	if v := r["ssh_host"]; v != "" {
		c.SSHHost = v
	}
	if v := r["ssh_port"]; v != "" {
		c.SSHPort = v
	}
	if v := r["ssh_user"]; v != "" {
		c.SSHUser = v
	}
	if v := r["ssh_key"]; v != "" {
		c.SSHKey = v
	}
	if v := r["ssh_password"]; v != "" {
		c.SSHPassword = v
	}
	if c.SSHPort == "" {
		c.SSHPort = "22"
	}
	return c
}
