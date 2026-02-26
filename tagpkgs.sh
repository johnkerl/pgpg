#!/bin/bash

if [ $# -ne 1 ]; then
  echo "$0: need tag argument in the form v0.1.0" 1>&2
  exit 1
fi
version="$1"
git tag $version
git tag go/$version
git push origin $version
git push origin go/$version
