package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/swhitake81/record-anywhere/internal/setup"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Check and install dependencies (BlackHole, PortAudio, ffmpeg)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking dependencies...")
		fmt.Println()

		allInstalled := true
		for _, dep := range setup.Dependencies {
			installed := setup.CheckDep(dep)
			status := "installed"
			if !installed {
				status = "MISSING"
				allInstalled = false
			}
			fmt.Printf("  %-15s %s\n", dep.Name, status)
		}
		fmt.Println()

		if allInstalled {
			fmt.Println("All dependencies are installed.")
			printMultiOutputGuidance()
			return nil
		}

		fmt.Println("Installing missing dependencies...")
		fmt.Println()

		for _, dep := range setup.Dependencies {
			if setup.CheckDep(dep) {
				continue
			}
			fmt.Printf("  Installing %s...\n", dep.Name)
			if err := setup.InstallDep(dep); err != nil {
				return fmt.Errorf("failed to install %s: %w", dep.Name, err)
			}
			fmt.Printf("  %s installed.\n", dep.Name)
		}

		fmt.Println()
		fmt.Println("All dependencies installed.")

		printMultiOutputGuidance()
		return nil
	},
}

func printMultiOutputGuidance() {
	fmt.Println()
	fmt.Println("IMPORTANT: To hear audio while recording, set up a Multi-Output Device:")
	fmt.Println("  1. Open Audio MIDI Setup (Spotlight → 'Audio MIDI Setup')")
	fmt.Println("  2. Click '+' at bottom-left → 'Create Multi-Output Device'")
	fmt.Println("  3. Check both your speakers/headphones AND your BlackHole device")
	fmt.Println("  4. Set this Multi-Output Device as your system output in System Settings → Sound")
	fmt.Println()
	fmt.Println("If BlackHole doesn't appear after install, you may need to:")
	fmt.Println("  - Approve it in System Settings → Privacy & Security")
	fmt.Println("  - Restart CoreAudio: sudo launchctl kickstart -kp system/com.apple.audio.coreaudiod")
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
