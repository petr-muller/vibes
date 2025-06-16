# fauxinnati

A mock implementation of the Red Hat OpenShift Cincinnati update graph protocol server.

## Features

- Generates realistic Cincinnati response update graphs
- Responses are generated to provide an update graph relevant for the querying cluster version
- Channel names are used to determine the type of graph to generate

## Usage

### Build and Run

```bash
# Build the server
go build -o fauxinnati ./cmd/fauxinnati

# Start server (default port 8080)
./fauxinnati

# Start on custom port
./fauxinnati --port 9090

# Get help
./fauxinnati --help
```

### API Testing

```bash
# Test the API
curl "http://localhost:8080/api/upgrades_info/graph?channel=version-not-found&version=4.17.5&arch=amd64"
```

## API Endpoint

- `GET /api/upgrades_info/graph` - Returns update graph based on channel and version parameters

### Required Parameters

- `channel` - Update channel (currently supports: `version-not-found`)
- `version` - Base version in semver format (e.g., `4.17.5`)

### Optional Parameters

- `arch` - Architecture (e.g., `amd64`)
