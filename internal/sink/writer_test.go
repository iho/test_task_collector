package sink

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"
)

func TestBufferedWriter(t *testing.T) {
	tmpFile := "test_log.txt"
	defer func() { _ = os.Remove(tmpFile) }()

	bw, err := NewBufferedWriter(tmpFile, "", 100, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	msg := []byte("hello world")
	if err := bw.Write(msg); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if err := bw.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "hello world\n"
	if string(content) != expected {
		t.Errorf("Expected content %q, got %q", expected, string(content))
	}
}

func TestBufferedWriter_Encryption(t *testing.T) {
	tmpFile := "test_enc_log.txt"
	defer func() { _ = os.Remove(tmpFile) }()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to read random key: %v", err)
	}
	keyHex := hex.EncodeToString(key)

	bw, err := NewBufferedWriter(tmpFile, keyHex, 100, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	msg := []byte("secret message")
	if err := bw.Write(msg); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	if err := bw.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) == "secret message\n" {
		t.Error("Content should be encrypted")
	}
}
