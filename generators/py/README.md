# PGPG Python codegen and runtime

- **codegen/** — Scripts and Jinja2 templates to generate Python lexers and parsers from JSON tables (output of Go lexgen-tables / parsegen-tables).
- **runtime/** — Token, ASTNode, AST, AbstractLexer (used by generated lexers and parsers).
- **tests/** — Unit tests for codegen.

Generated output is written to **../generated_py** (lexers/, parsers/). Use `PYTHONPATH=generators/py` when running codegen so that templates and runtime are found; when running generated code, use `PYTHONPATH=generators/py:generated_py`.

## Dependencies

```bash
pip install -r requirements.txt
```

## Tests

```bash
make test
# or
PYTHONPATH=. python3 -m unittest discover -s tests -v
```
