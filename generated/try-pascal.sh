#!/bin/bash

set -xeuo pipefail

#make -C ../manual
#make -C ../generator

#mkdir -p jsons
#mkdir -p pkg/lexers
#mkdir -p pkg/parsers

../generator/lexgen-tables   -o jsons/pascal-lex.json   bnfs/pascal.bnf
../generator/parsegen-tables -o jsons/pascal-parse.json bnfs/pascal.bnf

echo
echo DONE
