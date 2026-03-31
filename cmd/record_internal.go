package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/spf13/cobra"

	"github.com/swhitake81/record-anywhere/internal/audio"
	"github.com/swhitake81/record-anywhere/internal/process"
)

var recordInternalCmd = &cobra.Command{
	Use:    "_record",
	Short:  "Internal: run the recording loop (used by start command)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		format, _ := cmd.Flags().GetString("format")
		durationStr, _ := cmd.Flags().GetString("duration")

		if filePath == "" {
			return fmt.Errorf("--file is required")
		}

		log.Printf("Starting recording: file=%s format=%s duration=%s", filePath, format, durationStr)

		// Initialize PortAudio
		if err := portaudio.Initialize(); err != nil {
			return fmt.Errorf("initializing PortAudio: %w", err)
		}
		defer portaudio.Terminate()

		// Determine WAV path (always record to WAV first)
		wavPath := filePath
		if format == "mp3" {
			wavPath = strings.TrimSuffix(filePath, ".mp3") + ".wav"
		}

		// Create recorder
		recorder, err := audio.NewRecorder(wavPath)
		if err != nil {
			return fmt.Errorf("creating recorder: %w", err)
		}

		// Write PID and status
		process.WritePid(os.Getpid())
		process.WriteStatus(&process.Status{
			State:     process.StateRecording,
			StartedAt: time.Now(),
			FilePath:  filePath,
			Format:    format,
		})

		// Start recording
		if err := recorder.Start(); err != nil {
			process.RemovePid()
			process.RemoveStatus()
			return fmt.Errorf("starting recording: %w", err)
		}

		log.Println("Recording started, waiting for signal or duration...")

		// Wait for stop signal or duration timeout
		stopCh := make(chan struct{})
		var stopOnce sync.Once
		doStop := func() { stopOnce.Do(func() { close(stopCh) }) }

		// Duration timer
		if durationStr != "" && durationStr != "0" {
			dur, err := time.ParseDuration(durationStr)
			if err == nil && dur > 0 {
				go func() {
					timer := time.NewTimer(dur)
					defer timer.Stop()
					select {
					case <-timer.C:
						doStop()
					case <-stopCh:
					}
				}()
			}
		}

		// Signal handler
		go func() {
			process.WaitForSignal()
			doStop()
		}()

		<-stopCh

		log.Println("Stop signal received, finalizing...")

		// Stop recording
		if err := recorder.Stop(); err != nil {
			log.Printf("Error stopping recorder: %v", err)
		}

		// Convert to MP3 if needed
		if format == "mp3" {
			log.Println("Converting to MP3...")
			process.WriteStatus(&process.Status{
				State:     process.StateConverting,
				StartedAt: time.Now(),
				FilePath:  filePath,
				Format:    format,
			})

			if err := audio.ConvertToMP3(wavPath, filePath); err != nil {
				log.Printf("Error converting to MP3: %v", err)
				// Keep the WAV file as fallback
				process.WriteStatus(&process.Status{
					State:    process.StateStopped,
					FilePath: wavPath,
					Format:   "wav",
				})
			} else {
				log.Println("MP3 conversion complete.")
				process.WriteStatus(&process.Status{
					State:    process.StateStopped,
					FilePath: filePath,
					Format:   format,
				})
			}
		} else {
			process.WriteStatus(&process.Status{
				State:    process.StateStopped,
				FilePath: filePath,
				Format:   format,
			})
		}

		process.RemovePid()
		log.Println("Recording complete.")
		return nil
	},
}

func init() {
	recordInternalCmd.Flags().String("file", "", "Output file path")
	recordInternalCmd.Flags().String("format", "wav", "Output format (wav or mp3)")
	recordInternalCmd.Flags().String("duration", "0", "Recording duration (0 = unlimited)")
	rootCmd.AddCommand(recordInternalCmd)
}
