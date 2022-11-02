#!/usr/bin/env bash
set -euo pipefail

my_dir="$( cd "$(dirname "${0}")" && pwd )"
pushd "${my_dir}" > /dev/null
    source utils.sh
    set_env
popd > /dev/null

pushd "${release_dir}"
    echo "running with go version: $(go version)"
    make test-unit
popd > /dev/null
