#!/bin/bash

set -xeuo pipefail

make -C ../manual
make -C ../generator

mkdir -p jsons
mkdir -p pkg/parsers

../generator/parsegen-tables -o jsons/pemdas-parse.json     bnfs/pemdas.bnf
../generator/parsegen-tables -o jsons/statements-parse.json bnfs/statements.bnf

../generator/parsegen-code -o pkg/parsers/pemdas.go     -package parsers -type PEMDASParser     jsons/pemdas-parse.json
../generator/parsegen-code -o pkg/parsers/statements.go -package parsers -type StatementsParser jsons/statements-parse.json

echo
echo DONE
