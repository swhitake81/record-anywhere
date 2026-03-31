package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/swhitake81/record-anywhere/internal/config"
)

func pidPath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "record.pid"), nil
}

// WritePid writes the current process PID to the PID file.
func WritePid(pid int) error {
	p, err := pidPath()
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte(strconv.Itoa(pid)), 0644)
}

// ReadPid reads the PID from the PID file. Returns 0 if not found.
func ReadPid() (int, error) {
	p, err := pidPath()
	if err != nil {
		return 0, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("reading pid file: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parsing pid: %w", err)
	}
	return pid, nil
}

// RemovePid removes the PID file.
func RemovePid() error {
	p, err := pidPath()
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsRunning checks if a process with the given PID is alive.
func IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}
