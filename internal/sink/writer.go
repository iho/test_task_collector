package sink

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// BufferedWriter writes data to a file with buffering and optional encryption.
type BufferedWriter struct {
	file          *os.File
	writer        *bufio.Writer
	bufferSize    int
	flushInterval time.Duration
	mu            sync.Mutex
	done          chan struct{}
	ticker        *time.Ticker

	gcm cipher.AEAD
}

// NewBufferedWriter creates a new BufferedWriter.
func NewBufferedWriter(filename string, keyHex string, bufferSize int, flushInterval time.Duration) (*BufferedWriter, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	bw := &BufferedWriter{
		file:          f,
		writer:        bufio.NewWriterSize(f, bufferSize),
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		done:          make(chan struct{}),
	}

	if keyHex != "" {
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			return nil, fmt.Errorf("invalid hex key: %w", err)
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		bw.gcm = gcm
	}

	bw.startFlushTicker()
	return bw, nil
}

func (bw *BufferedWriter) startFlushTicker() {
	bw.ticker = time.NewTicker(bw.flushInterval)
	go func() {
		for {
			select {
			case <-bw.ticker.C:
				if err := bw.Flush(); err != nil {
					log.Printf("Failed to flush: %v", err)
				}
			case <-bw.done:
				return
			}
		}
	}()
}

func (bw *BufferedWriter) Write(data []byte) error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	toWrite := data
	if bw.gcm != nil {
		nonce := make([]byte, bw.gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		ciphertext := bw.gcm.Seal(nonce, nonce, data, nil)

		encoded := base64.StdEncoding.EncodeToString(ciphertext)
		toWrite = []byte(encoded)
	}

	if _, err := bw.writer.Write(toWrite); err != nil {
		return err
	}

	if _, err := bw.writer.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

// Flush writes any buffered data to the underlying file.
func (bw *BufferedWriter) Flush() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	return bw.writer.Flush()
}

// Close flushes the buffer and closes the file.
func (bw *BufferedWriter) Close() error {
	bw.mu.Lock()
	select {
	case <-bw.done:
		bw.mu.Unlock()
		return nil
	default:
		close(bw.done)
		bw.ticker.Stop()
	}
	bw.mu.Unlock()

	if err := bw.Flush(); err != nil { // Final flush
		log.Printf("Failed to final flush: %v", err)
	}
	return bw.file.Close()
}
