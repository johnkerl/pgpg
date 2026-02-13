# PGPG Python generated lexers and parsers

Python 3.10+.

## Layout

- **runtime/** — Token, ASTNode, AST, AbstractLexer (used by generated code).
- **parsers/** — Generated LR(1) parsers (from `*-parse.json`).
- **lexers/** — Generated DFA lexers (from `*-lex.json`).
- **codegen/** — Scripts and Jinja2 templates to generate parsers and lexers.
- **tests/** — Unit tests for codegen.

## Dependencies

```bash
pip install -r requirements.txt
```

## Generating parsers and lexers

Use the **same** JSON tables produced by the Go tools (`parsegen-tables`, `lexgen-tables`). No changes to table generation.

From the **repository root** (so paths to JSON and output are correct):

```bash
# Parser (default class prefix is pgpg_)
PYTHONPATH=generated_py python3 generated_py/codegen/parsegen_code.py \
  -o generated_py/parsers/json_parser.py -c JSONParser \
  ../generated/jsons/json-parse.json

# No prefix
PYTHONPATH=generated_py python3 generated_py/codegen/parsegen_code.py \
  -o generated_py/parsers/json_parser.py -c JSONParser --prefix "" \
  ../generated/jsons/json-parse.json

# Lexer
PYTHONPATH=generated_py python3 generated_py/codegen/lexgen_code.py \
  -o generated_py/lexers/json_lexer.py -c JSONLexer \
  ../generated/jsons/json-lex.json
```

- **--prefix** (default: `pgpg_`) is prepended to the class name (e.g. `pgpg_JSONParser`). Use `--prefix ""` for no prefix.

## Running generated parsers

Set `PYTHONPATH` so that `runtime`, `parsers`, and `lexers` are importable (e.g. `PYTHONPATH=generated_py` when running from repo root):

```python
from lexers.json_lexer import pgpg_JSONLexer
from parsers.json_parser import pgpg_JSONParser

lex = pgpg_JSONLexer('[1]')
parser = pgpg_JSONParser()
ast = parser.parse(lex, ast_mode="")  # "" | "noast" | "fullast"
if ast:
    ast.print_tree()
```

- **ast_mode**: `""` = use grammar hints or full tree; `"noast"` = syntax only (returns `None`); `"fullast"` = ignore hints, full parse tree.

## Tracing

```python
parser.attach_cli_trace(trace_tokens=True, trace_states=True, trace_stack=True)
ast = parser.parse(lex, "")
```

## Tests

From `generated_py/`:

```bash
python3 -m unittest discover -s tests -v
```

Or run a single test file:

```bash
python3 tests/test_parsegen_code.py -v
```
