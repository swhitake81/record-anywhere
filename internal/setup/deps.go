package setup

import (
	"fmt"
	"os/exec"
	"strings"
)

type Dependency struct {
	Name        string // display name
	BrewPackage string // homebrew formula/cask name
	CheckCmd    string // command to check existence (e.g. "which ffmpeg")
	IsCask      bool   // install via brew install --cask
}

var Dependencies = []Dependency{
	{
		Name:        "BlackHole",
		BrewPackage: "blackhole-2ch",
		CheckCmd:    "brew list --cask blackhole-2ch 2>/dev/null || brew list --cask blackhole-16ch 2>/dev/null || system_profiler SPAudioDataType 2>/dev/null | grep -q BlackHole",
		IsCask:      true,
	},
	{
		Name:        "PortAudio",
		BrewPackage: "portaudio",
		CheckCmd:    "which portaudio || pkg-config --exists portaudio-2.0",
		IsCask:      false,
	},
	{
		Name:        "ffmpeg",
		BrewPackage: "ffmpeg",
		CheckCmd:    "which ffmpeg",
		IsCask:      false,
	},
}

// CheckDep returns true if the dependency is installed.
func CheckDep(dep Dependency) bool {
	parts := strings.Fields(dep.CheckCmd)
	if len(parts) == 0 {
		return false
	}
	// Handle || in check commands by trying each alternative
	cmd := exec.Command("sh", "-c", dep.CheckCmd)
	err := cmd.Run()
	return err == nil
}

// InstallDep installs a dependency via Homebrew.
func InstallDep(dep Dependency) error {
	args := []string{"install"}
	if dep.IsCask {
		args = []string{"install", "--cask"}
	}
	args = append(args, dep.BrewPackage)

	cmd := exec.Command("brew", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("brew install %s failed: %s\n%s", dep.BrewPackage, err, string(output))
	}
	return nil
}

// CheckAll returns a map of dependency name → installed status.
func CheckAll() map[string]bool {
	result := make(map[string]bool)
	for _, dep := range Dependencies {
		result[dep.Name] = CheckDep(dep)
	}
	return result
}
