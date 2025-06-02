# Scripts

This directory contains utility scripts.

## search-ci-failures.fish

A Fish shell script that searches for failures in OpenShift CI jobs by scanning build logs from Google Cloud Storage. It searches through PR-based CI jobs and looks for specific patterns in the build logs, providing URLs to failed builds along with execution timestamps.

The script requires `gsutil` (Google Cloud SDK) and `gum` (terminal UI library) to be installed.

**Example usage:**
```bash
# Search for test failures in a specific presubmit job
./search-ci-failures.fish openshift/origin pull-ci-openshift-origin-master-e2e-aws 'test failed'

# Search with a minimum PR number to skip older PRs
./search-ci-failures.fish openshift/origin pull-ci-openshift-origin-master-e2e-aws 'test failed' 1000
```