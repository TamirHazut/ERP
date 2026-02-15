package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Checkpoint tracks the last successfully processed position in the WAL.
// Format: [8 bytes: file sequence] [8 bytes: byte offset]
type Checkpoint struct {
	path string     // Path to checkpoint file
	mu   sync.Mutex // Protects file operations
}

const (
	checkpointFileSeqSize = 8 // 8 bytes for file sequence (int64)
	checkpointOffsetSize  = 8 // 8 bytes for byte offset (int64)
	checkpointTotalSize   = checkpointFileSeqSize + checkpointOffsetSize
)

// NewCheckpoint creates a new Checkpoint instance.
func NewCheckpoint(dir string) (*Checkpoint, error) {
	checkpointPath := filepath.Join(dir, "checkpoint.dat")

	return &Checkpoint{
		path: checkpointPath,
	}, nil
}

// Save atomically saves the checkpoint to disk.
// Uses temp file + rename pattern for atomicity.
func (c *Checkpoint) Save(fileSeq, offset int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create temp file in same directory (ensures same filesystem for atomic rename)
	tempPath := c.path + ".tmp"

	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp checkpoint file: %w", err)
	}
	defer tempFile.Close()

	// Write file sequence (8 bytes)
	if err := binary.Write(tempFile, binary.LittleEndian, fileSeq); err != nil {
		os.Remove(tempPath) // Cleanup temp file
		return fmt.Errorf("failed to write file sequence: %w", err)
	}

	// Write byte offset (8 bytes)
	if err := binary.Write(tempFile, binary.LittleEndian, offset); err != nil {
		os.Remove(tempPath) // Cleanup temp file
		return fmt.Errorf("failed to write offset: %w", err)
	}

	// Flush and sync to ensure data reaches disk
	if err := tempFile.Sync(); err != nil {
		os.Remove(tempPath) // Cleanup temp file
		return fmt.Errorf("failed to sync temp checkpoint: %w", err)
	}

	// Close temp file before rename
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath) // Cleanup temp file
		return fmt.Errorf("failed to close temp checkpoint: %w", err)
	}

	// Atomic rename (overwrites existing checkpoint)
	if err := os.Rename(tempPath, c.path); err != nil {
		os.Remove(tempPath) // Cleanup temp file
		return fmt.Errorf("failed to rename temp checkpoint: %w", err)
	}

	return nil
}

// Load reads the checkpoint from disk.
// Returns (fileSeq=1, offset=0) if checkpoint doesn't exist (start from beginning).
func (c *Checkpoint) Load() (int64, int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			// No checkpoint exists, start from beginning
			return 1, 0, nil
		}
		return 0, 0, fmt.Errorf("failed to open checkpoint file: %w", err)
	}
	defer file.Close()

	// Read file sequence (8 bytes)
	var fileSeq int64
	if err := binary.Read(file, binary.LittleEndian, &fileSeq); err != nil {
		if err == io.EOF {
			// Empty checkpoint file, start from beginning
			return 1, 0, nil
		}
		return 0, 0, fmt.Errorf("failed to read file sequence: %w", err)
	}

	// Read byte offset (8 bytes)
	var offset int64
	if err := binary.Read(file, binary.LittleEndian, &offset); err != nil {
		if err == io.EOF {
			// Incomplete checkpoint, start from beginning
			return 1, 0, nil
		}
		return 0, 0, fmt.Errorf("failed to read offset: %w", err)
	}

	return fileSeq, offset, nil
}

// Delete removes the checkpoint file.
// Used for testing or manual reset scenarios.
func (c *Checkpoint) Delete() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.Remove(c.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}

	return nil
}
