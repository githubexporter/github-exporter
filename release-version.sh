#!/bin/bash

# ensure git is in the correct branch and has latest from remote.
git checkout master
git pull origin master

version=$(cat VERSION)
echo "version: $version"

# check if tag already exists.
if [ $(git tag -l "$version") ]; then
  echo "tag already exists. Ensure version number has been update in VERSION."
  exit 1
fi

# check version is in the correct format.
if ! [[ "$version" =~ ^[0-9.]+$ ]]; then
  echo "version: "$version" is in the wrong format."
  exit 1
fi

docker build -t infinityworks/github-exporter:latest -t infinityworks/github-exporter:$version .
docker push infinityworks/github-exporter:latest
docker push infinityworks/github-exporter:$version

