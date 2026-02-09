#!/bin/bash

set -xeuo pipefail

make -C ../manual
make -C ../generator

mkdir -p jsons
mkdir -p pkg/lexers

../generator/lexgen-tables -o jsons/sign-digit-lex.json bnfs/sign-digit.bnf
../generator/lexgen-tables -o jsons/pemdas-lex.json     bnfs/pemdas.bnf
../generator/lexgen-tables -o jsons/statements-lex.json bnfs/statements.bnf
../generator/lexgen-tables -o jsons/seng-lex.json       bnfs/seng.bnf
../generator/lexgen-tables -o jsons/lisp-lex.json       bnfs/lisp.bnf
../generator/lexgen-tables -o jsons/json-lex.json       bnfs/json.bnf

../generator/lexgen-code -o pkg/lexers/sign-digit.go -type SignDigitLexer  jsons/sign-digit-lex.json
../generator/lexgen-code -o pkg/lexers/pemdas.go     -type PEMDASLexer     jsons/pemdas-lex.json
../generator/lexgen-code -o pkg/lexers/statements.go -type StatementsLexer jsons/statements-lex.json
../generator/lexgen-code -o pkg/lexers/seng.go       -type SENGLexer       jsons/seng-lex.json
../generator/lexgen-code -o pkg/lexers/lisp.go       -type LISPLexer       jsons/lisp-lex.json
../generator/lexgen-code -o pkg/lexers/json.go       -type JSONLexer       jsons/json-lex.json

echo
echo DONE
