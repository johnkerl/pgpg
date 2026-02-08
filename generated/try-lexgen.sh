#!/bin/bash

set -xeuo pipefail

make -C ../manual
make -C ../generator

mkdir -p jsons
mkdir -p pkg/lexers

../generator/lexgen-tables -o jsons/sign-digit-lex.json bnfs/sign-digit.bnf
../generator/lexgen-tables -o jsons/pemdas-lex.json      bnfs/pemdas.bnf
../generator/lexgen-tables -o jsons/statements-lex.json bnfs/statements.bnf

../generator/lexgen-code -o pkg/lexers/sign-digit.go -type SignDigitLexer  jsons/sign-digit-lex.json
../generator/lexgen-code -o pkg/lexers/pemdas.go     -type PEMDASLexer     jsons/pemdas-lex.json
../generator/lexgen-code -o pkg/lexers/statements.go -type StatementsLexer jsons/statements-lex.json

echo
echo DONE
