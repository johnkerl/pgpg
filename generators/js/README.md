# PGPG JavaScript codegen and runtime

- **codegen/** — Scripts to generate JavaScript lexers and parsers from JSON tables (output of Go lexgen-tables / parsegen-tables). No template dependency.
- **runtime/** — Token, Location, ASTNode, AST (used by generated lexers and parsers).
- **tests/** — Unit tests for codegen.

Generated output is written to **../generated/js** (lexers/, parsers/). Generated files (under generated/js/) import from `../../../generators/js/runtime/index.js`, so run from repo root or ensure that path resolves.

## Tests

```bash
make test
# or
node --test tests/
```
