# walctl - WAL Debug CLI Tool

A command-line utility for inspecting and debugging Write-Ahead Log (WAL) files used by the event producer.

## Installation

From the repository root:

```bash
# Build walctl
make walctl

# Install to /usr/local/bin (optional)
make walctl-install
```

The binary will be built to: `internal/infra/resources/tools/bin/walctl`

## Usage

### Dump WAL File

Dump all entries from a WAL file as JSON:

```bash
walctl dump data/event_wal/events.wal.000001
```

With pretty formatting:

```bash
walctl dump data/event_wal/events.wal.000001 --pretty
```

### List WAL Files

List all WAL files in a directory with statistics:

```bash
walctl list data/event_wal/
```

Output:
```
FILE                     SEQUENCE  SIZE (MB)  STATUS
────                     ────────  ─────────  ──────
events.wal.000001        1         2.3        Processed
events.wal.000002        2         1.8        Active

Total: 2 files, 4.1 MB (0.00 GB)
```

### Get Specific Entry

Retrieve a specific entry by message ID:

```bash
walctl get data/event_wal/events.wal.000001 --id abc-123-def-456
```

### Version

Show walctl version:

```bash
walctl version
```

### Help

Show help message:

```bash
walctl help
```

## WAL File Format

WAL files use binary protobuf format:
- **Format**: `[4 bytes length][8 bytes CRC64 checksum][N bytes protobuf]`
- **Naming**: `events.wal.NNNNNN` (6-digit sequence number)
- **Checksum**: CRC64 for corruption detection
- **Proto**: Uses `DlqEntry` message from `event/v1/dlq.proto`

## Output Format

### JSON Entry Structure

```json
{
  "message_id": "abc-123-def-456",
  "topic": "user.created",
  "partition_key": "tenant-1:user:user-789",
  "message": { ... },
  "retries": 0,
  "max_retries": 10,
  "next_retry_at": "2026-02-15T10:30:00Z",
  "created_at": "2026-02-15T10:00:00Z",
  "updated_at": "2026-02-15T10:00:00Z",
  "last_error": "",
  "state": "DLQ_ENTRY_STATE_PENDING"
}
```

### Entry States

- `DLQ_ENTRY_STATE_PENDING` (0) - Entry not yet sent
- `DLQ_ENTRY_STATE_SENT` (1) - Entry successfully delivered

## Troubleshooting

### "Invalid WAL filename format"

WAL files must follow the naming convention: `events.wal.NNNNNN`

Example: `events.wal.000001`, `events.wal.000042`

### "Corrupted entry detected"

The tool will skip corrupted entries and continue processing. Corruption is detected via CRC64 checksum verification.

### "Failed to create WAL instance"

Ensure:
1. The WAL directory exists
2. You have read permissions on the directory
3. The directory contains valid WAL files

## Development

### Building from Source

```bash
cd internal/infra/resources/tools/walctl
make build
```

### Running Tests

```bash
cd internal/infra/resources/tools/walctl
go test -v
```

## Dependencies

- Go 1.25+
- `erp.localhost/infra/event/producer/wal` - WAL package
- Protocol Buffers (for DlqEntry deserialization)

## License

Internal tool - Part of ERP system
