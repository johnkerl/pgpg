#!/bin/bash

set -xeuo pipefail

#make -C ../manual
#make -C ../generator_go

#mkdir -p jsons
#mkdir -p pkg/lexers
#mkdir -p pkg/parsers

../generator_go/lexgen-tables   -o jsons/pascal-lex.json   bnfs/pascal.bnf
../generator_go/parsegen-tables -o jsons/pascal-parse.json bnfs/pascal.bnf

echo
echo DONE
