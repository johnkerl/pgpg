# Overview

PGPG is a combined lexer + LALR(1) parser generator focused on clarity and
reusability. The implementation is intentionally explicit, with small,
documented steps that are easy to follow and test.

## Components

- `grammar`: BNF reader + grammar AST.
- `lexer`: regex AST, NFA/DFA, token stream.
- `parser`: LR items, FIRST/FOLLOW, parsing tables.
- `ir`: JSON-first, language-independent representation.
- `codegen`: language-specific code output.
- `cli`: command entry points and workflows.

## Design notes

- Start with small, testable algorithms.
- Favor explicit data structures over clever abstractions.
- Each algorithm is accompanied by a narrative doc and example.
