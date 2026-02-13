# Generated Python lexers and parsers from JSON tables (produced by Go lexgen-tables / parsegen-tables).
# Run from repo root: make -C generated_py [target]
# Or from here: make JSON_PATH=../generated/jsons ...

JSON_PATH ?= ../generated/jsons

# Generate JSON parser + lexer
json: json-lexer json-parser

json-lexer:
	PYTHONPATH=. python3 codegen/lexgen_code.py \
		-o lexers/json_lexer.py -c JSONLexer \
		$(JSON_PATH)/json-lex.json

json-parser:
	PYTHONPATH=. python3 codegen/parsegen_code.py \
		-o parsers/json_parser.py -c JSONParser \
		$(JSON_PATH)/json-parse.json

# Regenerate all parsers/lexers that have JSON in generated/jsons
all: json
	@echo "Add more grammar targets as needed (e.g. pemdas, lisp)."

test:
	PYTHONPATH=. python3 -m unittest discover -s tests -v

.PHONY: all json json-lexer json-parser test
