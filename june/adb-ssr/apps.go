package main

import (
	"fmt"
	"sort"
	"strings"
)

type AppInfo struct {
	Package string
}

func parsePackageList(out string) []AppInfo {
	var apps []AppInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "package:") {
			continue
		}
		pkg := strings.TrimPrefix(line, "package:")
		if pkg == "" {
			continue
		}
		apps = append(apps, AppInfo{Package: pkg})
	}
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Package < apps[j].Package
	})
	return apps
}

func isPmHelp(text string) bool {
	return strings.Contains(text, "Package manager (package) commands")
}

func adbListApps(cfg Config) ([]AppInfo, error) {
	if err := maybeConnect(cfg); err != nil {
		return nil, err
	}

	// TV boxes often break `sh -c`; use argv-style `adb shell pm list packages -3`.
	out, err := runADB(cfg, "shell", "pm", "list", "packages", "-3")
	if err != nil || isPmHelp(out) || !strings.Contains(out, "package:") {
		out, err = adbShellOneLine(cfg, "pm list packages -3")
	}
	if err != nil {
		if isPmHelp(out) {
			return nil, fmt.Errorf("pm list failed on device (connect & authorize device, then Refresh)")
		}
		return nil, err
	}
	if isPmHelp(out) {
		return nil, fmt.Errorf("pm list failed on device (connect & authorize device, then Refresh)")
	}

	apps := parsePackageList(out)
	return apps, nil
}

func adbForceStop(cfg Config, pkg string) (string, error) {
	if err := maybeConnect(cfg); err != nil {
		return "", err
	}
	return runADB(cfg, "shell", "am", "force-stop", pkg)
}

func adbCloseAllApps(cfg Config) (string, error) {
	apps, err := adbListApps(cfg)
	if err != nil {
		return "", err
	}
	if len(apps) == 0 {
		return "No third-party apps found (pm list packages -3).", nil
	}
	var stopped int
	var failed []string
	for _, app := range apps {
		if _, err := adbForceStop(cfg, app.Package); err != nil {
			failed = append(failed, app.Package+": "+err.Error())
			continue
		}
		stopped++
	}
	msg := fmt.Sprintf("Force-stopped %d / %d apps.", stopped, len(apps))
	if len(failed) > 0 {
		msg += "\nFailed:\n" + strings.Join(failed, "\n")
	}
	return msg, nil
}
