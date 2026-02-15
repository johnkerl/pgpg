#!/bin/bash

set -euo pipefail

echo; echo ----------------------------------------------------------------
echo; ./apps/go/tryparse -e m:vic    'x = x + 1'
echo; ./apps/go/tryparse -e m:vbc    'a AND b OR c AND d'
echo; ./apps/go/tryparse -e m:pemdas '1*2+3'
echo; ./apps/go/tryparse -e m:pemdas '1+2*3'
echo; ./apps/go/tryparse -e g:pemdas '1*2+3'
echo; ./apps/go/tryparse -e g:pemdas '1+2*3'
echo; ./apps/go/tryparse -e g:stmts  'print(1);y=2; if(x=3)y=4;'
echo; ./apps/go/tryparse -e g:lisp   '(+ 1 (* 2 3) (* 4 5)) ; comment here'

echo; echo ----------------------------------------------------------------
echo; ./apps/go/tryparse -e g:seng 'the red cat quickly jumps over the green dog'

echo; echo ----------------------------------------------------------------
echo; ./apps/go/tryparse -e g:json  '[]'
echo; ./apps/go/tryparse -e g:json  '[1]'
echo; ./apps/go/tryparse -e g:json  '[1,2]'
echo; ./apps/go/tryparse -e g:json  '[1,2,3]'
echo; ./apps/go/tryparse -e g:json  '{}'
echo; ./apps/go/tryparse -e g:json  '{"a":1}'
echo; ./apps/go/tryparse -e g:json  '{"a":1, "b":2}'
echo; ./apps/go/tryparse -e g:json  '{"a":1, "b":2, "c":3}'
