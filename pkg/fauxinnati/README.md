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
- `Edge` - Connection between two nodes (represented as `[origin_index, destination_index]`)
- `ConditionalEdge` - Conditional updates with associated risks and edge groups
- `ConditionalUpdateRisk` - Risk information for conditional updates with matching rules
- `MatchingRule` - Rules for risk evaluation (Always, PromQL)
- `PromQLRule` - PromQL-based risk matching configuration

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

- **Unit tests** (`server_test.go`) - Server functionality and all graph generation functions
- **Type tests** (`types_test.go`) - JSON serialization/deserialization for all Cincinnati types
- **Integration tests** (`integration_test.go`) - Full HTTP server testing with httptest for all channels
- **Fixture tests** - Golden file testing with UPDATE=yes support for regression testing
- **TDD workflow** - Test-driven development with iterative refinement

## Graph Generation

The package provides multiple graph generation functions for different channel types:

### Basic Channel Graphs

#### `generateVersionNotFoundGraph`
Creates a three-node graph excluding the requested version:
1. Takes a base version (e.g., `4.17.5`)
2. Generates version A by bumping minor and resetting patch (`4.18.0`)
3. Generates versions B and C by incrementing patch (`4.18.1`, `4.18.2`)
4. Creates unconditional edges A→B→C
5. Adds realistic metadata matching Cincinnati response format

#### `generateChannelHeadGraph`
Creates a three-node graph with the client version as head:
1. Takes a base version (e.g., `4.18.5`)
2. Generates version A by decrementing minor and resetting patch (`4.17.0`)
3. Generates version B by incrementing patch from A (`4.17.1`)
4. Version C is the client's version (`4.18.5`)
5. Creates unconditional edges A→B→C

#### `generateSimpleGraph`
Creates a three-node linear progression from client version:
1. Version A is the client's version (e.g., `4.17.5`)
2. Version B increments patch (`4.17.6`)
3. Version C increments minor (`4.18.0`)
4. Creates unconditional edges A→B→C

### Risk-Based Channel Graphs

#### `generateRisksAlwaysGraph`
Creates a three-node graph with conditional edges using Always matching rules:
1. Version A is the client's version (e.g., `4.17.5`)
2. Version B increments patch (`4.17.6`)
3. Version C increments minor (`4.18.0`)
4. Creates conditional edges A→B and A→C with SyntheticRisk (Always type)
5. Risk always applies, effectively blocking updates

#### `generateRisksMatchingGraph`
Creates a three-node graph with conditional edges using PromQL that matches:
1. Version A is the client's version (e.g., `4.17.5`)
2. Version B increments patch (`4.17.6`)
3. Version C increments minor (`4.18.0`)
4. Creates conditional edges A→B and A→C with SyntheticRisk (PromQL `vector(1)`)
5. PromQL always matches, effectively blocking updates

#### `generateRisksNonmatchingGraph`
Creates a three-node graph with conditional edges using PromQL that doesn't match:
1. Version A is the client's version (e.g., `4.17.5`)
2. Version B increments patch (`4.17.6`)
3. Version C increments minor (`4.18.0`)
4. Creates conditional edges A→B and A→C with SyntheticRisk (PromQL `vector(0)`)
5. PromQL never matches, allowing updates to proceed

### Comprehensive Test Graph

#### `generateSmokeTestGraph`
Creates a complex 13-node graph exercising most Cincinnati features:
1. **13 versions**: D(4.16.0), E(4.17.5), F(4.16.1), G(4.17.6), H(4.18.0), I(4.17.7), J(4.18.1), K(4.17.8), L(4.18.2), M(4.17.9), N(4.18.3), O(4.17.10), P(4.18.4)
2. **4 unconditional edges**: D→E, D→F, E→G, E→H (basic graph structure)
3. **4 conditional edge groups**:
   - RiskA (Always): F→I, G→K with always-matching risk
   - RiskBMatches (PromQL vector(1)): H→J, I→L with always-matching PromQL
   - RiskCNoMatch (PromQL vector(0)): J→N, K→O with never-matching PromQL
   - Combined risks: L→P, M→P with all three risk types in single conditional edge group
4. **Purpose**: Comprehensive testing of graph traversal, risk evaluation, and conditional logic

## Dependencies

- `github.com/blang/semver/v4` - Semantic version parsing and manipulation
- `github.com/spf13/cobra` - CLI framework (used by cmd/fauxinnati)