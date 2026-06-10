package main

import (
	"fmt"
	"os"
	"time"

	"Go-Monitor/internal"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	refreshRate int
	topN        int
)

var rootCmd = &cobra.Command{
	Use:   "Go-Monitor",
	Short: "A lightweight system monitor for the terminal",
	Long: `Go-Monitor — real-time system monitor written in Go.
Displays CPU, memory, and top processes in your terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		interval := time.Duration(refreshRate) * time.Millisecond
		if interval < 200*time.Millisecond {
			return fmt.Errorf("refresh rate must be at least 200ms")
		}

		m := internal.NewModel(interval, topN)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running Go-Monitor: %w", err)
		}
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Go-Monitor",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Go-Monitor v0.1.0")
	},
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Print a one-time system snapshot and exit",
	RunE: func(cmd *cobra.Command, args []string) error {
		snap, err := internal.Collect(topN)
		if err != nil {
			return err
		}
		fmt.Println(internal.RenderSnapshot(snap))
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&refreshRate, "refresh", "r", 1000, "Refresh rate in milliseconds")
	rootCmd.PersistentFlags().IntVarP(&topN, "top", "n", 10, "Number of top processes to display")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(snapshotCmd)
}
