#!/bin/bash

if [ $# -ne 1 ]; then
  echo "$0: need tag argument in the form v0.1.0" 1>&2
  exit 1
fi
version="$1"
git tag $version
git tag lib/go/$version
git tag generators/go/$version
git push origin $version
git push origin lib/go/$version
git push origin generators/go/$version
