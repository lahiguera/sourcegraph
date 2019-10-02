#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -euxo pipefail

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

echo "--- go build"
for pkg in github.com/sourcegraph/sourcegraph/cmd/github-proxy; do
    go build -ldflags "-X github.com/sourcegraph/sourcegraph/pkg/version.version=$VERSION" -buildmode exe -tags dist -o $OUTPUT_DIR/$(basename $pkg) $pkg
done

