package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "record-anywhere",
	Short: "Record system audio on macOS using BlackHole",
	Long:  "A CLI tool that records system audio and saves it to a user-configured folder.\nUses BlackHole (virtual audio driver) + PortAudio for capture and ffmpeg for MP3 encoding.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
