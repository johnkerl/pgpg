# PGPG Python apps

Python CLI tools that use the generated Python lexers and parsers (from `generated/py`). Run from repo root or with `PYTHONPATH` including `generators/py` and `generated/py`.

## trylex

Run a generated lexer on strings or files.

```bash
python3 apps/py/trylex.py g:json expr '[1, 2]'
python3 apps/py/trylex.py g:pemdas expr '1+2*3'
python3 apps/py/trylex.py g:json file input.txt
```

Lexer names: `g:json`, `g:pemdas`.

## tryparse

Run a generated parser on strings or files.

```bash
python3 apps/py/tryparse.py g:json expr '[1, 2, 3]'
python3 apps/py/tryparse.py g:pemdas expr '1+2*3'
python3 apps/py/tryparse.py -tokens -states -stack g:json expr '{}'
python3 apps/py/tryparse.py -noast g:pemdas expr '42'
```

Options: `-tokens`, `-states`, `-stack`, `-noast`, `-fullast`. Parser names: `g:json`, `g:pemdas`.

## pemdas_eval

Parse PEMDAS arithmetic and print the integer result.

```bash
python3 apps/py/pemdas_eval.py expr '1+2*3'
python3 apps/py/pemdas_eval.py -v expr '10/3'
python3 apps/py/pemdas_eval.py file expr.txt
```

Option `-v`: print AST before evaluation.

## Dependencies

Generated lexers/parsers must exist in `generated/py`. From repo root:

```bash
make -C generated/py all   # json + pemdas
```

Scripts add `generators/py` and `generated/py` to `sys.path` automatically when run from the repo.
