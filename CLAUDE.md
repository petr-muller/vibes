# Claude Instructions

This file contains instructions and context for Claude when working on this project.

## Project Overview

Vibes is a collection of utility scripts and tools for OpenShift development and testing.

## Tools

### fauxinnati

A mock Cincinnati update graph server implementation:
- Located in `cmd/fauxinnati/` (main) and `pkg/fauxinnati/` (business logic)
- Uses Cobra CLI framework
- Implements Cincinnati protocol endpoint: `/api/upgrades_info/graph`
- Currently supports `version-not-found` channel with A→B→C version progression
- No external OpenShift API dependencies (uses local type definitions)

**Build and run commands:**
```bash
go build -o fauxinnati ./cmd/fauxinnati
./fauxinnati --port 8080
```

**Test commands:**
```bash
# Test the API
curl "http://localhost:8080/api/upgrades_info/graph?channel=version-not-found&version=4.17.5&arch=amd64"

# Run unit tests with gotestsum (preferred)
gotestsum ./pkg/fauxinnati

# Run unit tests with standard go test
go test ./pkg/fauxinnati -v
```

## Development Guidelines

- Keep dependencies minimal
- Follow Go best practices
- Use local type definitions instead of external APIs when possible
- Maintain backward compatibility for existing scripts
- Update README.md when adding new tools or significant changes
- Use gotestsum for running tests to get enhanced output and better test reporting
- Run `gofmt -w .` to format Go code before committing

## File Structure

```
.
├── cmd/fauxinnati/          # fauxinnati CLI application
├── pkg/fauxinnati/          # fauxinnati business logic
├── scripts/                 # Utility scripts
├── README.md               # Project documentation
├── CLAUDE.md              # This file
└── go.mod                 # Go module definition
```