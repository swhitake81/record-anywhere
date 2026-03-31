package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/swhitake81/record-anywhere/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "First-time setup — set output directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)

		// Prompt for output directory
		defaultDir := cfg.OutputDir
		if defaultDir == "" {
			home, _ := os.UserHomeDir()
			defaultDir = filepath.Join(home, "Recordings")
		}

		fmt.Printf("Output directory [%s]: ", defaultDir)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			input = defaultDir
		}

		abs, err := filepath.Abs(input)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(abs, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		cfg.OutputDir = abs

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Config saved. Recordings will be saved to: %s\n", abs)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Print a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		val, err := cfg.Get(args[0])
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.Set(args[0], args[1]); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("%s = %s\n", args[0], args[1])
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
