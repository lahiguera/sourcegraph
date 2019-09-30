#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../../..
set -ex

# Override packages with their "enterprise" counterparts
export SERVER_PKG=github.com/sourcegraph/sourcegraph/enterprise/cmd/server
export FRONTEND_PKG=github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend
export MANAGEMENT_CONSOLE_PKG=github.com/sourcegraph/sourcegraph/enterprise/cmd/management-console
export REPO_UPDATER_PKG=github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater

export PRE_BUILD_SCRIPT=enterprise/cmd/server/pre-build.sh

./cmd/server/docker.sh
