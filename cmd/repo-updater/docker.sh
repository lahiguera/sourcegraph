#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

docker build -f cmd/repo-updater/Dockerfile -t $IMAGE . \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION \
    --build-arg REPO_UPDATER_PKG
