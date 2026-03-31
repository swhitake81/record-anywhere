package audio

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

const (
	SampleRate  = 44100
	NumChannels = 2
	BitsPerSample = 16
	wavHeaderSize = 44
)

// WavWriter writes PCM audio data to a WAV file.
// It writes a placeholder header on creation and finalizes it on Close.
type WavWriter struct {
	file      *os.File
	mu        sync.Mutex
	dataSize  uint32
	closed    bool
}

// NewWavWriter creates a new WAV file with a placeholder header.
func NewWavWriter(path string) (*WavWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating wav file: %w", err)
	}

	w := &WavWriter{file: f}
	if err := w.writeHeader(0); err != nil {
		f.Close()
		return nil, err
	}
	return w, nil
}

// Write appends PCM data to the WAV file.
func (w *WavWriter) Write(pcm []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return fmt.Errorf("wav writer is closed")
	}

	n, err := w.file.Write(pcm)
	if err != nil {
		return fmt.Errorf("writing pcm data: %w", err)
	}
	w.dataSize += uint32(n)
	return nil
}

// FlushHeader rewrites the WAV header with the current data size.
// Call periodically for crash safety.
func (w *WavWriter) FlushHeader() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writeHeaderAt(w.dataSize)
}

// Close finalizes the WAV header and closes the file.
func (w *WavWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true

	if err := w.writeHeaderAt(w.dataSize); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}

func (w *WavWriter) writeHeaderAt(dataSize uint32) error {
	if _, err := w.file.Seek(0, 0); err != nil {
		return fmt.Errorf("seeking to header: %w", err)
	}
	if err := w.writeHeader(dataSize); err != nil {
		return err
	}
	// Seek back to end for continued writing
	if _, err := w.file.Seek(0, 2); err != nil {
		return fmt.Errorf("seeking to end: %w", err)
	}
	return nil
}

func (w *WavWriter) writeHeader(dataSize uint32) error {
	byteRate := uint32(SampleRate * NumChannels * BitsPerSample / 8)
	blockAlign := uint16(NumChannels * BitsPerSample / 8)
	fileSize := dataSize + wavHeaderSize - 8

	header := make([]byte, wavHeaderSize)
	copy(header[0:4], "RIFF")
	binary.LittleEndian.PutUint32(header[4:8], fileSize)
	copy(header[8:12], "WAVE")
	copy(header[12:16], "fmt ")
	binary.LittleEndian.PutUint32(header[16:20], 16) // PCM chunk size
	binary.LittleEndian.PutUint16(header[20:22], 1)  // PCM format
	binary.LittleEndian.PutUint16(header[22:24], NumChannels)
	binary.LittleEndian.PutUint32(header[24:28], SampleRate)
	binary.LittleEndian.PutUint32(header[28:32], byteRate)
	binary.LittleEndian.PutUint16(header[32:34], blockAlign)
	binary.LittleEndian.PutUint16(header[34:36], BitsPerSample)
	copy(header[36:40], "data")
	binary.LittleEndian.PutUint32(header[40:44], dataSize)

	_, err := w.file.Write(header)
	return err
}
