#!/bin/bash

set -xeuo pipefail

make -C ../manual
make -C ../generator

mkdir -p jsons
mkdir -p pkg/parsers

../generator/parsegen-tables -o jsons/arith-parse.json  bnfs/arith.bnf
../generator/parsegen-tables -o jsons/statements-parse.json bnfs/statements.bnf

../generator/parsegen-code -o pkg/parsers/arith-parse.go  -package parsers -type ArithParser           jsons/arith-parse.json
../generator/parsegen-code -o pkg/parsers/statements-parse.go -package parsers -type StatementsParser jsons/statements-parse.json

echo
echo DONE
