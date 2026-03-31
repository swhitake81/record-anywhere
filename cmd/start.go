package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/swhitake81/record-anywhere/internal/config"
	"github.com/swhitake81/record-anywhere/internal/process"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start recording system audio in the background",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load and validate config
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return err
		}

		// Check for existing recording
		pid, err := process.ReadPid()
		if err != nil {
			return err
		}
		if pid > 0 {
			if process.IsRunning(pid) {
				return fmt.Errorf("a recording is already in progress (PID %d) — run 'record-anywhere stop' first", pid)
			}
			// Stale PID, clean up
			process.RemovePid()
			process.RemoveStatus()
		}

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		format, _ := cmd.Flags().GetString("format")
		duration, _ := cmd.Flags().GetString("duration")

		if format == "" {
			format = cfg.DefaultFormat
		}
		if duration == "" {
			duration = cfg.DefaultDuration
		}

		// Build output file path
		if name == "" {
			name = time.Now().Format("2006-01-02_15-04-05")
		}

		ext := ".wav"
		if format == "mp3" {
			ext = ".mp3"
		}
		outputPath := filepath.Join(cfg.OutputDir, name+ext)

		// Ensure output directory exists
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		// Spawn background recorder
		childPid, err := process.SpawnRecorder(name, format, duration, outputPath)
		if err != nil {
			return fmt.Errorf("failed to start recording: %w", err)
		}

		fmt.Printf("Recording started (PID %d)\n", childPid)
		fmt.Printf("  Format:   %s\n", format)
		fmt.Printf("  File:     %s\n", outputPath)
		if duration != "0" && duration != "" {
			fmt.Printf("  Duration: %s\n", duration)
		} else {
			fmt.Printf("  Duration: unlimited (run 'record-anywhere stop' to finish)\n")
		}

		return nil
	},
}

func init() {
	startCmd.Flags().String("name", "", "Recording file name (without extension)")
	startCmd.Flags().String("format", "", "Output format: mp3 or wav (default from config)")
	startCmd.Flags().String("duration", "", "Recording duration, e.g. 30m, 1h (default from config)")
	rootCmd.AddCommand(startCmd)
}
