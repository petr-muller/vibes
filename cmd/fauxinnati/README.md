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
# Test basic channels
curl "http://localhost:8080/api/upgrades_info/graph?channel=version-not-found&version=4.17.5&arch=amd64"
curl "http://localhost:8080/api/upgrades_info/graph?channel=channel-head&version=4.18.5&arch=amd64"
curl "http://localhost:8080/api/upgrades_info/graph?channel=simple&version=4.17.5&arch=amd64"

# Test risk-based channels
curl "http://localhost:8080/api/upgrades_info/graph?channel=risks-always&version=4.17.5&arch=amd64"
curl "http://localhost:8080/api/upgrades_info/graph?channel=risks-matching&version=4.17.5&arch=amd64"
curl "http://localhost:8080/api/upgrades_info/graph?channel=risks-nonmatching&version=4.17.5&arch=amd64"

# Test comprehensive smoke-test channel
curl "http://localhost:8080/api/upgrades_info/graph?channel=smoke-test&version=4.17.5&arch=amd64"
```

## API Endpoint

- `GET /api/upgrades_info/graph` - Returns update graph based on channel and version parameters

### Required Parameters

- `channel` - Update channel (supports: `version-not-found`, `channel-head`, `simple`, `risks-always`, `risks-matching`, `risks-nonmatching`, `smoke-test`)
- `version` - Base version in semver format (e.g., `4.17.5`)

### Optional Parameters

- `arch` - Architecture (e.g., `amd64`)

## Channel Behaviors

### Basic Channels

#### `version-not-found`
Generates a three-node graph where the requested version is excluded:
- A: Next minor version with patch 0 (e.g., 4.17.5 → 4.18.0)  
- B: A + patch 1 (e.g., 4.18.1)
- C: A + patch 2 (e.g., 4.18.2)
- Graph: A → B → C (unconditional edges)

#### `channel-head`  
Generates a three-node graph where the client's version is the head:
- A: Previous minor version with patch 0 (e.g., 4.18.5 → 4.17.0)
- B: A + patch 1 (e.g., 4.17.1) 
- C: Client's version (e.g., 4.18.5)
- Graph: A → B → C (unconditional edges)

#### `simple`
Generates a three-node linear progression from the client's version:
- A: Client's version (e.g., 4.17.5)
- B: A + patch 1 (e.g., 4.17.6)
- C: Next minor version (e.g., 4.18.0)
- Graph: A → B → C (unconditional edges)

### Risk-Based Channels

#### `risks-always`
Generates a three-node graph with conditional edges that always apply:
- A: Client's version (e.g., 4.17.5)
- B: A + patch 1 (e.g., 4.17.6)
- C: Next minor version (e.g., 4.18.0)
- Graph: A conditionally connects to B and C with "Always" matching risk
- Risk: SyntheticRisk with Always matching rule (always blocks updates)

#### `risks-matching`
Generates a three-node graph with PromQL conditional edges that match:
- A: Client's version (e.g., 4.17.5)
- B: A + patch 1 (e.g., 4.17.6)
- C: Next minor version (e.g., 4.18.0)
- Graph: A conditionally connects to B and C with PromQL `vector(1)` risk
- Risk: SyntheticRisk with PromQL that always matches (always blocks updates)

#### `risks-nonmatching`
Generates a three-node graph with PromQL conditional edges that don't match:
- A: Client's version (e.g., 4.17.5)
- B: A + patch 1 (e.g., 4.17.6)
- C: Next minor version (e.g., 4.18.0)
- Graph: A conditionally connects to B and C with PromQL `vector(0)` risk
- Risk: SyntheticRisk with PromQL that never matches (never blocks updates)

### Comprehensive Test Channel

#### `smoke-test`
Generates a comprehensive 13-node graph exercising most Cincinnati features:
- **13 nodes**: D(4.16.0), E(4.17.5), F(4.16.1), G(4.17.6), H(4.18.0), I(4.17.7), J(4.18.1), K(4.17.8), L(4.18.2), M(4.17.9), N(4.18.3), O(4.17.10), P(4.18.4)
- **4 unconditional edges**: D→E, D→F, E→G, E→H
- **4 conditional edge groups**:
  - **RiskA (Always)**: F→I, G→K with always-matching risk
  - **RiskBMatches (PromQL vector(1))**: H→J, I→L with always-matching PromQL
  - **RiskCNoMatch (PromQL vector(0))**: J→N, K→O with never-matching PromQL
  - **Combined risks**: L→P, M→P with all three risk types combined
- **Purpose**: Comprehensive testing of graph traversal, risk evaluation, and complex conditional logic
