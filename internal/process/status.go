package process

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/swhitake81/record-anywhere/internal/config"
)

type RecordingState string

const (
	StateRecording  RecordingState = "recording"
	StateConverting RecordingState = "converting"
	StateStopped    RecordingState = "stopped"
)

type Status struct {
	State     RecordingState `json:"state"`
	StartedAt time.Time     `json:"started_at"`
	FilePath  string        `json:"file_path"`
	Format    string        `json:"format"`
}

func statusPath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "status.json"), nil
}

// WriteStatus writes the recording status to disk.
func WriteStatus(s *Status) error {
	p, err := statusPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding status: %w", err)
	}
	return os.WriteFile(p, data, 0644)
}

// ReadStatus reads the recording status from disk.
func ReadStatus() (*Status, error) {
	p, err := statusPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading status: %w", err)
	}

	var s Status
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing status: %w", err)
	}
	return &s, nil
}

// RemoveStatus removes the status file.
func RemoveStatus() error {
	p, err := statusPath()
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
