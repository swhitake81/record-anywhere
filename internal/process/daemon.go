package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/swhitake81/record-anywhere/internal/config"
)

// SpawnRecorder re-invokes the current binary with the hidden _record command
// in a detached background process.
func SpawnRecorder(name, format, duration, outputPath string) (int, error) {
	self, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("finding executable path: %w", err)
	}

	// Build args for the hidden _record command
	args := []string{
		"_record",
		"--file", outputPath,
		"--format", format,
	}
	if duration != "" && duration != "0" {
		args = append(args, "--duration", duration)
	}

	// Open log file for child output
	dir, err := config.ConfigDir()
	if err != nil {
		return 0, err
	}
	logFile, err := os.OpenFile(filepath.Join(dir, "recorder.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return 0, fmt.Errorf("opening log file: %w", err)
	}

	cmd := exec.Command(self, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return 0, fmt.Errorf("spawning recorder process: %w", err)
	}

	pid := cmd.Process.Pid

	// Detach — don't wait for child
	cmd.Process.Release()
	logFile.Close()

	return pid, nil
}
