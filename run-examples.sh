#!/bin/bash

set -euo pipefail

echo; echo ----------------------------------------------------------------
echo; ./apps/go/tryparse m:vic     expr 'x = x + 1'
echo; ./apps/go/tryparse m:vbc     expr 'a AND b OR c AND d'
echo; ./apps/go/tryparse m:pemdas  expr '1*2+3'
echo; ./apps/go/tryparse m:pemdas  expr '1+2*3'
echo; ./apps/go/tryparse g:pemdas  expr '1*2+3'
echo; ./apps/go/tryparse g:pemdas  expr '1+2*3'
echo; ./apps/go/tryparse g:stmts   expr 'print(1);y=2; if(x=3)y=4;'
echo; ./apps/go/tryparse g:lisp    expr '(+ 1 (* 2 3) (* 4 5)) ; comment here'

echo; echo ----------------------------------------------------------------
echo; ./apps/go/tryparse g:seng expr 'the red cat quickly jumps over the green dog'

echo; echo ----------------------------------------------------------------
echo; ./apps/go/tryparse g:json    expr '[]'
echo; ./apps/go/tryparse g:json    expr '[1]'
echo; ./apps/go/tryparse g:json    expr '[1,2]'
echo; ./apps/go/tryparse g:json    expr '[1,2,3]'
echo; ./apps/go/tryparse g:json    expr '{}'
echo; ./apps/go/tryparse g:json    expr '{"a":1}'
echo; ./apps/go/tryparse g:json    expr '{"a":1, "b":2}'
echo; ./apps/go/tryparse g:json    expr '{"a":1, "b":2, "c":3}'
