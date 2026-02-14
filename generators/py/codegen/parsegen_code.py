#!/usr/bin/env python3
"""
Generate a Python LR(1) parser from JSON tables (output of parsegen-tables).
Reads the same *-parse.json files as the Go parsegen-code.
Usage:
  python parsegen_code.py -o parsers/json_parser.py -c JSONParser [--prefix pgpg_] json-parse.json
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


def load_tables(path: Path) -> dict:
    with open(path, encoding="utf-8") as f:
        return json.load(f)


def build_actions(raw: dict) -> list[dict]:
    actions = raw.get("actions", {})
    out = []
    for state_str in sorted(actions.keys(), key=int):
        state = int(state_str)
        entries = []
        for term in sorted(actions[state_str].keys()):
            act = actions[state_str][term]
            kind = act["type"]
            target = act.get("target", 0)
            kind_literal = f"ActionKind.SHIFT" if kind == "shift" else (
                f"ActionKind.REDUCE" if kind == "reduce" else "ActionKind.ACCEPT"
            )
            entries.append({
                "terminal_literal": repr(term),
                "kind_literal": kind_literal,
                "target": target,
                "has_target": kind in ("shift", "reduce"),
            })
        out.append({"state": state, "entries": entries})
    return out


def build_gotos(raw: dict) -> list[dict]:
    gotos = raw.get("gotos", {})
    out = []
    for state_str in sorted(gotos.keys(), key=int):
        state = int(state_str)
        entries = []
        for nonterm in sorted(gotos[state_str].keys()):
            target = gotos[state_str][nonterm]
            entries.append({
                "nonterm_literal": repr(nonterm),
                "target": target,
            })
        out.append({"state": state, "entries": entries})
    return out


def _list_or_empty(v: list | None) -> list:
    return v if v is not None else []


def build_productions(raw: dict) -> list[dict]:
    productions = raw.get("productions", [])
    out = []
    for prod in productions:
        rhs = prod.get("rhs", [])
        rhs_count = len(rhs)
        hint = prod.get("hint")
        info = {
            "lhs_literal": repr(prod["lhs"]),
            "rhs_count": rhs_count,
            "has_hint": False,
            "has_passthrough": False,
            "has_parent_literal": False,
            "has_with_appended_children": False,
            "has_with_prepended_children": False,
            "has_with_adopted_grandchildren": False,
            "parent_index": 0,
            "passthrough_index": 0,
            "parent_literal": "",
            "child_indices": [],
            "with_appended_children": [],
            "with_prepended_children": [],
            "with_adopted_grandchildren": [],
            "node_type": "",
        }
        if hint is not None:
            if hint.get("pass-through") is not None:
                info["has_passthrough"] = True
                info["passthrough_index"] = hint["pass-through"]
            else:
                info["has_hint"] = True
                if hint.get("parent_literal") is not None:
                    info["has_parent_literal"] = True
                    info["parent_literal"] = hint["parent_literal"]
                else:
                    info["parent_index"] = hint.get("parent", 0)
                info["child_indices"] = _list_or_empty(hint.get("children"))
                info["with_appended_children"] = _list_or_empty(hint.get("with_appended_children"))
                info["has_with_appended_children"] = len(info["with_appended_children"]) > 0
                info["with_prepended_children"] = _list_or_empty(hint.get("with_prepended_children"))
                info["has_with_prepended_children"] = len(info["with_prepended_children"]) > 0
                info["with_adopted_grandchildren"] = _list_or_empty(hint.get("with_adopted_grandchildren"))
                info["has_with_adopted_grandchildren"] = len(info["with_adopted_grandchildren"]) > 0
                info["node_type"] = hint.get("type", "")
        out.append(info)
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description="Generate Python parser from parsegen JSON tables")
    ap.add_argument("json_file", type=Path, help="Path to *-parse.json")
    ap.add_argument("-o", "--output", type=Path, required=True, help="Output .py file")
    ap.add_argument("-c", "--class-name", required=True, help="Parser class name (e.g. JSONParser)")
    ap.add_argument("--prefix", default="pgpg_", help="Prefix for class name (default: pgpg_)")
    args = ap.parse_args()

    raw = load_tables(args.json_file)
    hint_mode = raw.get("hint_mode", "")

    class_name = args.prefix + args.class_name if args.prefix else args.class_name

    ctx = {
        "class_name": class_name,
        "actions": build_actions(raw),
        "gotos": build_gotos(raw),
        "productions": build_productions(raw),
        "hint_mode": hint_mode,
    }

    try:
        from jinja2 import Environment, FileSystemLoader
    except ImportError:
        print("jinja2 is required: pip install jinja2", file=sys.stderr)
        return 1

    template_dir = Path(__file__).resolve().parent / "templates"
    env = Environment(loader=FileSystemLoader(str(template_dir)), keep_trailing_newline=True)
    env.filters["repr"] = repr

    template = env.get_template("parser.py.j2")
    out_text = template.render(**ctx)

    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(out_text, encoding="utf-8")
    return 0


if __name__ == "__main__":
    sys.exit(main())
