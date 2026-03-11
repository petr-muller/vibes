#!/usr/bin/env bash

set -euo pipefail

command="fauxinnati"
git_commit="$( git describe --tags --always --dirty )"
build_date="$( date -u '+%Y%m%d' )"
version="v${build_date}-${git_commit}"

eval $(go env | grep -e "GOHOSTOS" -e "GOHOSTARCH")
GOOS=${GOOS:-${GOHOSTOS}}

set -x

CGO_ENABLED=0 GOOS="${GOOS}" go build -ldflags "-X 'github.com/petr-muller/vibes/pkg/version.Name=${command}' -X 'github.com/petr-muller/vibes/pkg/version.Version=${version}'" -a -installsuffix cgo -o "${command}" "./cmd/${command}"
