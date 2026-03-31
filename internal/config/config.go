package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const configDirName = "record-anywhere"

type Config struct {
	OutputDir       string `json:"output_dir"`
	DefaultFormat   string `json:"default_format"`
	DefaultDuration string `json:"default_duration"`
}

func DefaultConfig() *Config {
	return &Config{
		OutputDir:       "",
		DefaultFormat:   "mp3",
		DefaultDuration: "0",
	}
}

// ConfigDir returns ~/.config/record-anywhere
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", configDirName), nil
}

func configPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config from disk. Returns default config if file doesn't exist.
func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	p := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	return os.WriteFile(p, data, 0644)
}

// Get returns the value of a config key.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "output_dir":
		return c.OutputDir, nil
	case "default_format":
		return c.DefaultFormat, nil
	case "default_duration":
		return c.DefaultDuration, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// Set sets a config key to a value after validation.
func (c *Config) Set(key, value string) error {
	switch key {
	case "output_dir":
		abs, err := filepath.Abs(value)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
		c.OutputDir = abs
	case "default_format":
		if value != "mp3" && value != "wav" {
			return fmt.Errorf("format must be 'mp3' or 'wav', got '%s'", value)
		}
		c.DefaultFormat = value
	case "default_duration":
		if value != "0" {
			if _, err := time.ParseDuration(value); err != nil {
				return fmt.Errorf("invalid duration '%s': %w", value, err)
			}
		}
		c.DefaultDuration = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// Validate checks the config has required fields.
func (c *Config) Validate() error {
	if c.OutputDir == "" {
		return fmt.Errorf("output_dir not set — run 'record-anywhere config init' first")
	}
	if c.DefaultFormat != "mp3" && c.DefaultFormat != "wav" {
		return fmt.Errorf("default_format must be 'mp3' or 'wav'")
	}
	return nil
}
