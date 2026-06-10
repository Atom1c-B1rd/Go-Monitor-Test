package internal

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	colorGreen  = lipgloss.Color("#00FF87")
	colorYellow = lipgloss.Color("#FFD700")
	colorRed    = lipgloss.Color("#FF5F57")
	colorBlue   = lipgloss.Color("#5AF")
	colorDim    = lipgloss.Color("#555")
	colorWhite  = lipgloss.Color("#EEE")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			PaddingBottom(1)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorDim)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorDim).
			PaddingTop(1)

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(0, 1)
)

type tickMsg time.Time
type snapshotMsg *Snapshot
type errMsg error

type Model struct {
	interval time.Time
	topN     int
	refresh  time.Duration
	snap     *Snapshot
	err      error
	width    int
	height   int
}

func NewModel(refresh time.Duration, topN int) Model {
	return Model{refresh: refresh, topN: topN}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(doCollect(m.topN), tickEvery(m.refresh))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tea.Batch(doCollect(m.topN), tickEvery(m.refresh))

	case snapshotMsg:
		m.snap = msg
		m.err = nil

	case errMsg:
		m.err = msg

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the current state.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n\nPress q to quit.", m.err)
	}
	if m.snap == nil {
		return "Collecting metrics…"
	}

	var sb strings.Builder

	sb.WriteString(styleTitle.Render("⚡ gomon — system monitor") + "\n")
	sb.WriteString(renderCPU(m.snap, m.width) + "\n")
	sb.WriteString(renderMemory(m.snap, m.width) + "\n")
	sb.WriteString(renderProcesses(m.snap, m.width) + "\n")
	sb.WriteString(styleHelp.Render(
		fmt.Sprintf("updated %s   q quit", m.snap.CollectedAt.Format("15:04:05")),
	))

	return sb.String()
}

func renderCPU(snap *Snapshot, width int) string {
	var sb strings.Builder
	sb.WriteString(styleHeader.Render("CPU") + "\n")

	total := snap.CPU.Total
	bar := progressBar(total, 40)
	color := barColor(total)
	sb.WriteString(fmt.Sprintf("  Total  %s %s\n",
		lipgloss.NewStyle().Foreground(color).Render(bar),
		lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%.1f%%", total)),
	))
	cores := snap.CPU.PerCPU
	if len(cores) > 8 {
		cores = cores[:8]
	}
	for i := 0; i < len(cores); i += 2 {
		line := fmt.Sprintf("  cpu%-2d  %s %5.1f%%", i,
			lipgloss.NewStyle().Foreground(barColor(cores[i])).Render(progressBar(cores[i], 16)),
			cores[i],
		)
		if i+1 < len(cores) {
			line += fmt.Sprintf("   cpu%-2d  %s %5.1f%%", i+1,
				lipgloss.NewStyle().Foreground(barColor(cores[i+1])).Render(progressBar(cores[i+1], 16)),
				cores[i+1],
			)
		}
		sb.WriteString(line + "\n")
	}

	return styleBorder.Render(sb.String())
}

func renderMemory(snap *Snapshot, width int) string {
	mem := snap.Memory
	bar := progressBar(mem.UsedPercent, 40)
	color := barColor(mem.UsedPercent)

	content := fmt.Sprintf("%s\n  %s %s\n  Used  %s / %s   Free  %s",
		styleHeader.Render("Memory"),
		lipgloss.NewStyle().Foreground(color).Render(bar),
		lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%.1f%%", mem.UsedPercent)),
		lipgloss.NewStyle().Foreground(colorWhite).Render(fmt.Sprintf("%.1f GB", BytesToGB(mem.Used))),
		fmt.Sprintf("%.1f GB", BytesToGB(mem.Total)),
		fmt.Sprintf("%.1f GB", BytesToGB(mem.Free)),
	)

	return styleBorder.Render(content)
}

func renderProcesses(snap *Snapshot, width int) string {
	var sb strings.Builder

	header := fmt.Sprintf("%-6s  %-20s  %7s  %7s  %9s  %s",
		"PID", "NAME", "CPU%", "MEM%", "RSS(MB)", "STATUS",
	)
	sb.WriteString(styleHeader.Render("Processes") + "\n")
	sb.WriteString(styleHeader.Render(header) + "\n")

	for _, p := range snap.Processes {
		name := p.Name
		if len(name) > 20 {
			name = name[:19] + "…"
		}
		rss := BytesToMB(p.MemRSS)
		line := fmt.Sprintf("%-6d  %-20s  %6.1f%%  %6.1f%%  %8.1f  %s",
			p.PID, name, p.CPU, p.Memory, rss, p.Status,
		)
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(colorWhite).Render(line) + "\n")
	}

	return styleBorder.Render(sb.String())
}

func progressBar(pct float64, width int) string {
	if pct > 100 {
		pct = 100
	}
	filled := int(pct / 100 * float64(width))
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func barColor(pct float64) lipgloss.Color {
	switch {
	case pct >= 80:
		return colorRed
	case pct >= 50:
		return colorYellow
	default:
		return colorGreen
	}
}

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func doCollect(topN int) tea.Cmd {
	return func() tea.Msg {
		snap, err := Collect(topN)
		if err != nil {
			return errMsg(err)
		}
		return snapshotMsg(snap)
	}
}
func RenderSnapshot(snap *Snapshot) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("gomon snapshot — %s\n\n", snap.CollectedAt.Format(time.RFC1123)))
	sb.WriteString(fmt.Sprintf("CPU Total: %.1f%%\n", snap.CPU.Total))
	for i, c := range snap.CPU.PerCPU {
		sb.WriteString(fmt.Sprintf("  cpu%d: %.1f%%\n", i, c))
	}
	sb.WriteString(fmt.Sprintf("\nMemory: %.1f%% used (%.1f / %.1f GB)\n\n",
		snap.Memory.UsedPercent,
		BytesToGB(snap.Memory.Used),
		BytesToGB(snap.Memory.Total),
	))

	sb.WriteString(fmt.Sprintf("%-6s  %-20s  %7s  %7s  %9s\n", "PID", "NAME", "CPU%", "MEM%", "RSS(MB)"))
	for _, p := range snap.Processes {
		name := p.Name
		if len(name) > 20 {
			name = name[:19] + "…"
		}
		sb.WriteString(fmt.Sprintf("%-6d  %-20s  %6.1f%%  %6.1f%%  %8.1f\n",
			p.PID, name, p.CPU, p.Memory, BytesToMB(p.MemRSS),
		))
	}
	return sb.String()
}
