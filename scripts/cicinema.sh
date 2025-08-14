#!/bin/bash
# 
INPUT="$1"
TEMP_DIR=""
CLEANUP_TEMP=false

DELAY=1 #second

# Check for gsutil presence
if ! command -v gsutil &> /dev/null; then
    echo "Error: gsutil is not installed or not in PATH" >&2
    echo "Please install Google Cloud SDK or ensure gsutil is available" >&2
    exit 1
fi

extract_gcs_path() {
    local url="$1"
    if [[ "$url" =~ https://prow\.ci\.openshift\.org/view/gs/(.+) ]]; then
        local path="${BASH_REMATCH[1]}"
        # Ensure trailing slash for directory operations
        [[ "$path" != */ ]] && path="${path}/"
        echo "gs://${path}"
    else
        echo ""
    fi
}

find_oc_adm_upgrade_status_snapshots() {
    local base_path="$1"
    local glob_pattern="${base_path}artifacts/*/*/artifacts/junit/adm-upgrade-status/"
    
    local first_file=$(gsutil ls "$glob_pattern" 2>/dev/null | head -1)
    
    if [[ -n "$first_file" ]]; then
        local junit_dir=$(dirname "$first_file")/
        echo "$junit_dir"
        return 0
    fi
    
    return 1
}

cleanup() {
    if [[ "$CLEANUP_TEMP" == "true" && -n "$TEMP_DIR" && -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

trap cleanup EXIT

if [[ "$INPUT" =~ ^https:// ]]; then
    # Extract GCS bucket path from URL
    GCS_BASE=$(extract_gcs_path "$INPUT")
    
    if [[ -z "$GCS_BASE" ]]; then
        echo "Error: Unable to extract GCS path from URL: $INPUT" >&2
        exit 1
    fi
    
    echo "Extracted GCS base path: $GCS_BASE"
    
    # Find the junit/adm-upgrade-status path
    JUNIT_PATH=$(find_oc_adm_upgrade_status_snapshots "$GCS_BASE")
    
    if [[ -z "$JUNIT_PATH" ]]; then
        echo "Error: Could not find junit/adm-upgrade-status path under $GCS_BASE" >&2
        exit 1
    fi
    
    echo "Found junit path: $JUNIT_PATH"
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    CLEANUP_TEMP=true
    
    echo "Downloading artifacts to temporary directory: $TEMP_DIR"
    
    # Download the directory contents
    if ! gsutil -m cp -r "$JUNIT_PATH*" "$TEMP_DIR/" 2>/dev/null; then
        echo "Error: Failed to download from $JUNIT_PATH" >&2
        exit 1
    fi
    
    DIR="$TEMP_DIR"
else
    echo "Input needs to be URL to OpenShift CI ProwJob" >&2
    echo "Example (PR): https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/30109/pull-ci-openshift-origin-main-e2e-aws-ovn-upgrade/1955704357424992256"
    echo "Example (Periodic): https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-release-master-ci-4.20-upgrade-from-stable-4.19-e2e-aws-ovn-upgrade/1955683492754886656"
fi

if [[ ! -d "$DIR" ]]; then
    echo "Error: Directory does not exist: $DIR" >&2
    exit 1
fi

# Escape sequence for "move cursor to top-left"
CURSOR_HOME="\033[H"

# Escape sequence for "clear to end of screen" (optional, to remove leftovers)
CLEAR_EOS="\033[J"

FILES=("$DIR"/*)
IFS=$'\n' FILES=($(sort <<<"${FILES[*]}"))
unset IFS

TOTAL=${#FILES[@]}
CURRENT=1

for file in "${FILES[@]}"; do
    echo -ne "$CURSOR_HOME$CLEAR_EOS"
    cat "$file"
    echo ""
    echo ""
    echo "Snapshot: $CURRENT/$TOTAL"
    sleep "$DELAY"
    ((CURRENT++))
done
