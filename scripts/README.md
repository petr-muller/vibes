# Scripts

This directory contains utility scripts for OpenShift development and testing.

## oc-login-token.fish

An enhanced OpenShift login script that integrates with Kerberos authentication and uses `ocp-sso-token` for automatic token retrieval. Features a modern terminal UI with `gum` for better user experience.

**Requirements:**
- `gum` (terminal UI library)
- `oc` (OpenShift CLI)
- `ocp-sso-token` (OAuth token retrieval tool)
- `klist` and `kinit` (Kerberos tools)
- `yq` (YAML processor)

**Usage:**
```bash
# Basic usage (interactive context selection)
./oc-login-token.fish

# With custom Kerberos user
./oc-login-token.fish --user myuser

# With custom SSO provider
./oc-login-token.fish --provider my-sso-provider

# Get help
./oc-login-token.fish --help
```

**Features:**
- Automatic Kerberos ticket validation and renewal
- Interactive context selection from existing kubeconfig contexts
- Automatic token retrieval using `ocp-sso-token`
- Token refresh for existing contexts (no new context creation)
- Comprehensive error handling and user feedback
- Modern terminal UI with colors and styling

**Function usage:**
The script can also be sourced to provide the `oc-login-token` function:
```fish
source scripts/oc-login-token.fish
oc-login-token --user myuser --provider my-provider
```

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

## NetworkPolicy Testing Scripts

Three Fish shell scripts for testing NetworkPolicy functionality in OpenShift clusters. All scripts require `oc` (OpenShift CLI) and `gum` (terminal UI library) to be installed.

### test-networkpolicy-isolation.fish

Tests that NetworkPolicy denies outbound internet access from pods by deploying a test pod that attempts to curl an external URL.

**Usage:**
```bash
./test-networkpolicy-isolation.fish <namespace>
```

**What it does:**
- Creates a pod with a curl container that attempts to reach https://www.google.com
- Waits for the operation to complete and checks the exit code
- Returns success (0) if the connection is blocked, failure (1) if allowed

### test-networkpolicy-pod-to-pod.fish

Tests that NetworkPolicy blocks pod-to-pod communication within a namespace by deploying server and client pods.

**Usage:**
```bash
./test-networkpolicy-pod-to-pod.fish <namespace>
```

**What it does:**
- Creates a server pod running a Python HTTP server on port 8080
- Creates a service to expose the server pod
- Creates a client pod that attempts to curl the server via the service
- Returns success (0) if the connection is blocked, failure (1) if allowed

### test-networkpolicy-external-access.fish

Tests that NetworkPolicy blocks external access to pods by creating a pod with an HTTP server, exposing it via Route, and attempting to access it from outside the cluster.

**Usage:**
```bash
./test-networkpolicy-external-access.fish <namespace>
```

**What it does:**
- Creates a pod running a Python HTTP server on port 8080
- Creates a service and route to expose the pod externally
- Attempts to curl the route from the local machine
- Returns success (0) if the connection is blocked, failure (1) if allowed

All NetworkPolicy test scripts automatically clean up created resources after testing and provide colored output for easy interpretation of results.