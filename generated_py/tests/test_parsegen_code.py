"""Unit tests for parser codegen: ensure hint 'type' is applied for with_adopted_grandchildren."""
from __future__ import annotations

import json
import sys
import tempfile
import unittest
from pathlib import Path

# Add generated_py so we can import codegen and runtime
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from codegen.parsegen_code import main


class TestParsegenCode(unittest.TestCase):
    """Parser codegen tests."""

    def test_with_adopted_grandchildren_respects_type_in_generated_code(self) -> None:
        """Generated parser must apply hint 'type' in with_adopted_grandchildren branch."""
        tables = {
            "start_symbol": "Root",
            "actions": {"0": {"EOF": {"type": "accept"}}},
            "gotos": {},
            "productions": [
                {
                    "lhs": "Root",
                    "rhs": [
                        {"name": "lbracket", "terminal": True},
                        {"name": "Elements", "terminal": False},
                        {"name": "rbracket", "terminal": True},
                    ],
                    "hint": {
                        "parent_literal": "[]",
                        "with_adopted_grandchildren": [1],
                        "type": "array",
                    },
                },
            ],
            "hint_mode": "hints",
        }
        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".json", delete=False
        ) as f:
            json.dump(tables, f)
            json_path = Path(f.name)
        try:
            out_path = Path(tempfile.gettempdir()) / "test_array_parser_py_pgpg.py"
            argv = [
                "parsegen_code.py",
                "-o", str(out_path),
                "-c", "ArrayParser",
                "--prefix", "",
                str(json_path),
            ]
            old_argv = sys.argv
            sys.argv = argv
            try:
                self.assertEqual(main(), 0)
            finally:
                sys.argv = old_argv
            code = out_path.read_text()
            self.assertIn(
                "node_type = prod.node_type or parent_type",
                code,
                "generated code should apply hint type in with_adopted_grandchildren branch",
            )
            self.assertTrue(
                'node_type="array"' in code or "node_type='array'" in code,
                "generated production table should include node_type for the hint",
            )
            self.assertIn(
                "new_ast_node(parent_token, node_type, new_children)",
                code,
                "generated code should build node with node_type in with_adopted_grandchildren branch",
            )
        finally:
            json_path.unlink(missing_ok=True)
            out_path.unlink(missing_ok=True)


if __name__ == "__main__":
    unittest.main()
