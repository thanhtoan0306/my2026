package main

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type systemStats struct {
	CPUPercent float64

	MemTotal uint64
	MemUsed  uint64
	MemFree  uint64
	MemUsedPercent float64

	BootTime time.Time
	Uptime   time.Duration
}

func readSystemStats() (systemStats, error) {
	var out systemStats

	c, err := cpu.Percent(250*time.Millisecond, false)
	if err != nil {
		return out, fmt.Errorf("cpu: %w", err)
	}
	if len(c) > 0 {
		out.CPUPercent = c[0]
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return out, fmt.Errorf("mem: %w", err)
	}
	out.MemTotal = vm.Total
	out.MemUsed = vm.Used
	out.MemFree = vm.Available
	out.MemUsedPercent = vm.UsedPercent

	u, err := host.Uptime()
	if err == nil {
		out.Uptime = time.Duration(u) * time.Second
		out.BootTime = time.Now().Add(-out.Uptime)
	}

	return out, nil
}

