#!/bin/bash

# ensure git is in the correct branch and has latest from remote.
git checkout master
git pull origin master

version=$(cat VERSION.txt)
echo "version: $version"

# exist if tag alredy exists.
if [ $(git tag -l "$version") ]; then
  echo "tag already exists. Ensure verion number has been update in VERSION."
  exit 1
fi

# check version is in the correct format.
if ! [[ "$version" =~ ^[0-9.]+$ ]]; then
  echo "version: "$version" is in the wrong format."
  exit 1
fi

# tag current latest commit.
echo "tagging..."
git tag $version

# push tag to trigger build.
echo "pushing tag to trigger build..."
git push origin $version
