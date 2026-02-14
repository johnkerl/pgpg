# PGPG JavaScript generated lexers and parsers

ES modules; Node 18+ (or modern bundler).

This directory holds **generated** lexers and parsers only. Codegen scripts and runtime (Token, AST, AbstractLexer) live in **../generators/js**.

## Layout

- **lexers/** — Generated DFA lexers (from `*-lex.json`).
- **parsers/** — Generated LR(1) parsers (from `*-parse.json`).
- **Makefile** — Targets to generate (via generators/js codegen) and run tests.

## Generating parsers and lexers

Use the **same** JSON tables produced by the Go tools (`parsegen-tables`, `lexgen-tables`).

From the **repository root**:

```bash
make -C generated/js json
```

Or from **generated/js/**:

```bash
make json
```

To run codegen by hand (from repo root, or from generated/js with paths adjusted):

```bash
# Parser (default class prefix is pgpg_)
node generators/js/codegen/parsegen_code.js \
  -o generated/js/parsers/json_parser.js -c JSONParser \
  generated/jsons/json-parse.json

# Lexer
node generators/js/codegen/lexgen_code.js \
  -o generated/js/lexers/json_lexer.js -c JSONLexer \
  generated/jsons/json-lex.json
```

- **--prefix** (default: `pgpg_`) is prepended to the class name. Use `--prefix ""` for no prefix.

## Running generated parsers

Generated lexers and parsers import the runtime from `../../../generators/js/runtime/index.js`, so run from the **repository root** (or ensure Node resolves that path). Example from repo root:

```javascript
import { pgpg_JSONLexer } from "./generated/js/lexers/json_lexer.js";
import { pgpg_JSONParser } from "./generated/js/parsers/json_parser.js";

const lex = new pgpg_JSONLexer("[1]");
const parser = new pgpg_JSONParser();
const ast = parser.parse(lex, ""); // "" | "noast" | "fullast"
if (ast) ast.printTree();
```

- **ast_mode**: `""` = use grammar hints or full tree; `"noast"` = syntax only (returns `null`); `"fullast"` = ignore hints, full parse tree.

## Tracing

```javascript
parser.attachCLITrace(true, true, true); // traceTokens, traceStates, traceStack
const ast = parser.parse(lex, "");
```

## Tests

Tests live in **generators/js**. From **generated/js/**:

```bash
make test
```

Or from **generators/js/**:

```bash
make -C generators/js test
```

Or run a single test file:

```bash
node --test generators/js/tests/test_parsegen_code.js
```
