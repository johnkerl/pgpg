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
echo; ./tryparse m:vic    'x = x + 1'
echo; ./tryparse m:vbc    'a AND b OR c AND d'
echo; ./tryparse m:pemdas '1*2+3'
echo; ./tryparse m:pemdas '1+2*3'
echo; ./tryparse g:arith  '1*2+3'
echo; ./tryparse g:arith  '1+2*3'
echo; ./tryparse g:stmts  'print(1);y=2; if(x=3)y=4;'
