package internal

import (
	"fmt"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type Snapshot struct {
	CollectedAt time.Time
	CPU         CPUStats
	Memory      MemStats
	Processes   []ProcessInfo
}

type CPUStats struct {
	Total  float64
	PerCPU []float64
}

type MemStats struct {
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

type ProcessInfo struct {
	PID    int32
	Name   string
	CPU    float64
	Memory float32
	MemRSS uint64
	Status string
}

func Collect(topN int) (*Snapshot, error) {
	cpuStats, err := collectCPU()
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}

	memStats, err := collectMemory()
	if err != nil {
		return nil, fmt.Errorf("memory: %w", err)
	}

	procs, err := collectProcesses(topN)
	if err != nil {
		procs = []ProcessInfo{}
	}

	return &Snapshot{
		CollectedAt: time.Now(),
		CPU:         cpuStats,
		Memory:      memStats,
		Processes:   procs,
	}, nil
}

func collectCPU() (CPUStats, error) {
	total, err := cpu.Percent(0, false)
	if err != nil || len(total) == 0 {
		return CPUStats{}, err
	}

	perCPU, err := cpu.Percent(0, true)
	if err != nil {
		perCPU = []float64{}
	}

	return CPUStats{
		Total:  total[0],
		PerCPU: perCPU,
	}, nil
}

func collectMemory() (MemStats, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return MemStats{}, err
	}

	return MemStats{
		Total:       v.Total,
		Used:        v.Used,
		Free:        v.Free,
		UsedPercent: v.UsedPercent,
	}, nil
}

func collectProcesses(topN int) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	infos := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		name, _ := p.Name()
		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()
		memInfo, _ := p.MemoryInfo()
		statuses, _ := p.Status()

		status := "?"
		if len(statuses) > 0 {
			status = statuses[0]
		}

		var rss uint64
		if memInfo != nil {
			rss = memInfo.RSS
		}

		infos = append(infos, ProcessInfo{
			PID:    p.Pid,
			Name:   name,
			CPU:    cpuPct,
			Memory: memPct,
			MemRSS: rss,
			Status: status,
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		if infos[i].CPU != infos[j].CPU {
			return infos[i].CPU > infos[j].CPU
		}
		return infos[i].Memory > infos[j].Memory
	})

	if topN > 0 && len(infos) > topN {
		infos = infos[:topN]
	}

	return infos, nil
}
func BytesToMB(b uint64) float64 {
	return float64(b) / 1024 / 1024
}
func BytesToGB(b uint64) float64 {
	return float64(b) / 1024 / 1024 / 1024
}
