package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

// FindBlackHoleDevice finds the BlackHole 2ch input device via PortAudio.
func FindBlackHoleDevice() (*portaudio.DeviceInfo, error) {
	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("enumerating audio devices: %w", err)
	}

	// Prefer BlackHole 2ch, fall back to BlackHole 16ch
	var fallback *portaudio.DeviceInfo
	for _, d := range devices {
		if d.MaxInputChannels < 2 {
			continue
		}
		if strings.Contains(d.Name, "BlackHole 2ch") {
			return d, nil
		}
		if strings.Contains(d.Name, "BlackHole 16ch") {
			fallback = d
		}
	}
	if fallback != nil {
		return fallback, nil
	}

	return nil, fmt.Errorf("BlackHole device not found — run 'record-anywhere setup' to install BlackHole 2ch or 16ch")
}

// Recorder captures audio from BlackHole and writes to a WAV file.
type Recorder struct {
	stream    *portaudio.Stream
	wav       *WavWriter
	stopCh    chan struct{}
	doneCh    chan struct{}
	mu        sync.Mutex
	startTime time.Time
}

// NewRecorder creates a new recorder targeting the given WAV file path.
func NewRecorder(wavPath string) (*Recorder, error) {
	w, err := NewWavWriter(wavPath)
	if err != nil {
		return nil, err
	}

	return &Recorder{
		wav:    w,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}, nil
}

// Start begins capturing audio from BlackHole.
func (r *Recorder) Start() error {
	device, err := FindBlackHoleDevice()
	if err != nil {
		return err
	}

	// Buffer: 1024 frames of stereo float32
	bufferSize := 1024
	buffer := make([]float32, bufferSize*NumChannels)

	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: NumChannels,
			Latency:  device.DefaultLowInputLatency,
		},
		SampleRate:      SampleRate,
		FramesPerBuffer: bufferSize,
	}

	stream, err := portaudio.OpenStream(params, buffer)
	if err != nil {
		return fmt.Errorf("opening audio stream: %w", err)
	}

	if err := stream.Start(); err != nil {
		stream.Close()
		return fmt.Errorf("starting audio stream: %w", err)
	}

	r.stream = stream
	r.startTime = time.Now()

	// Capture loop in goroutine
	go r.captureLoop(buffer)

	return nil
}

// Stop stops the recording and finalizes the WAV file.
func (r *Recorder) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-r.stopCh:
		// Already stopped
		return nil
	default:
		close(r.stopCh)
	}

	// Wait for capture loop to finish
	<-r.doneCh

	var errs []error
	if r.stream != nil {
		if err := r.stream.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("stopping stream: %w", err))
		}
		if err := r.stream.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing stream: %w", err))
		}
	}
	if err := r.wav.Close(); err != nil {
		errs = append(errs, fmt.Errorf("closing wav: %w", err))
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// Duration returns how long the recording has been running.
func (r *Recorder) Duration() time.Duration {
	if r.startTime.IsZero() {
		return 0
	}
	return time.Since(r.startTime)
}

func (r *Recorder) captureLoop(buffer []float32) {
	defer close(r.doneCh)

	// PCM conversion buffer (reusable)
	pcmBuf := make([]byte, len(buffer)*2) // 2 bytes per sample for 16-bit
	flushTicker := time.NewTicker(10 * time.Second)
	defer flushTicker.Stop()

	for {
		select {
		case <-r.stopCh:
			return
		default:
		}

		if err := r.stream.Read(); err != nil {
			// PortAudio may return overflow errors; continue recording
			continue
		}

		// Convert float32 samples to 16-bit signed PCM
		for i, sample := range buffer {
			// Clamp to [-1, 1]
			if sample > 1.0 {
				sample = 1.0
			} else if sample < -1.0 {
				sample = -1.0
			}
			val := int16(sample * math.MaxInt16)
			binary.LittleEndian.PutUint16(pcmBuf[i*2:], uint16(val))
		}

		if err := r.wav.Write(pcmBuf); err != nil {
			return
		}

		// Periodic header flush for crash safety
		select {
		case <-flushTicker.C:
			r.wav.FlushHeader()
		default:
		}
	}
}
