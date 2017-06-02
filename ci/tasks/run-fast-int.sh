#!/usr/bin/env bash

set -e

my_dir="$( cd $(dirname $0) && pwd )"
release_dir="$( cd ${my_dir} && cd ../.. && pwd )"
workspace_dir="$( cd ${release_dir} && cd ../../../.. && pwd )"

pushd ${release_dir} > /dev/null

source ci/tasks/utils.sh

popd > /dev/null

check_param google_json_key_data

echo $google_json_key_data > key.json
gcloud auth activate-service-account --key-file=key.json

export GOPATH=${workspace_dir}
export PATH=${GOPATH}/bin:${PATH}

pushd ${release_dir} > /dev/null

echo $google_json_key_data > key.json
GOOGLE_APPLICATION_CREDENTIALS=`pwd`/key.json make test-fast-int

make clean-gcs

popd > /dev/null