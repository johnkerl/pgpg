#!/bin/bash

us=$(basename $0)
usage() {
  echo "Usage: $us {version}" 1>&2
  echo "Version must match vN.N.N (e.g. v1.2.3)" 1>&2
  exit 1
}

if [ $# -ne 1 ]; then
  usage
fi
version="$1"

if ! [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "$(basename "$0"): invalid version '$version'" 1>&2
  usage
fi

git tag    "$version"
git tag "go/$version"
git push origin    "$version"
git push origin "go/$version"
