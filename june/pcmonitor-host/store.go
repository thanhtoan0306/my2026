package main

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

type procStat struct {
	PID     int32
	Name    string
	CPU     float64
	RSS     uint64
	MemPct  float32
	Cmdline string
}

type snapshot struct {
	At time.Time

	Hostname string

	Uptime   time.Duration
	BootTime time.Time

	Load1  float64
	Load5  float64
	Load15 float64

	CPUAllPercent float64
	CPUPerCore    []float64

	MemTotal uint64
	MemUsed  uint64
	MemAvail uint64
	MemPct   float64

	SwapTotal uint64
	SwapUsed  uint64
	SwapFree  uint64
	SwapPct   float64

	Procs []procStat
}

type store struct {
	mu   sync.RWMutex
	last snapshot
}

func (s *store) get() snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.last
}

func (s *store) set(v snapshot) {
	s.mu.Lock()
	s.last = v
	s.mu.Unlock()
}

func startSampler(s *store, every time.Duration) {
	go func() {
		_, _ = cpu.Percent(250*time.Millisecond, true)

		t := time.NewTicker(every)
		defer t.Stop()
		for {
			s.set(readSnapshot(every))
			<-t.C
		}
	}()
}

func readSnapshot(interval time.Duration) snapshot {
	out := snapshot{At: time.Now()}

	if hostn, err := host.Info(); err == nil && hostn != nil {
		out.Hostname = hostn.Hostname
	}

	if u, err := host.Uptime(); err == nil {
		out.Uptime = time.Duration(u) * time.Second
		out.BootTime = time.Now().Add(-out.Uptime)
	}

	if la, err := load.Avg(); err == nil && la != nil {
		out.Load1, out.Load5, out.Load15 = la.Load1, la.Load5, la.Load15
	}

	if all, err := cpu.Percent(0, false); err == nil && len(all) > 0 {
		out.CPUAllPercent = all[0]
	}
	if per, err := cpu.Percent(0, true); err == nil && len(per) > 0 {
		out.CPUPerCore = per
	}

	if vm, err := mem.VirtualMemory(); err == nil && vm != nil {
		out.MemTotal, out.MemUsed, out.MemAvail = vm.Total, vm.Used, vm.Available
		out.MemPct = vm.UsedPercent
	}

	if sm, err := mem.SwapMemory(); err == nil && sm != nil {
		out.SwapTotal, out.SwapUsed, out.SwapFree = sm.Total, sm.Used, sm.Free
		out.SwapPct = sm.UsedPercent
	}

	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
	defer cancel()

	if ps, err := process.ProcessesWithContext(ctx); err == nil {
		out.Procs = collectTopProcs(ctx, ps, 30)
	}

	return out
}

func collectTopProcs(ctx context.Context, ps []*process.Process, limit int) []procStat {
	type cand struct {
		p    *process.Process
		cpu  float64
		rss  uint64
		mem  float32
		name string
		cmd  string
	}

	cands := make([]cand, 0, len(ps))
	for _, p := range ps {
		if p == nil {
			continue
		}

		select {
		case <-ctx.Done():
			goto done
		default:
		}

		name, _ := p.NameWithContext(ctx)
		cpuPct, _ := p.CPUPercentWithContext(ctx)
		memPct, _ := p.MemoryPercentWithContext(ctx)
		mi, _ := p.MemoryInfoWithContext(ctx)
		cmd, _ := p.CmdlineWithContext(ctx)

		var rss uint64
		if mi != nil {
			rss = mi.RSS
		}

		cands = append(cands, cand{p: p, cpu: cpuPct, rss: rss, mem: memPct, name: name, cmd: cmd})
	}

done:
	sort.Slice(cands, func(i, j int) bool { return cands[i].cpu > cands[j].cpu })
	if limit > 0 && len(cands) > limit {
		cands = cands[:limit]
	}

	out := make([]procStat, 0, len(cands))
	for _, c := range cands {
		out = append(out, procStat{
			PID:     c.p.Pid,
			Name:    c.name,
			CPU:     c.cpu,
			RSS:     c.rss,
			MemPct:  c.mem,
			Cmdline: c.cmd,
		})
	}
	return out
}

