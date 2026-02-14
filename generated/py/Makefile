# Generated Python lexers and parsers from JSON tables (produced by Go lexgen-tables / parsegen-tables).
# Run from repo root: make -C generated/py [target]
# Or from here: make JSON_PATH=../jsons ...
# Codegen and runtime live in ../../generators/py.

GEN_PY ?= ../../generators/py
JSON_PATH ?= ../jsons

# Generate JSON parser + lexer
json: json-lexer json-parser

json-lexer:
	PYTHONPATH=$(GEN_PY) python3 $(GEN_PY)/codegen/lexgen_code.py \
		-o lexers/json_lexer.py -c JSONLexer \
		$(JSON_PATH)/json-lex.json

json-parser:
	PYTHONPATH=$(GEN_PY) python3 $(GEN_PY)/codegen/parsegen_code.py \
		-o parsers/json_parser.py -c JSONParser \
		$(JSON_PATH)/json-parse.json

pemdas: pemdas-lexer pemdas-parser

pemdas-lexer:
	PYTHONPATH=$(GEN_PY) python3 $(GEN_PY)/codegen/lexgen_code.py \
		-o lexers/pemdas_lexer.py -c PEMDASLexer \
		$(JSON_PATH)/pemdas-lex.json

pemdas-parser:
	PYTHONPATH=$(GEN_PY) python3 $(GEN_PY)/codegen/parsegen_code.py \
		-o parsers/pemdas_parser.py -c PEMDASParser \
		$(JSON_PATH)/pemdas-parse.json

# Regenerate all parsers/lexers that have JSON in generated/jsons
all: json pemdas
	@echo "Add more grammar targets as needed (e.g. lisp)."

test:
	$(MAKE) -C $(GEN_PY) test

fmt:
	black .

.PHONY: all json json-lexer json-parser pemdas pemdas-lexer pemdas-parser test fmt
