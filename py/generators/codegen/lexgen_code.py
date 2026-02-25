#!/usr/bin/env python3
"""
Generate a Python DFA lexer from JSON tables (output of lexgen-tables).
Reads the same *-lex.json files as the Go lexgen-code.
Usage:
  python lexgen_code.py -o lexers/json_lexer.py -c JSONLexer [--prefix pgpg_] json-lex.json
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


def load_tables(path: Path) -> dict:
    with open(path, encoding="utf-8") as f:
        return json.load(f)


def build_transitions(raw: dict) -> list[dict]:
    transitions = raw.get("transitions", {})
    out = []
    for state_str in sorted(transitions.keys(), key=int):
        state = int(state_str)
        ranges = []
        for tr in transitions[state_str]:
            ranges.append(
                {
                    "from": tr["from"],
                    "to": tr["to"],
                    "next": tr["next"],
                }
            )
        out.append({"state": state, "ranges": ranges})
    return out


def build_actions(raw: dict) -> list[dict]:
    actions = raw.get("actions", {})
    out = []
    for state_str in sorted(actions.keys(), key=int):
        out.append(
            {
                "state": int(state_str),
                "token_type": actions[state_str],
            }
        )
    return out


def main() -> int:
    ap = argparse.ArgumentParser(
        description="Generate Python lexer from lexgen JSON tables"
    )
    ap.add_argument("json_file", type=Path, help="Path to *-lex.json")
    ap.add_argument("-o", "--output", type=Path, required=True, help="Output .py file")
    ap.add_argument(
        "-c", "--class-name", required=True, help="Lexer class name (e.g. JSONLexer)"
    )
    ap.add_argument(
        "--prefix", default="pgpg_", help="Prefix for class name (default: pgpg_)"
    )
    args = ap.parse_args()

    raw = load_tables(args.json_file)
    start_state = raw.get("start_state", 0)
    actions = build_actions(raw)
    has_ignored = any(a["token_type"].startswith("!") for a in actions)

    class_name = (args.prefix + args.class_name) if args.prefix else args.class_name

    ctx = {
        "class_name": class_name,
        "start_state": start_state,
        "has_ignored": has_ignored,
        "transitions": build_transitions(raw),
        "actions": actions,
    }

    try:
        from jinja2 import Environment, FileSystemLoader
    except ImportError:
        print("jinja2 is required: pip install jinja2", file=sys.stderr)
        return 1

    template_dir = Path(__file__).resolve().parent / "templates"
    env = Environment(
        loader=FileSystemLoader(str(template_dir)), keep_trailing_newline=True
    )
    env.filters["repr"] = repr
    template = env.get_template("lexer.py.j2")
    out_text = template.render(**ctx)

    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(out_text, encoding="utf-8")
    return 0


if __name__ == "__main__":
    sys.exit(main())
