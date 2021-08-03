#!/usr/bin/env bash

set -eu

if [[ $(lpass status -q; echo $?) != 0 ]]; then
  echo "Login with lpass first"
  exit 1
fi

fly -t bosh-ecosystem set-pipeline -p "bosh-gcs-cli" \
    -c $(dirname $0)/pipeline.yml \
    --load-vars-from <(lpass show -G "gcscli-concourse-secrets" --notes)
