package cmd

import (
	"fmt"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/swhitake81/record-anywhere/internal/process"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the current recording",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := process.ReadPid()
		if err != nil {
			return err
		}
		if pid == 0 || !process.IsRunning(pid) {
			// Clean up stale files if any
			process.RemovePid()
			process.RemoveStatus()
			fmt.Println("No recording in progress.")
			return nil
		}

		fmt.Printf("Stopping recording (PID %d)...\n", pid)

		// Send SIGTERM
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			return fmt.Errorf("sending stop signal: %w", err)
		}

		// Poll until process exits or status shows stopped/converting
		timeout := time.After(60 * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				return fmt.Errorf("timed out waiting for recording to stop")
			case <-ticker.C:
				status, _ := process.ReadStatus()
				if status != nil {
					switch status.State {
					case process.StateConverting:
						fmt.Print("\rConverting to MP3...")
						continue
					case process.StateStopped:
						fmt.Println()
						fmt.Printf("Recording saved: %s\n", status.FilePath)
						// Clean up status file after reading final state
						process.RemoveStatus()
						return nil
					}
				}

				// Check if process is still alive
				if !process.IsRunning(pid) {
					// Process exited, read final status
					status, _ := process.ReadStatus()
					if status != nil {
						fmt.Println()
						fmt.Printf("Recording saved: %s\n", status.FilePath)
						process.RemoveStatus()
					} else {
						fmt.Println()
						fmt.Println("Recording stopped.")
					}
					process.RemovePid()
					return nil
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
