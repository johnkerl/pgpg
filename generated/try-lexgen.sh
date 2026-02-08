#!/bin/bash

set -xeuo pipefail

make -C ../generator

mkdir -p jsons
mkdir -p pkg/lexers

../generator/lexgen-tables -o jsons/sign-digit-lex.json bnfs/sign-digit.bnf
../generator/lexgen-tables -o jsons/arith-lex.json      bnfs/arith.bnf
../generator/lexgen-tables -o jsons/arithw-lex.json     bnfs/arithw.bnf

../generator/lexgen-code -o pkg/lexers/sign-digit-lex.go -type SignDigitLexer       jsons/sign-digit-lex.json
../generator/lexgen-code -o pkg/lexers/arith-lex.go      -type ArithLexer           jsons/arith-lex.json
../generator/lexgen-code -o pkg/lexers/arithw-lex.go     -type ArithWhitespaceLexer jsons/arithw-lex.json

echo
echo DONE
