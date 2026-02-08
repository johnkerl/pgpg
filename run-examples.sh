#!/bin/bash

set -euo pipefail

echo; ./runners/tryparse m:vic    'x = x + 1'
echo; ./runners/tryparse m:vbc    'a AND b OR c AND d'
echo; ./runners/tryparse m:pemdas '1*2+3'
echo; ./runners/tryparse m:pemdas '1+2*3'
echo; ./runners/tryparse g:arith  '1*2+3'
echo; ./runners/tryparse g:arith  '1+2*3'
echo; ./runners/tryparse g:stmts  'print(1);y=2; if(x=3)y=4;'
