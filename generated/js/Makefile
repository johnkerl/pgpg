# Generated JavaScript lexers and parsers from JSON tables (produced by Go lexgen-tables / parsegen-tables).
# Run from repo root: make -C generated/js [target]
# Codegen and runtime live in ../../generators/js.

GEN_JS ?= ../../generators/js
JSON_PATH ?= ../jsons

json: json-lexer json-parser

json-lexer:
	node $(GEN_JS)/codegen/lexgen_code.js -o lexers/json_lexer.js -c JSONLexer $(JSON_PATH)/json-lex.json

json-parser:
	node $(GEN_JS)/codegen/parsegen_code.js -o parsers/json_parser.js -c JSONParser $(JSON_PATH)/json-parse.json

all: json

test:
	$(MAKE) -C $(GEN_JS) test

.PHONY: all json json-lexer json-parser test
