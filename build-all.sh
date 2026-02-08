#!/bin/bash

set -euo pipefail

sep() {
  echo
  echo ================================================================
  echo
}

sep
make -C manual
make -C manual test

sep
make -C generator
make -C generator test

sep
cd generated
./try-lexgen.sh
./try-parsegen.sh

sep

cd ../runners
make
cd ..
