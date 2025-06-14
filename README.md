# Vibes

A collection of utility scripts and tools for OpenShift development and testing.

## Tools

### fauxinnati

A mock implementation of the Red Hat OpenShift Cincinnati update graph protocol server.

**Features:**
- Simple Cobra CLI application
- Supports the `version-not-found` channel with A→B→C version progression
- Generates realistic Cincinnati response metadata
- No external OpenShift API dependencies

**Usage:**
```bash
# Build the server
go build -o fauxinnati ./cmd/fauxinnati

# Start server (default port 8080)
./fauxinnati

# Start on custom port
./fauxinnati --port 9090

# Test the API
curl "http://localhost:8080/api/upgrades_info/graph?channel=version-not-found&version=4.17.5&arch=amd64"
```

**API Endpoint:**
- `GET /api/upgrades_info/graph` - Returns update graph based on channel and version parameters

## Scripts

The `scripts/` directory contains various utility scripts:

- **search-ci-failures.fish**: Search for failures in OpenShift CI jobs by scanning build logs from Google Cloud Storage

See [scripts/README.md](scripts/README.md) for detailed documentation.

## Requirements

- Go 1.24+ (for fauxinnati)
- Fish shell (for Fish scripts)
- Google Cloud SDK (`gsutil`)
- `gum` terminal UI library

## Usage

### fauxinnati Server

```bash
go build -o fauxinnati ./cmd/fauxinnati
./fauxinnati --port 8080
```

### Testing

Run tests using gotestsum for enhanced output:

```bash
# Run all tests with enhanced output
gotestsum ./pkg/fauxinnati

# Run tests with verbose output
gotestsum --format testname ./pkg/fauxinnati

# Traditional go test also works
go test ./pkg/fauxinnati -v
```

### Scripts

Scripts can be run directly from the `scripts/` directory:

```bash
./scripts/search-ci-failures.fish openshift/origin pull-ci-openshift-origin-master-e2e-aws 'test failed'
```