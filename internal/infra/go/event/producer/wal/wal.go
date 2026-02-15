package wal

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"hash/crc64"
	"io"
	"os"
	"path/filepath"
	"sync"

	eventv1 "erp.localhost/infra/model/event/v1"
	"google.golang.org/protobuf/proto"
)

// WAL represents a Write-Ahead Log for event persistence.
// It provides durable, crash-safe storage of events when both Kafka and MongoDB are unavailable.
type WAL struct {
	dir            string      // Directory containing WAL files
	currentFile    *os.File    // Current active file for writes
	currentSeq     int64       // Sequence number of current file
	maxFileSize    int64       // Maximum file size before rotation (bytes)
	checkpointFile string      // Path to checkpoint file
	writer         *bufio.Writer // Buffered writer for performance
	mu             sync.Mutex  // Protects file operations
	crcTable       *crc64.Table // CRC64 table for checksums
}

const (
	// WAL file naming pattern: events.wal.000001
	walFilePattern = "events.wal.%06d"

	// Entry format:
	// [4 bytes: length] [8 bytes: CRC64] [N bytes: protobuf data]
	entryLengthSize   = 4
	entryChecksumSize = 8
	entryHeaderSize   = entryLengthSize + entryChecksumSize
)

// NewWAL creates a new WAL instance.
func NewWAL(dir string, maxFileSize int64) (*WAL, error) {
	// Create WAL directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	wal := &WAL{
		dir:            dir,
		maxFileSize:    maxFileSize,
		checkpointFile: filepath.Join(dir, "checkpoint.dat"),
		crcTable:       crc64.MakeTable(crc64.ECMA),
	}

	// Find the highest sequence number from existing files
	seq, err := wal.findLatestSequence()
	if err != nil {
		return nil, err
	}

	// Open or create the latest file
	if err := wal.openFile(seq); err != nil {
		return nil, err
	}

	return wal, nil
}

// Append adds a new entry to the WAL.
// This is the primary write path for storing events locally.
func (w *WAL) Append(entry *eventv1.DlqEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Serialize entry to protobuf
	data, err := proto.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ entry: %w", err)
	}

	// Calculate CRC64 checksum
	checksum := crc64.Checksum(data, w.crcTable)

	// Write entry: [length][checksum][data]
	length := uint32(len(data))

	// Write length (4 bytes)
	if err := binary.Write(w.writer, binary.LittleEndian, length); err != nil {
		return fmt.Errorf("failed to write entry length: %w", err)
	}

	// Write checksum (8 bytes)
	if err := binary.Write(w.writer, binary.LittleEndian, checksum); err != nil {
		return fmt.Errorf("failed to write entry checksum: %w", err)
	}

	// Write protobuf data
	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write entry data: %w", err)
	}

	// Flush to ensure data is written to file (still in OS buffer)
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush WAL writer: %w", err)
	}

	// fsync for durability (ensures data reaches disk)
	if err := w.currentFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync WAL file: %w", err)
	}

	// Check if file should be rotated
	stat, err := w.currentFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat WAL file: %w", err)
	}

	if stat.Size() >= w.maxFileSize {
		if err := w.rotateFile(); err != nil {
			return fmt.Errorf("failed to rotate WAL file: %w", err)
		}
	}

	return nil
}

// ReadFrom returns an iterator starting at the given file sequence and offset.
func (w *WAL) ReadFrom(fileSeq, offset int64) (*Iterator, error) {
	filePath := filepath.Join(w.dir, fmt.Sprintf(walFilePattern, fileSeq))

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file for reading: %w", err)
	}

	// Seek to offset
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	return &Iterator{
		file:     file,
		reader:   bufio.NewReader(file),
		offset:   offset,
		fileSeq:  fileSeq,
		crcTable: w.crcTable,
	}, nil
}

// RotateFile closes the current file and creates a new one with an incremented sequence number.
func (w *WAL) rotateFile() error {
	// Close current file
	if w.currentFile != nil {
		if err := w.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer during rotation: %w", err)
		}
		if err := w.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close current file during rotation: %w", err)
		}
	}

	// Increment sequence number
	w.currentSeq++

	// Open new file
	return w.openFile(w.currentSeq)
}

// DeleteFile removes a WAL file by sequence number.
// Should only be called after all entries in the file have been successfully processed.
func (w *WAL) DeleteFile(fileSeq int64) error {
	filePath := filepath.Join(w.dir, fmt.Sprintf(walFilePattern, fileSeq))

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete WAL file: %w", err)
	}

	return nil
}

// GetTotalSize returns the total size of all WAL files in bytes.
func (w *WAL) GetTotalSize() (int64, error) {
	var totalSize int64

	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read WAL directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip checkpoint file
		if entry.Name() == "checkpoint.dat" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		totalSize += info.Size()
	}

	return totalSize, nil
}

// GetFileCount returns the number of WAL files (excluding checkpoint).
func (w *WAL) GetFileCount() (int, error) {
	count := 0

	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read WAL directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip checkpoint file
		if entry.Name() == "checkpoint.dat" {
			continue
		}

		count++
	}

	return count, nil
}

// Close flushes and closes the current WAL file.
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}
	}

	if w.currentFile != nil {
		if err := w.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close current file: %w", err)
		}
	}

	return nil
}

// openFile opens or creates a WAL file with the given sequence number.
func (w *WAL) openFile(seq int64) error {
	filePath := filepath.Join(w.dir, fmt.Sprintf(walFilePattern, seq))

	// Open file in append mode (create if doesn't exist)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file: %w", err)
	}

	w.currentFile = file
	w.currentSeq = seq
	w.writer = bufio.NewWriter(file)

	return nil
}

// findLatestSequence scans the WAL directory and returns the highest sequence number.
// Returns 1 if no files exist (start with events.wal.000001).
func (w *WAL) findLatestSequence() (int64, error) {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil // Start with sequence 1
		}
		return 0, fmt.Errorf("failed to read WAL directory: %w", err)
	}

	var maxSeq int64 = 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Parse sequence number from filename (events.wal.000001)
		var seq int64
		_, err := fmt.Sscanf(entry.Name(), walFilePattern, &seq)
		if err != nil {
			continue // Skip non-WAL files
		}

		if seq > maxSeq {
			maxSeq = seq
		}
	}

	// If no files found, start with sequence 1
	if maxSeq == 0 {
		maxSeq = 1
	}

	return maxSeq, nil
}
