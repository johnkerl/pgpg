# PGPG JavaScript generated lexers and parsers

ES modules; Node 18+ (or modern bundler).

## Layout

- **runtime/** — Token, Location, ASTNode, AST, and helpers (used by generated code).
- **parsers/** — Generated LR(1) parsers (from `*-parse.json`).
- **lexers/** — Generated DFA lexers (from `*-lex.json`).
- **codegen/** — Scripts to generate parsers and lexers (no template dependency).
- **tests/** — Unit tests for codegen.

## Generating parsers and lexers

Use the **same** JSON tables produced by the Go tools (`parsegen-tables`, `lexgen-tables`). No changes to table generation.

From the **repository root**:

```bash
# Parser (default class prefix is pgpg_)
cd generated_js && node codegen/parsegen_code.js \
  -o parsers/json_parser.js -c JSONParser \
  ../generated/jsons/json-parse.json

# No prefix
node codegen/parsegen_code.js -o parsers/json_parser.js -c JSONParser --prefix "" \
  ../generated/jsons/json-parse.json

# Lexer
node codegen/lexgen_code.js -o lexers/json_lexer.js -c JSONLexer \
  ../generated/jsons/json-lex.json
```

- **--prefix** (default: `pgpg_`) is prepended to the class name (e.g. `pgpg_JSONParser`). Use `--prefix ""` for no prefix.

## Running generated parsers

From `generated_js/` (or with module resolution pointing at it):

```javascript
import { pgpg_JSONLexer } from "./lexers/json_lexer.js";
import { pgpg_JSONParser } from "./parsers/json_parser.js";

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

From `generated_js/`:

```bash
node --test tests/
```

Or a single file:

```bash
node --test tests/test_parsegen_code.js
```
