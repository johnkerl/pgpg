#!/usr/bin/env python3
"""
Pemdas-eval: parse PEMDAS arithmetic expressions and evaluate them.
Usage: pemdas_eval.py [options] expr {one or more strings to parse ...}
       pemdas_eval.py [options] file [one or more filenames]  (none = stdin)
"""
from __future__ import annotations

import argparse
import sys
from pathlib import Path

# Add repo root so we can add generators/py and generated/py to path
_REPO_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(_REPO_ROOT / "generators" / "py"))
sys.path.insert(0, str(_REPO_ROOT / "generated" / "py"))


def main() -> int:
    argparser = argparse.ArgumentParser(
        description="Parse and evaluate PEMDAS arithmetic expressions."
    )
    argparser.add_argument(
        "-v", action="store_true", help="Print AST before evaluation"
    )
    argparser.add_argument(
        "mode",
        choices=["expr", "file"],
        help="expr = strings as args; file = read filenames",
    )
    argparser.add_argument(
        "args",
        nargs="*",
        help="Strings to parse (expr) or filenames (file); file with none reads stdin",
    )
    args = argparser.parse_args()

    try:
        if args.mode == "expr":
            if not args.args:
                print("pemdas_eval: expr requires at least one string", file=sys.stderr)
                return 1
            for s in args.args:
                run_once(s, args.v)
        else:
            if not args.args:
                run_once(sys.stdin.read(), args.v)
            else:
                for filename in args.args:
                    path = Path(filename)
                    if not path.exists():
                        print(
                            f"pemdas_eval: {filename}: no such file",
                            file=sys.stderr,
                        )
                        return 1
                    run_once(path.read_text(), args.v)
    except ValueError as e:
        print(f"pemdas_eval: {e}", file=sys.stderr)
        return 1
    return 0


def run_once(input_str: str, verbose: bool) -> None:
    from lexers import pemdas_lexer
    from parsers import pemdas_parser

    lexer = pemdas_lexer.pgpg_PEMDASLexer(input_str)
    parser = pemdas_parser.pgpg_PEMDASParser()
    ast = parser.parse(lexer, "")
    if ast is None:
        raise ValueError("(nil AST)")
    if verbose:
        ast.print_tree()
    v = evaluate_ast_node(ast.root_node)
    print(v)


def evaluate_ast_node(node) -> int:
    """Evaluate a single AST node; returns integer value."""
    if node.type == "int_literal":
        return _evaluate_literal(node)
    if node.type == "operator":
        return _evaluate_binary_operator(node)
    if node.type == "unary":
        return _evaluate_unary_operator(node)
    raise ValueError(f'Unhandled node type "{node.type}"')


def _evaluate_literal(node) -> int:
    if node.token is None:
        raise ValueError("Literal node has no token")
    return int(node.token.lexeme)


def _evaluate_binary_operator(node) -> int:
    op = node.token.lexeme if node.token else ""
    if len(node.children) != 2:
        raise ValueError(
            f'Expected two operands for operator "{op}"; got {len(node.children)}'
        )
    c1 = evaluate_ast_node(node.children[0])
    c2 = evaluate_ast_node(node.children[1])
    if op == "+":
        return c1 + c2
    if op == "-":
        return c1 - c2
    if op == "*":
        return c1 * c2
    if op == "/":
        return c1 // c2
    if op == "%":
        return c1 % c2
    if op == "**":
        if c2 < 0:
            return 0
        return pow(c1, c2)
    raise ValueError(f'Unhandled operator "{op}"')


def _evaluate_unary_operator(node) -> int:
    op = node.token.lexeme if node.token else ""
    if len(node.children) != 1:
        raise ValueError(
            f'Expected one operand for unary "{op}"; got {len(node.children)}'
        )
    v = evaluate_ast_node(node.children[0])
    if op == "+":
        return v
    if op == "-":
        return -v
    raise ValueError(f'Unhandled unary operator "{op}"')


if __name__ == "__main__":
    sys.exit(main())
