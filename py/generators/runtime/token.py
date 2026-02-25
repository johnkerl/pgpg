"""Token and location types for lexer/parser."""
from dataclasses import dataclass
from typing import Optional


@dataclass(frozen=True)
class Location:
    """Source location (line, column, byte offset)."""

    line: int = 1
    column: int = 1
    byte_offset: int = 0


# Sentinel type for lexer errors (same as Go TokenTypeError).
TOKEN_TYPE_ERROR = "!error"
TOKEN_TYPE_EOF = "EOF"


@dataclass
class Token:
    """A single token from the lexer."""

    type: str
    lexeme: str
    location: Optional[Location] = None

    def __repr__(self) -> str:
        loc = (
            f" line={self.location.line} col={self.location.column}"
            if self.location
            else ""
        )
        return f"Token({self.type!r}, {self.lexeme!r}{loc})"


def new_token(
    lexeme: str, token_type: str, location: Optional[Location] = None
) -> Token:
    """Build a token (for parser-constructed tokens, e.g. parent_literal)."""
    return Token(type=token_type, lexeme=lexeme, location=location)


def new_eof_token(location: Optional[Location] = None) -> Token:
    """Build an EOF token."""
    return Token(type=TOKEN_TYPE_EOF, lexeme="", location=location)


def new_error_token(message: str, location: Optional[Location] = None) -> Token:
    """Build an error token (lexer/parser can use this to signal errors)."""
    return Token(type=TOKEN_TYPE_ERROR, lexeme=message, location=location)
