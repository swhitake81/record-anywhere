package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/swhitake81/record-anywhere/internal/process"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current recording status",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := process.ReadPid()
		if err != nil {
			return err
		}

		if pid == 0 || !process.IsRunning(pid) {
			// Clean up stale files
			if pid > 0 {
				process.RemovePid()
				process.RemoveStatus()
			}
			fmt.Println("No recording in progress.")
			return nil
		}

		status, err := process.ReadStatus()
		if err != nil {
			return err
		}
		if status == nil {
			fmt.Printf("Recording in progress (PID %d) but no status available.\n", pid)
			return nil
		}

		duration := time.Since(status.StartedAt).Truncate(time.Second)

		fmt.Printf("State:    %s\n", status.State)
		fmt.Printf("PID:      %d\n", pid)
		fmt.Printf("Duration: %s\n", duration)
		fmt.Printf("File:     %s\n", status.FilePath)
		fmt.Printf("Format:   %s\n", status.Format)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
