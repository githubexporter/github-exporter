#!/bin/bash

set -euo pipefail

# ensure git is in the correct branch and has latest from remote.
git checkout master
git pull origin master

version=$(cat VERSION)
echo "Version: $version"

# check version is in the correct format.
if ! [[ "$version" =~ ^v[0-9.]+$ ]]; then
  echo "Version ($version) is in the wrong format."
  exit 1
fi

# use the toolchain version, falling back to the go version if toolchain is omitted.
go_version=$(go mod edit -json | jq -r '.Toolchain // .Go' | sed 's/go//')
echo "Go Version: $go_version"
if ! [[ "$go_version" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
  echo "Go Version ($go_version) is in the wrong format."
  exit 1
fi

if docker manifest inspect githubexporter/github-exporter:$version > /dev/null 2>&1; then
    echo "Image for version ($version) already exists on the registry. Skipping build."
    exit 1
fi

docker buildx build --platform linux/amd64,linux/arm64 --build-arg GOLANG_VERSION=$go_version -t githubexporter/github-exporter:latest -t githubexporter/github-exporter:$version --push .