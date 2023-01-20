#!/usr/bin/env bash
set -euo pipefail

if [[ $(lpass status -q; echo $?) != 0 ]]; then
  echo "Login with lpass first"
  exit 1
fi

fly -t bosh-ecosystem set-pipeline -p "bosh-gcs-cli" \
    -c "$(dirname "${0}")/pipeline.yml"
