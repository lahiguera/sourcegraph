#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

docker build -f cmd/management-console/Dockerfile -t $IMAGE . \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION \
    --build-arg MANAGEMENT_CONSOLE_PKG
