#!/bin/bash

set -xeuo pipefail

make -C ../manual
make -C ../generator

mkdir -p jsons
mkdir -p pkg/lexers

../generator/lexgen-tables -o jsons/sign-digit-lex.json bnfs/sign-digit.bnf
../generator/lexgen-tables -o jsons/arith-lex.json      bnfs/arith.bnf
../generator/lexgen-tables -o jsons/statements-lex.json bnfs/statements.bnf

../generator/lexgen-code -o pkg/lexers/sign-digit-lex.go -type SignDigitLexer       jsons/sign-digit-lex.json
../generator/lexgen-code -o pkg/lexers/arith-lex.go      -type ArithLexer           jsons/arith-lex.json
../generator/lexgen-code -o pkg/lexers/statements-lex.go -type StatementsLexer      jsons/statements-lex.json

echo
echo DONE
