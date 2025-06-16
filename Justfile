refresh-fixture-inputs:
    curl 'https://api.openshift.com/api/upgrades_info/graph?channel=candidate-4.20' > pkg/fauxinnati/testdata/zz_fixture_TestGraph_JSONcandidate_4.20.json.input

# Build container images
images:
    #!/usr/bin/env bash
    set -euo pipefail
    
    # Get git commit hash (first 8 characters)
    git_sha=$(git rev-parse --short=8 HEAD)
    
    # Check if repository is dirty
    if [[ -n $(git status --porcelain) ]]; then
        tag="${git_sha}-dirty"
    else
        tag="${git_sha}"
    fi
    
    echo "Building fauxinnati image with tag: quay.io/petr-muller/fauxinnati:${tag}"
    
    # Build the image
    podman build -t "quay.io/petr-muller/fauxinnati:${tag}" -f images/fauxinnati/Containerfile .
    
    echo "Successfully built: quay.io/petr-muller/fauxinnati:${tag}"

# Build and publish container images
publish: images
    #!/usr/bin/env bash
    set -euo pipefail
    
    # Get git commit hash (first 8 characters)
    git_sha=$(git rev-parse --short=8 HEAD)
    
    # Check if repository is dirty
    if [[ -n $(git status --porcelain) ]]; then
        echo "Repository is dirty - refusing to publish dirty images"
        echo "Please commit your changes before publishing"
        exit 1
    fi
    
    # Tag is clean (no -dirty suffix)
    tag="${git_sha}"
    image_name="quay.io/petr-muller/fauxinnati"
    
    echo "Publishing ${image_name}:${tag} to registry..."
    
    # Push the digest tag
    podman push "${image_name}:${tag}"
    
    # Tag and push as latest
    podman tag "${image_name}:${tag}" "${image_name}:latest"
    podman push "${image_name}:latest"
    
    echo "Successfully published:"
    echo "  ${image_name}:${tag}"
    echo "  ${image_name}:latest"
