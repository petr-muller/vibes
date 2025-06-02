# Vibes

A collection of utility scripts and tools.

## Scripts

The `scripts/` directory contains various utility scripts:

- **search-ci-failures.fish**: Search for failures in OpenShift CI jobs by scanning build logs from Google Cloud Storage

See [scripts/README.md](scripts/README.md) for detailed documentation.

## Requirements

- Fish shell (for Fish scripts)
- Google Cloud SDK (`gsutil`)
- `gum` terminal UI library

## Usage

Scripts can be run directly from the `scripts/` directory:

```bash
./scripts/search-ci-failures.fish openshift/origin pull-ci-openshift-origin-master-e2e-aws 'test failed'
```