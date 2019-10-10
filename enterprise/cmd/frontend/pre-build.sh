#!/usr/bin/env bash

set -exo pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../..

parallel_run() {
    log_file=$(mktemp)
    trap "rm -rf $log_file" EXIT

    parallel --keep-order --line-buffer --tag --joblog $log_file "$@"
    cat $log_file
}

echo "--- yarn root"
yarn --mutex network --frozen-lockfile --network-timeout 60000

NODE_ENV=${NODE_ENV:-production}
TARGETS=${TARGETS:-phabricator}

parallel_run {} ::: "env NODE_ENV=$NODE_ENV browser/build.sh" "env TARGETS=$TARGETS web/build.sh"

echo "--- generate"
enterprise/dev/generate.sh
