# Vibes

A collection of utility scripts and tools for OpenShift development and testing.

## Tools

- **fauxinnati** - Mock Cincinnati update graph server (see [cmd/fauxinnati/README.md](cmd/fauxinnati/README.md))

## Scripts

The `scripts/` directory contains various utility scripts:

- **search-ci-failures.fish**: Search for failures in OpenShift CI jobs by scanning build logs from Google Cloud Storage

See [scripts/README.md](scripts/README.md) for detailed documentation.

## Requirements

- Go 1.24+ (for Go tools)
- Fish shell (for Fish scripts)
- Google Cloud SDK (`gsutil`)
- `gum` terminal UI library

## Development

### Go Code Standards

- Run `gofmt -w .` to format Go code before committing
- Use `gotestsum` for running tests to get enhanced output
- Keep dependencies minimal
- Follow Go best practices and idioms

### Testing

```bash
# Run tests with enhanced output
gotestsum ./...

# Run tests for specific package
gotestsum ./pkg/packagename

# Traditional go test also works
go test ./... -v
```

### General Guidelines

- Follow existing code conventions in each directory
- Update relevant README files when adding features
- Maintain backward compatibility for existing tools