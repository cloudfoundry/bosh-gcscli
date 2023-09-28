#!/usr/bin/env bash
set -euo pipefail

my_dir="$( cd "$(dirname "${0}")" && pwd )"
pushd "${my_dir}" > /dev/null
    source utils.sh
    set_env
popd > /dev/null

# inputs
semver_dir="${workspace_dir}/version-semver"

# outputs
output_dir=${workspace_dir}/out

semver="$(cat "${semver_dir}/number")"
export CGO_ENABLED=0

binname="bosh-gcscli-${semver}-${GOOS}-amd64"
if [ "${GOOS}" = "windows" ]; then
	binname="${binname}.exe"
fi

pushd "${release_dir}" > /dev/null
  echo -e "\n building artifact using $(go version)..."
  go build -ldflags "-X main.version=${semver}" \
    -o "out/${binname}"                          \
    github.com/cloudfoundry/bosh-gcscli

  echo -e "\n sha1 of artifact..."
  sha1sum "out/${binname}"

  mv "out/${binname}" "${output_dir}/"
popd > /dev/null
