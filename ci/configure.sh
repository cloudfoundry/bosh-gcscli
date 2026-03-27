#!/usr/bin/env bash
set -euo pipefail

fly -t "${CONCOURSE_TARGET:-"storage-cli"}" set-pipeline \
    -p "bosh-gcs-cli" \
    -c "$(dirname "${0}")/pipeline.yml"
