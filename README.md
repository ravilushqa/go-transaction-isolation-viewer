# Transaction Isolation Levels Demo

An interactive CLI tool that demonstrates database transaction isolation levels using real database instances. Built with Go, testcontainers, and charmbracelet/bubbletea.

## Features

- 🎓 **Educational**: Learn how transaction isolation levels work through interactive demonstrations
- 🐳 **Real Databases**: Uses testcontainers to spin up actual database instances
- 🎨 **Beautiful TUI**: Interactive terminal UI built with bubbletea and lipgloss
- 🔌 **Extensible**: Plugin architecture for adding more database providers

## Supported Databases

- **MongoDB** - Demonstrates read concern levels and snapshot isolation

## Scenarios

### MongoDB

1. **Dirty Read Prevention** - Shows how transactions prevent reading uncommitted data
2. **Read Committed Isolation** - Demonstrates `readConcern: "majority"` behavior
3. **Snapshot Isolation** - Shows how snapshot isolation provides consistent reads
4. **Write Conflict Detection** - Demonstrates how concurrent write conflicts are handled

## Prerequisites

- Go 1.21+
- Docker (for testcontainers)

## Installation

```bash
# Clone and build
go build -o txviewer ./cmd/txviewer

# Or run directly
go run ./cmd/txviewer
```

## Usage

```bash
# Run the interactive demo
./txviewer

# Or
go run ./cmd/txviewer
```

### Navigation

- `↑/↓` or `j/k` - Navigate menus
- `Enter` - Select item
- `Esc` or `q` - Go back / Quit
- `Ctrl+C` - Force quit (cleans up containers)

## Architecture

```
├── cmd/txviewer/           # Entry point
├── internal/
│   ├── provider/         # Database provider interface
│   │   └── mongodb/      # MongoDB implementation
│   ├── scenario/         # Scenario interface
│   │   └── mongodb/      # MongoDB scenarios
│   └── ui/               # Bubbletea UI components
```

## Adding a New Database Provider

1. Create a new package under `internal/provider/<dbname>/`
2. Implement the `provider.Provider` interface
3. Create scenarios under `internal/scenario/<dbname>/`
4. Register the provider in `cmd/txviewer/main.go`

## License

MIT
