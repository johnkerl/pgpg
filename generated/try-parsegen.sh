#!/bin/bash

set -xeuo pipefail

make -C ../manual
make -C ../generator

mkdir -p jsons
mkdir -p pkg/parsers

../generator/parsegen-tables -o jsons/pemdas-parse.json     bnfs/pemdas.bnf
../generator/parsegen-tables -o jsons/statements-parse.json bnfs/statements.bnf
../generator/parsegen-tables -o jsons/seng-parse.json       bnfs/seng.bnf
../generator/parsegen-tables -o jsons/lisp-parse.json       bnfs/lisp.bnf
../generator/parsegen-tables -o jsons/json-parse.json       bnfs/json.bnf

../generator/parsegen-code -o pkg/parsers/pemdas.go     -package parsers -type PEMDASParser     jsons/pemdas-parse.json
../generator/parsegen-code -o pkg/parsers/statements.go -package parsers -type StatementsParser jsons/statements-parse.json
../generator/parsegen-code -o pkg/parsers/seng.go       -package parsers -type SENGParser       jsons/seng-parse.json
../generator/parsegen-code -o pkg/parsers/lisp.go       -package parsers -type LISPParser       jsons/lisp-parse.json
../generator/parsegen-code -o pkg/parsers/json.go       -package parsers -type JSONParser       jsons/json-parse.json

echo
echo DONE
