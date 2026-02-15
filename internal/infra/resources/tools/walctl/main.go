package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"erp.localhost/infra/event/producer/wal"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "dump":
		handleDump()
	case "list":
		handleList()
	case "get":
		handleGet()
	case "version":
		fmt.Printf("walctl version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("walctl - WAL (Write-Ahead Log) Debug Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  walctl dump <wal-file> [--pretty]     Dump WAL file as JSON")
	fmt.Println("  walctl list <wal-dir>                  List all WAL files with stats")
	fmt.Println("  walctl get <wal-file> --id <msg-id>   Get specific entry by message ID")
	fmt.Println("  walctl version                         Show version")
	fmt.Println("  walctl help                            Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  walctl dump data/event_wal/events.wal.000001")
	fmt.Println("  walctl dump data/event_wal/events.wal.000001 --pretty")
	fmt.Println("  walctl list data/event_wal/")
	fmt.Println("  walctl get data/event_wal/events.wal.000001 --id abc-123")
}

func handleDump() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: Missing WAL file path")
		fmt.Fprintln(os.Stderr, "Usage: walctl dump <wal-file> [--pretty]")
		os.Exit(1)
	}

	filePath := os.Args[2]
	pretty := false

	if len(os.Args) > 3 && os.Args[3] == "--pretty" {
		pretty = true
	}

	if err := dumpWALFile(filePath, pretty); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleList() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: Missing WAL directory path")
		fmt.Fprintln(os.Stderr, "Usage: walctl list <wal-dir>")
		os.Exit(1)
	}

	dirPath := os.Args[2]

	if err := listWALFiles(dirPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleGet() {
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, "Error: Missing arguments")
		fmt.Fprintln(os.Stderr, "Usage: walctl get <wal-file> --id <message-id>")
		os.Exit(1)
	}

	filePath := os.Args[2]
	if os.Args[3] != "--id" {
		fmt.Fprintln(os.Stderr, "Error: Expected --id flag")
		os.Exit(1)
	}
	messageID := os.Args[4]

	if err := getEntryByID(filePath, messageID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func dumpWALFile(filePath string, pretty bool) error {
	// Parse sequence number from filename
	var fileSeq int64
	_, err := fmt.Sscanf(filepath.Base(filePath), "events.wal.%06d", &fileSeq)
	if err != nil {
		return fmt.Errorf("invalid WAL filename format: %w", err)
	}

	// Open iterator
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Create temporary WAL instance just for iterator
	tempWAL, err := wal.NewWAL(filepath.Dir(filePath), 100*1024*1024)
	if err != nil {
		return fmt.Errorf("failed to create WAL instance: %w", err)
	}
	defer tempWAL.Close()

	iter, err := tempWAL.ReadFrom(fileSeq, 0)
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	fmt.Printf("# WAL File: %s\n", filepath.Base(filePath))
	fmt.Printf("# File Size: %d bytes (%.2f MB)\n", stat.Size(), float64(stat.Size())/(1024*1024))
	fmt.Printf("# Sequence: %d\n", fileSeq)
	fmt.Println()

	entryCount := 0
	for {
		entry, _, err := iter.Next()
		if err != nil {
			if err == wal.ErrEOF {
				break
			}
			if err == wal.ErrCorrupted {
				fmt.Fprintf(os.Stderr, "Warning: Corrupted entry at offset %d, skipping\n", iter.GetOffset())
				continue
			}
			return fmt.Errorf("failed to read entry: %w", err)
		}

		entryCount++

		var output []byte
		if pretty {
			output, err = json.MarshalIndent(entry, "", "  ")
		} else {
			output, err = json.Marshal(entry)
		}

		if err != nil {
			return fmt.Errorf("failed to marshal entry: %w", err)
		}

		fmt.Println(string(output))
	}

	fmt.Printf("\n# Total entries: %d\n", entryCount)
	return nil
}

func listWALFiles(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILE\tSEQUENCE\tSIZE (MB)\tSTATUS")
	fmt.Fprintln(w, "────\t────────\t─────────\t──────")

	walFiles := 0
	totalSize := int64(0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip checkpoint file
		if name == "checkpoint.dat" {
			continue
		}

		// Parse WAL file sequence
		var seq int64
		_, err := fmt.Sscanf(name, "events.wal.%06d", &seq)
		if err != nil {
			continue // Skip non-WAL files
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		sizeMB := float64(info.Size()) / (1024 * 1024)
		status := "Processed"

		// Check if this is the latest file (active)
		// Simple heuristic: largest sequence number
		entries2, _ := os.ReadDir(dirPath)
		maxSeq := seq
		for _, e2 := range entries2 {
			var s2 int64
			if _, err := fmt.Sscanf(e2.Name(), "events.wal.%06d", &s2); err == nil {
				if s2 > maxSeq {
					maxSeq = s2
				}
			}
		}

		if seq == maxSeq {
			status = "Active"
		}

		fmt.Fprintf(w, "%s\t%d\t%.2f\t%s\n", name, seq, sizeMB, status)
		walFiles++
		totalSize += info.Size()
	}

	w.Flush()

	fmt.Printf("\nTotal: %d files, %.2f MB (%.2f GB)\n",
		walFiles,
		float64(totalSize)/(1024*1024),
		float64(totalSize)/(1024*1024*1024))

	return nil
}

func getEntryByID(filePath string, messageID string) error {
	// Parse sequence number from filename
	var fileSeq int64
	_, err := fmt.Sscanf(filepath.Base(filePath), "events.wal.%06d", &fileSeq)
	if err != nil {
		return fmt.Errorf("invalid WAL filename format: %w", err)
	}

	// Create temporary WAL instance
	tempWAL, err := wal.NewWAL(filepath.Dir(filePath), 100*1024*1024)
	if err != nil {
		return fmt.Errorf("failed to create WAL instance: %w", err)
	}
	defer tempWAL.Close()

	iter, err := tempWAL.ReadFrom(fileSeq, 0)
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Close()

	found := false
	for {
		entry, _, err := iter.Next()
		if err != nil {
			if err == wal.ErrEOF {
				break
			}
			if err == wal.ErrCorrupted {
				continue
			}
			return fmt.Errorf("failed to read entry: %w", err)
		}

		if entry.MessageId == messageID {
			output, err := json.MarshalIndent(entry, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal entry: %w", err)
			}
			fmt.Println(string(output))
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("entry with message_id '%s' not found", messageID)
	}

	return nil
}
