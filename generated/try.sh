#!/bin/bash

set -x

make -C ../generator

../generator/lexgen-tables -o jsons/arith-lex-whitespace.json bnfs/arith-lex-whitespace.bnf
../generator/lexgen-code   -o pkg/arith-lex-whitespace.go -type ArithLexWhitespaceLexer jsons/arith-lex-whitespace.json

../generator/lexgen-tables -o jsons/arith-lex.json bnfs/arith-lex.bnf
../generator/lexgen-code   -o pkg/arith-lex.go -type ArithLexLexer jsons/arith-lex.json

../generator/lexgen-tables -o jsons/sign-digit-lex.json bnfs/sign-digit-lex.bnf
../generator/lexgen-code   -o pkg/sign-digit-lex.go -type SignDigitLexLexer jsons/sign-digit-lex.json
