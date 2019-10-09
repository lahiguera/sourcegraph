#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -euxo pipefail

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

cp -a ./cmd/server/rootfs/. "$OUTPUT_DIR"
export BINDIR="$OUTPUT_DIR/usr/local/bin"
mkdir -p "$BINDIR"

export CTAGS_D_OUTPUT_PATH="$OUTPUT_DIR/.ctags.d"
export SYMBOLS_EXECUTABLE_OUTPUT_PATH="$BINDIR/symbols"
export BUILD_TYPE=dist

parallel --keep-order --line-buffer --tag {} ::: "cmd/server/build-go.sh" "cmd/symbols/build.sh buildSymbolsDockerImageDependencies"

echo "--- prometheus config"
cp -r docker-images/prometheus/config "$OUTPUT_DIR/sg_config_prometheus"
mkdir "$OUTPUT_DIR/sg_prometheus_add_ons"
cp dev/prometheus/linux/prometheus_targets.yml "$OUTPUT_DIR/sg_prometheus_add_ons"

echo "--- grafana config"
cp -r docker-images/grafana/config "$OUTPUT_DIR/sg_config_grafana"
cp -r dev/grafana/linux "$OUTPUT_DIR/sg_config_grafana/provisioning/datasources"
