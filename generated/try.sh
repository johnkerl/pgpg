#!/bin/bash

set -x

make -C ../generator

for name in \
  arith-lex-whitespace \
  arith-lex \
  sign-digit-lex
do
  ../generator/lexgen-tables -o jsons/$name.json bnfs/$name.bnf
  ../generator/lexgen-code   -o pkg/$name.go     jsons/$name.json
done
