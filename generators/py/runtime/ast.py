"""AST node and root type for parser output."""
from dataclasses import dataclass, field
from typing import Optional

from runtime.token import (
    Token,
)  # noqa: I001 (runtime is a top-level package when PYTHONPATH=generators/py)


@dataclass
class ASTNode:
    """A single node in the abstract syntax tree."""

    token: Optional[Token]
    type: str  # node_type in Go; 'type' is reserved in some contexts
    children: list["ASTNode"] = field(default_factory=list)

    def __repr__(self) -> str:
        return f"ASTNode({self.type!r}, {len(self.children)} children)"


def new_ast_node(
    token: Optional[Token],
    node_type: str,
    children: list[ASTNode],
) -> ASTNode:
    """Build an AST node (for generated parser use)."""
    return ASTNode(token=token, type=node_type, children=children or [])


def new_ast_node_terminal(token: Token, node_type: str) -> ASTNode:
    """Build a terminal AST node (shift in parser)."""
    return ASTNode(token=token, type=node_type, children=[])


@dataclass
class AST:
    """Root of an abstract syntax tree."""

    root_node: ASTNode

    def print_tree(self, indent: int = 0) -> None:
        """Pretty-print the tree (idiomatic replacement for Go Print())."""
        pad = "    " * indent
        node = self.root_node
        if node.token:
            lexeme = node.token.lexeme.replace("\n", "\\n")
            print(f"{pad}{lexeme!r} [tt:{node.token.type}] [nt:{node.type}]")
        else:
            print(f"{pad}[nt:{node.type}]")
        for child in node.children:
            child_ast = AST(root_node=child)
            child_ast.print_tree(indent + 1)
