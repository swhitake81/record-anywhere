package audio

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ConvertToMP3 converts a WAV file to MP3 using ffmpeg.
// On success, the original WAV file is deleted.
func ConvertToMP3(wavPath, mp3Path string) error {
	cmd := exec.Command("ffmpeg",
		"-y",           // overwrite output
		"-i", wavPath,  // input
		"-codec:a", "libmp3lame",
		"-qscale:a", "2", // high quality VBR
		mp3Path,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	// Remove the source WAV file
	if err := os.Remove(wavPath); err != nil {
		return fmt.Errorf("removing wav file after conversion: %w", err)
	}

	return nil
}
