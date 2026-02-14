#!/usr/bin/env python3
"""
Tryparse: run a generated Python parser on expr strings or files.
Usage: tryparse.py [options] {parser name} expr {one or more strings to parse ...}
       tryparse.py [options] {parser name} file {one or more filenames}
"""
from __future__ import annotations

import argparse
import sys
from pathlib import Path
from typing import Callable, Optional

# Add repo root so we can add generators/py and generated/py to path
_REPO_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(_REPO_ROOT / "generators" / "py"))
sys.path.insert(0, str(_REPO_ROOT / "generated" / "py"))

from lexers import json_lexer
from lexers import pemdas_lexer
from parsers import json_parser
from parsers import pemdas_parser


def main() -> int:
    parsers_help = {
        "g:json": "Generated JSON parser from bnfs/json.bnf.",
        "g:pemdas": "Generated PEMDAS parser from bnfs/pemdas.bnf.",
    }

    argparser = argparse.ArgumentParser(
        description="Run a generated parser on expr strings or files.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="Parser names:\n"
        + "\n".join(f"  {k:<10} {v}" for k, v in sorted(parsers_help.items())),
    )
    argparser.add_argument(
        "-tokens", action="store_true", help="Print tokens as they're read"
    )
    argparser.add_argument(
        "-states", action="store_true", help="Show parser state transitions"
    )
    argparser.add_argument(
        "-stack", action="store_true", help="Show parser stack after each action"
    )
    argparser.add_argument(
        "-noast", action="store_true", help="Syntax-only: do not build or print AST"
    )
    argparser.add_argument(
        "-fullast",
        action="store_true",
        help="Ignore AST hints and build full parse tree",
    )
    argparser.add_argument(
        "parser_name", choices=list(parsers_help), help="Parser to use"
    )
    argparser.add_argument(
        "mode",
        choices=["expr", "file"],
        help="expr = strings as args; file = read filenames",
    )
    argparser.add_argument(
        "args", nargs="+", help="Strings to parse (expr) or filenames (file)"
    )
    args = argparser.parse_args()

    if args.noast and args.fullast:
        print("cannot use -noast and -fullast together", file=sys.stderr)
        return 1

    ast_mode = "noast" if args.noast else ("fullast" if args.fullast else "")

    if args.parser_name == "g:json":
        parser = make_json_parser(args.tokens, args.states, args.stack, ast_mode)
    else:
        parser = make_pemdas_parser(args.tokens, args.states, args.stack, ast_mode)

    if args.mode == "expr":
        for arg in args.args:
            try:
                run_parser_once(parser, arg)
            except ValueError as e:
                print(f"tryparse: {e}", file=sys.stderr)
                return 1
    else:
        for filename in args.args:
            path = Path(filename)
            if not path.exists():
                print(f"tryparse: {filename}: no such file", file=sys.stderr)
                return 1
            try:
                run_parser_once(parser, path.read_text())
            except ValueError as e:
                print(f"tryparse: {e}", file=sys.stderr)
                return 1
    return 0


def make_json_parser(
    trace_tokens: bool, trace_states: bool, trace_stack: bool, ast_mode: str
):
    def parser(input: str):
        lex = json_lexer.pgpg_JSONLexer(input)
        p = json_parser.pgpg_JSONParser()
        p.attach_cli_trace(
            trace_tokens=trace_tokens,
            trace_states=trace_states,
            trace_stack=trace_stack,
        )
        return p.parse(lex, ast_mode=ast_mode)

    return parser


def make_pemdas_parser(
    trace_tokens: bool, trace_states: bool, trace_stack: bool, ast_mode: str
):
    def parser(input: str):
        lex = pemdas_lexer.pgpg_PEMDASLexer(input)
        p = pemdas_parser.pgpg_PEMDASParser()
        p.attach_cli_trace(
            trace_tokens=trace_tokens,
            trace_states=trace_states,
            trace_stack=trace_stack,
        )
        return p.parse(lex, ast_mode=ast_mode)

    return parser


def run_parser_once(
    parser: Callable[[str], Optional[object]],
    input_str: str,
) -> None:
    """Run parser on one string; print input and AST."""
    print(input_str)
    ast = parser(input_str)
    if ast is not None:
        ast.print_tree()


if __name__ == "__main__":
    sys.exit(main())
