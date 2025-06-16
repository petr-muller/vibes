# Claude Instructions for pkg/fauxinnati

## Package Overview
The `pkg/fauxinnati` package implements the business logic for a mock Cincinnati update graph server. It provides realistic Cincinnati protocol responses for OpenShift update graph queries.

## Architecture & Key Concepts

### Core Types (types.go)
- **Graph**: Complete Cincinnati update graph with nodes, edges, and conditional edges
- **Node**: Individual OpenShift version with payload image and metadata  
- **Edge**: Connection between nodes represented as `[origin_index, destination_index]`
- **ConditionalEdge**: Advanced update paths with risk information
- **Server**: HTTP server handling Cincinnati API endpoints

### Server Implementation (server.go)
- Single endpoint: `GET /api/upgrades_info/graph`
- Channel-based graph generation (currently supports `version-not-found`)
- Version-based graph generation using semver parsing

## Key Implementation Details

### Version Generation Logic
The `generateVersionNotFoundGraph` function creates a three-node linear graph:
1. **Version A**: Increments minor version, resets patch (`4.17.5` → `4.18.0`)
2. **Version B**: Increments patch from A (`4.18.0` → `4.18.1`) 
3. **Version C**: Increments patch from B (`4.18.1` → `4.18.2`)
4. **Edges**: Linear progression A→B→C

### Metadata Generation
Nodes include realistic OpenShift metadata:
- `io.openshift.upgrades.graph.release.channels` - Channel name
- `io.openshift.upgrades.graph.release.manifestref` - SHA256 reference
- `url` - RHSA errata URL
- `release.openshift.io/architecture` - Architecture (if provided)

## AIDEV Notes & Gotchas

### AIDEV-NOTE: Edge Array Indexing
Edges are represented as `[2]int` arrays where values are **node indices**, not version strings. Index 0 = first node in the nodes array, etc.

### AIDEV-NOTE: JSON Marshaling
The `Node.Image` field is tagged as `json:"payload"` to match Cincinnati protocol expectations. Don't change this field name without understanding the protocol impact.

### AIDEV-NOTE: Version Parsing
All version handling uses `github.com/blang/semver/v4`. Always validate version strings before processing.

### AIDEV-NOTE: Channel Support
Currently only supports `version-not-found` channel. Adding new channels requires:
1. New case in `handleGraph` switch statement
2. New graph generation function
3. Update documentation in cmd/fauxinnati/README.md

## Testing Strategy

### Test File Organization
- `types_test.go` - JSON serialization/deserialization testing
- `server_test.go` - Unit tests for server logic and graph generation  
- `integration_test.go` - Full HTTP server testing with httptest

### AIDEV-TODO: Test Coverage Gaps
Consider adding tests for:
- Invalid semver handling edge cases
- Metadata generation accuracy
- Different architecture parameter handling
- Error response formats

## Development Guidelines

### Adding New Channels
When adding support for new channels:
1. Add case to `handleGraph` switch statement
2. Implement new `generate*Graph` function following existing patterns
3. Ensure realistic metadata generation
4. Add comprehensive tests
5. Update both README files

### Modifying Graph Structure
- Preserve Cincinnati protocol compatibility
- Maintain realistic version progression logic
- Keep edge indices consistent with node array positions
- Test JSON marshaling/unmarshaling thoroughly

### AIDEV-QUESTION: Future Extensions
Consider if the package needs:
- Support for conditional updates with actual risks
- More complex graph topologies (branching, parallel paths)
- Configuration-driven graph generation
- Multiple architecture support per request

## Dependencies
- `github.com/blang/semver/v4` - Semantic version handling (required)
- Standard library only for HTTP and JSON handling

## Protocol Compliance
This package implements a subset of the Cincinnati protocol. For full protocol specification, refer to the upstream Cincinnati documentation.