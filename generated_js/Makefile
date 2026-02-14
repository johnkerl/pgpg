# Generated JavaScript lexers and parsers from JSON tables (produced by Go lexgen-tables / parsegen-tables).
# Run from repo root: make -C generated_js [target]
# Or from here: make JSON_PATH=../generated/jsons ...

JSON_PATH ?= ../generated/jsons

json: json-lexer json-parser

json-lexer:
	node codegen/lexgen_code.js -o lexers/json_lexer.js -c JSONLexer $(JSON_PATH)/json-lex.json

json-parser:
	node codegen/parsegen_code.js -o parsers/json_parser.js -c JSONParser $(JSON_PATH)/json-parse.json

all: json

test:
	node --test tests/test_parsegen_code.js

.PHONY: all json json-lexer json-parser test
