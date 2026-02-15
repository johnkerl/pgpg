#!/usr/bin/env python3
"""
Trylex: run a generated Python lexer on expressions or files.
Usage: trylex.py {lexer name} [-e] [file ...]
  With -e: one or more arguments are expressions to lex (error if none).
  Without -e: zero arguments = read from stdin; one or more = read from those files.
"""
from __future__ import annotations

import argparse
import sys
from pathlib import Path

# Add repo root so we can add generators/py and generated/py to path
_REPO_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(_REPO_ROOT / "generators" / "py"))
sys.path.insert(0, str(_REPO_ROOT / "generated" / "py"))

from runtime.token import TOKEN_TYPE_EOF, TOKEN_TYPE_ERROR
from lexers import json_lexer
from lexers import pemdas_lexer


def main() -> int:
    lexers = {
        "g:json": (
            lambda s: json_lexer.pgpg_JSONLexer(s),
            "Generated JSON lexer from bnfs/json.bnf.",
        ),
        "g:pemdas": (
            lambda s: pemdas_lexer.pgpg_PEMDASLexer(s),
            "Generated PEMDAS lexer from bnfs/pemdas.bnf.",
        ),
    }

    argparser = argparse.ArgumentParser(
        description="Run a generated lexer on expr strings or files.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="Lexer names:\n"
        + "\n".join(f"  {k:<10} {v[1]}" for k, v in sorted(lexers.items())),
    )
    argparser.add_argument(
        "-e",
        action="store_true",
        help="Arguments are expressions to lex (at least one required)",
    )
    argparser.add_argument(
        "lexer_name", choices=list(lexers.keys()), help="Lexer to use"
    )
    argparser.add_argument(
        "args",
        nargs="*",
        help="Expressions (-e) or filenames; with no -e and no args, read stdin",
    )
    args = argparser.parse_args()

    maker, _ = lexers[args.lexer_name]

    if args.e:
        if not args.args:
            print("trylex: -e requires at least one argument", file=sys.stderr)
            return 1
        for s in args.args:
            run_lexer(maker(s))
    else:
        if not args.args:
            run_lexer(maker(sys.stdin.read()))
        else:
            for filename in args.args:
                path = Path(filename)
                if not path.exists():
                    print(f"trylex: {filename}: no such file", file=sys.stderr)
                    return 1
                for line in path.read_text().splitlines():
                    run_lexer(maker(line))
    return 0


def run_lexer(lexer) -> None:
    """Print tokens from lexer until EOF or error."""
    while True:
        token = lexer.scan()
        loc = token.location
        line = loc.line if loc else 0
        col = loc.column if loc else 0
        print(
            f"Line {line:4d} column {col:4d} type {token.type:<16} token <<{token.lexeme!s}>>"
        )
        if token.type == TOKEN_TYPE_EOF or token.type == TOKEN_TYPE_ERROR:
            break


if __name__ == "__main__":
    sys.exit(main())
