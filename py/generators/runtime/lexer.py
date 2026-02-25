"""Lexer protocol for parser input."""
from typing import Optional, Protocol

from runtime.token import Token  # noqa: I001


class AbstractLexer(Protocol):
    """Protocol for lexers: one method returning the next token or None at EOF."""

    def scan(self) -> Optional[Token]:
        """Return the next token, or None if at end of input (EOF already consumed)."""
        ...
