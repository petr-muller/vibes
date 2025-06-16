# pkg/fauxinnati

Business logic package for the fauxinnati Cincinnati update graph mock server.

## Package Structure

- `types.go` - Data structures for Cincinnati protocol (Graph, Node, Edge, etc.)
- `server.go` - HTTP server implementation and graph generation logic
- `*_test.go` - Test files for unit, integration, and HTTP testing

## Types

### Core Types

- `Graph` - Complete update graph with nodes, edges, and conditional edges
- `Node` - Individual version with metadata (version, image, metadata)
- `Edge` - Connection between two nodes (represented as `[origin, destination]`)
- `ConditionalEdge` - Conditional updates with associated risks
- `ConditionalUpdateRisk` - Risk information for conditional updates

### Server

- `Server` - HTTP server with Cincinnati API endpoint
- `NewServer()` - Creates new server instance
- `Start(port int)` - Starts HTTP server on specified port

## Testing

The package includes comprehensive tests:

```bash
# Run all tests with gotestsum (preferred)
gotestsum ./pkg/fauxinnati

# Run with different output formats
gotestsum --format testname ./pkg/fauxinnati

# Traditional go test
go test ./pkg/fauxinnati -v
```

### Test Coverage

- **Unit tests** (`server_test.go`) - Server functionality and graph generation
- **Type tests** (`types_test.go`) - JSON serialization/deserialization 
- **Integration tests** (`integration_test.go`) - Full HTTP server testing with httptest

## Graph Generation

The `generateVersionNotFoundGraph` function creates a simple three-node graph:

1. Takes a base version (e.g., `4.17.5`)
2. Generates version A by bumping minor and resetting patch (`4.18.0`)
3. Generates versions B and C by incrementing patch (`4.18.1`, `4.18.2`)
4. Creates edges A→B→C
5. Adds realistic metadata matching Cincinnati response format

## Dependencies

- `github.com/blang/semver/v4` - Semantic version parsing and manipulation
- `github.com/spf13/cobra` - CLI framework (used by cmd/fauxinnati)