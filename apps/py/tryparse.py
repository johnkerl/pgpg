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


def run_parser_once(
    run: Callable[[str], Optional[object]],
    input_str: str,
) -> None:
    """Run parser on one string; print input and AST."""
    print(input_str)
    ast = run(input_str)
    if ast is not None:
        ast.print_tree()


def main() -> int:
    from lexers import json_lexer
    from lexers import pemdas_lexer
    from parsers import json_parser
    from parsers import pemdas_parser

    def make_run_json(
        trace_tokens: bool, trace_states: bool, trace_stack: bool, ast_mode: str
    ):
        def run(s: str):
            lex = json_lexer.pgpg_JSONLexer(s)
            p = json_parser.pgpg_JSONParser()
            p.attach_cli_trace(
                trace_tokens=trace_tokens,
                trace_states=trace_states,
                trace_stack=trace_stack,
            )
            return p.parse(lex, ast_mode=ast_mode)

        return run

    def make_run_pemdas(
        trace_tokens: bool, trace_states: bool, trace_stack: bool, ast_mode: str
    ):
        def run(s: str):
            lex = pemdas_lexer.pgpg_PEMDASLexer(s)
            p = pemdas_parser.pgpg_PEMDASParser()
            p.attach_cli_trace(
                trace_tokens=trace_tokens,
                trace_states=trace_states,
                trace_stack=trace_stack,
            )
            return p.parse(lex, ast_mode=ast_mode)

        return run

    parsers_help = {
        "g:json": "Generated JSON parser from bnfs/json.bnf.",
        "g:pemdas": "Generated PEMDAS parser from bnfs/pemdas.bnf.",
    }

    parser = argparse.ArgumentParser(
        description="Run a generated parser on expr strings or files.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="Parser names:\n"
        + "\n".join(f"  {k:<10} {v}" for k, v in sorted(parsers_help.items())),
    )
    parser.add_argument(
        "-tokens", action="store_true", help="Print tokens as they're read"
    )
    parser.add_argument(
        "-states", action="store_true", help="Show parser state transitions"
    )
    parser.add_argument(
        "-stack", action="store_true", help="Show parser stack after each action"
    )
    parser.add_argument(
        "-noast", action="store_true", help="Syntax-only: do not build or print AST"
    )
    parser.add_argument(
        "-fullast",
        action="store_true",
        help="Ignore AST hints and build full parse tree",
    )
    parser.add_argument("parser_name", choices=list(parsers_help), help="Parser to use")
    parser.add_argument(
        "mode",
        choices=["expr", "file"],
        help="expr = strings as args; file = read filenames",
    )
    parser.add_argument(
        "args", nargs="+", help="Strings to parse (expr) or filenames (file)"
    )
    args = parser.parse_args()

    if args.noast and args.fullast:
        print("cannot use -noast and -fullast together", file=sys.stderr)
        return 1

    ast_mode = "noast" if args.noast else ("fullast" if args.fullast else "")

    if args.parser_name == "g:json":
        run = make_run_json(args.tokens, args.states, args.stack, ast_mode)
    else:
        run = make_run_pemdas(args.tokens, args.states, args.stack, ast_mode)

    if args.mode == "expr":
        for s in args.args:
            try:
                run_parser_once(run, s)
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
                run_parser_once(run, path.read_text())
            except ValueError as e:
                print(f"tryparse: {e}", file=sys.stderr)
                return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
