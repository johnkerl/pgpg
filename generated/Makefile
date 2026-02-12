# ----------------------------------------------------------------
LEXGENS=pkg/lexers/sign-digit.go pkg/lexers/pemdas.go pkg/lexers/pemdas_hinted.go pkg/lexers/statements.go pkg/lexers/seng.go pkg/lexers/lisp.go pkg/lexers/json.go
PARSEGENS=pkg/parsers/pemdas.go pkg/parsers/pemdas_hinted.go pkg/parsers/statements.go pkg/parsers/seng.go pkg/parsers/lisp.go pkg/parsers/json.go

all: $(LEXGENS) $(PARSEGENS)

# ----------------------------------------------------------------
pkg/lexers/sign-digit.go: jsons/sign-digit-lex.json
	../generator/lexgen-code -o pkg/lexers/sign-digit.go -type SignDigitLexer jsons/sign-digit-lex.json
pkg/lexers/pemdas.go: jsons/pemdas-lex.json
	../generator/lexgen-code -o pkg/lexers/pemdas.go -type PEMDASLexer jsons/pemdas-lex.json
pkg/lexers/pemdas_hinted.go: jsons/pemdas_hinted-lex.json
	../generator/lexgen-code -o pkg/lexers/pemdas_hinted.go -type PEMDASHintedLexer jsons/pemdas_hinted-lex.json
pkg/lexers/statements.go: jsons/statements-lex.json
	../generator/lexgen-code -o pkg/lexers/statements.go -type StatementsLexer jsons/statements-lex.json
pkg/lexers/seng.go: jsons/seng-lex.json
	../generator/lexgen-code -o pkg/lexers/seng.go -type SENGLexer jsons/seng-lex.json
pkg/lexers/lisp.go: jsons/lisp-lex.json
	../generator/lexgen-code -o pkg/lexers/lisp.go -type LISPLexer jsons/lisp-lex.json
pkg/lexers/json.go: jsons/json-lex.json
	../generator/lexgen-code -o pkg/lexers/json.go -type JSONLexer jsons/json-lex.json

jsons/sign-digit-lex.json: bnfs/sign-digit.bnf
	../generator/lexgen-tables -o jsons/sign-digit-lex.json bnfs/sign-digit.bnf
jsons/pemdas-lex.json: bnfs/pemdas.bnf
	../generator/lexgen-tables -o jsons/pemdas-lex.json bnfs/pemdas.bnf
jsons/pemdas_hinted-lex.json: bnfs/pemdas_hinted.bnf
	../generator/lexgen-tables -o jsons/pemdas_hinted-lex.json bnfs/pemdas_hinted.bnf
jsons/statements-lex.json: bnfs/statements.bnf
	../generator/lexgen-tables -o jsons/statements-lex.json bnfs/statements.bnf
jsons/seng-lex.json: bnfs/seng.bnf
	../generator/lexgen-tables -o jsons/seng-lex.json bnfs/seng.bnf
jsons/lisp-lex.json: bnfs/lisp.bnf
	../generator/lexgen-tables -o jsons/lisp-lex.json bnfs/lisp.bnf
jsons/json-lex.json: bnfs/json.bnf
	../generator/lexgen-tables -o jsons/json-lex.json bnfs/json.bnf

pkg/parsers/pemdas.go: jsons/pemdas-parse.json
	../generator/parsegen-code -o pkg/parsers/pemdas.go -package parsers -type PEMDASParser jsons/pemdas-parse.json
pkg/parsers/pemdas_hinted.go: jsons/pemdas_hinted-parse.json
	../generator/parsegen-code -o pkg/parsers/pemdas_hinted.go -package parsers -type PEMDASHintedParser jsons/pemdas_hinted-parse.json
pkg/parsers/statements.go: jsons/statements-parse.json
	../generator/parsegen-code -o pkg/parsers/statements.go -package parsers -type StatementsParser jsons/statements-parse.json
pkg/parsers/seng.go: jsons/seng-parse.json
	../generator/parsegen-code -o pkg/parsers/seng.go -package parsers -type SENGParser jsons/seng-parse.json
pkg/parsers/lisp.go: jsons/lisp-parse.json
	../generator/parsegen-code -o pkg/parsers/lisp.go -package parsers -type LISPParser jsons/lisp-parse.json
pkg/parsers/json.go: jsons/json-parse.json
	../generator/parsegen-code -o pkg/parsers/json.go -package parsers -type JSONParser jsons/json-parse.json

jsons/pemdas-parse.json: bnfs/pemdas.bnf
	../generator/parsegen-tables -o jsons/pemdas-parse.json bnfs/pemdas.bnf
jsons/pemdas_hinted-parse.json: bnfs/pemdas_hinted.bnf
	../generator/parsegen-tables -o jsons/pemdas_hinted-parse.json bnfs/pemdas_hinted.bnf
jsons/statements-parse.json: bnfs/statements.bnf
	../generator/parsegen-tables -o jsons/statements-parse.json bnfs/statements.bnf
jsons/seng-parse.json: bnfs/seng.bnf
	../generator/parsegen-tables -o jsons/seng-parse.json bnfs/seng.bnf
jsons/lisp-parse.json: bnfs/lisp.bnf
	../generator/parsegen-tables -o jsons/lisp-parse.json bnfs/lisp.bnf
jsons/json-parse.json: bnfs/json.bnf
	../generator/parsegen-tables -o jsons/json-parse.json bnfs/json.bnf


# ----------------------------------------------------------------
# Formatting
# go fmt ./... finds Python files which we want to ignore.
fmt:
	-go fmt ./pkg/...

# ----------------------------------------------------------------
# Needs first: go install honnef.co/go/tools/cmd/staticcheck@latest
# See also: https://staticcheck.io
staticcheck:
	staticcheck ./...

# ================================================================
# Go does its own dependency management, outside of make.
.PHONY: build check fmt staticcheck dev
