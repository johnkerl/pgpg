#!/bin/bash

set -euo pipefail

echo; ./runners/tryparse m:vic     expr 'x = x + 1'
echo; ./runners/tryparse m:vbc     expr 'a AND b OR c AND d'
echo; ./runners/tryparse m:pemdas  expr '1*2+3'
echo; ./runners/tryparse m:pemdas  expr '1+2*3'
echo; ./runners/tryparse g:pemdas  expr '1*2+3'
echo; ./runners/tryparse g:pemdas  expr '1+2*3'
echo; ./runners/tryparse g:stmts   expr 'print(1);y=2; if(x=3)y=4;'
echo; ./runners/tryparse g:lisp    expr '(+ 1 (* 2 3) (* 4 5)) ; comment here'
