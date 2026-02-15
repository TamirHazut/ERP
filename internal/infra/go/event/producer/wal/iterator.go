package wal

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc64"
	"io"
	"os"

	eventv1 "erp.localhost/infra/model/event/v1"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrCorrupted indicates a corrupted WAL entry (checksum mismatch).
	ErrCorrupted = errors.New("WAL entry corrupted")

	// ErrEOF indicates the end of the WAL file has been reached.
	ErrEOF = io.EOF
)

// Iterator reads WAL entries sequentially from a file.
type Iterator struct {
	file     *os.File      // Open file handle
	reader   *bufio.Reader // Buffered reader for performance
	offset   int64         // Current byte offset in file
	fileSeq  int64         // File sequence number
	crcTable *crc64.Table  // CRC64 table for checksum verification
}

// Next reads and returns the next DLQ entry from WAL.
// Returns ErrEOF when the end of file is reached.
// Returns ErrCorrupted if the entry checksum doesn't match.
func (i *Iterator) Next() (*eventv1.DlqEntry, int64, error) {
	startOffset := i.offset

	// Read entry length (4 bytes)
	var length uint32
	if err := binary.Read(i.reader, binary.LittleEndian, &length); err != nil {
		if err == io.EOF {
			return nil, startOffset, ErrEOF
		}
		return nil, startOffset, fmt.Errorf("failed to read entry length: %w", err)
	}
	i.offset += entryLengthSize

	// Read checksum (8 bytes)
	var expectedChecksum uint64
	if err := binary.Read(i.reader, binary.LittleEndian, &expectedChecksum); err != nil {
		if err == io.EOF {
			// Incomplete entry (length written but checksum missing)
			return nil, startOffset, ErrEOF
		}
		return nil, startOffset, fmt.Errorf("failed to read entry checksum: %w", err)
	}
	i.offset += entryChecksumSize

	// Read protobuf data
	data := make([]byte, length)
	if _, err := io.ReadFull(i.reader, data); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// Incomplete entry (length/checksum written but data missing)
			return nil, startOffset, ErrEOF
		}
		return nil, startOffset, fmt.Errorf("failed to read entry data: %w", err)
	}
	i.offset += int64(length)

	// Verify checksum
	actualChecksum := crc64.Checksum(data, i.crcTable)
	if actualChecksum != expectedChecksum {
		// Corrupted entry detected
		return nil, startOffset, ErrCorrupted
	}

	// Deserialize protobuf
	entry := &eventv1.DlqEntry{}
	if err := proto.Unmarshal(data, entry); err != nil {
		return nil, startOffset, fmt.Errorf("failed to unmarshal DLQ entry: %w", err)
	}

	// Return entry and offset of NEXT entry (for checkpoint)
	return entry, i.offset, nil
}

// NextN reads up to N entries and returns them.
// Stops early if EOF is reached or an error occurs.
func (i *Iterator) NextN(n int) ([]*eventv1.DlqEntry, int64, error) {
	entries := make([]*eventv1.DlqEntry, 0, n)
	lastOffset := i.offset

	for j := 0; j < n; j++ {
		entry, newOffset, err := i.Next()
		if err != nil {
			if err == ErrEOF {
				// End of file reached, return what we have
				return entries, lastOffset, nil
			}
			if err == ErrCorrupted {
				// Skip corrupted entry and try to continue
				// Log this error in production with appropriate context
				lastOffset = newOffset
				continue
			}
			// Other errors are fatal
			return entries, lastOffset, err
		}

		entries = append(entries, entry)
		lastOffset = newOffset
	}

	return entries, lastOffset, nil
}

// Close closes the underlying file.
func (i *Iterator) Close() error {
	if i.file != nil {
		return i.file.Close()
	}
	return nil
}

// GetFileSeq returns the file sequence number being read.
func (i *Iterator) GetFileSeq() int64 {
	return i.fileSeq
}

// GetOffset returns the current byte offset in the file.
func (i *Iterator) GetOffset() int64 {
	return i.offset
}
