#!/bin/bash

set -xeuo pipefail

#make -C ../manual
#make -C ../generators/go

#mkdir -p jsons
#mkdir -p go/pkg/lexers
#mkdir -p go/pkg/parsers

../generators/go/lexgen-tables   -o jsons/pascal-lex.json   ../bnfs/pascal.bnf
../generators/go/parsegen-tables -o jsons/pascal-parse.json ../bnfs/pascal.bnf

echo
echo DONE
