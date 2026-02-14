# PGPG Python generated lexers and parsers

Python 3.10+.

This directory holds **generated** lexers and parsers only. Codegen scripts, runtime (Token, AST, AbstractLexer), and tests live in **../generators/py**.

## Layout

- **lexers/** — Generated DFA lexers (from `*-lex.json`).
- **parsers/** — Generated LR(1) parsers (from `*-parse.json`).
- **Makefile** — Targets to generate (via generators/py codegen) and run tests.

## Dependencies

```bash
pip install -r requirements.txt
```

(Jinja2 is required for codegen; install from generators/py if you run codegen directly.)

## Generating parsers and lexers

Use the **same** JSON tables produced by the Go tools (`parsegen-tables`, `lexgen-tables`).

From the **repository root**:

```bash
make -C generated/py json
```

Or from **generated/py/** (output goes into this directory):

```bash
make json
```

To run codegen by hand (from repo root):

```bash
# Parser (default class prefix is pgpg_)
PYTHONPATH=generators/py python3 generators/py/codegen/parsegen_code.py \
  -o generated/py/parsers/json_parser.py -c JSONParser \
  generated/jsons/json-parse.json

# Lexer
PYTHONPATH=generators/py python3 generators/py/codegen/lexgen_code.py \
  -o generated/py/lexers/json_lexer.py -c JSONLexer \
  generated/jsons/json-lex.json
```

- **--prefix** (default: `pgpg_`) is prepended to the class name. Use `--prefix ""` for no prefix.

## Running generated parsers

Set `PYTHONPATH` so that **generators/py** (runtime) and **generated/py** (lexers, parsers) are on the path, e.g. from repo root: `PYTHONPATH=generators/py:generated/py`.

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

Tests live in **generators/py**. From **generated/py/**:

```bash
make test
```

Or from **generators/py/**:

```bash
make -C generators/py test
```

Or run a single test file:

```bash
PYTHONPATH=generators/py python3 generators/py/tests/test_parsegen_code.py -v
```
