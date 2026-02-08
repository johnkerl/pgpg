# Lexgen: Regex/NFA/DFA Lexer Generation

This document captures the design choices for the regex-capable lexer generator
implemented in `generator/pkg/lexgen`.

## Goals

- Support lexer rules that use identifiers and repetition (`{ ... }` / `*`).
- Use Thompson NFA construction followed by subset DFA construction.
- Do not perform DFA minimization (deferred).
- Keep generated tables compact by using rune ranges.
- Preserve longest-match behavior and deterministic rule priority.

## Inputs and Rule Selection

- Input grammar is EBNF parsed by `manual/pkg/parsers` (same as before).
- Lexer rules are selected by name:
  - Names starting with `_` or a lowercase letter are treated as lexer rules.
  - Names starting with `!` are treated as lexer rules that should be ignored.
  - All other rules are assumed to be parser rules (e.g., `Root`).
- Identifiers used inside lexer rules must reference other lexer rules.
  - This avoids accidentally pulling parser rules into the lexer.

## Regex AST

EBNF nodes are mapped into a small regex AST:

- `Literal` → literal node
- `Sequence` → concatenation
- `Alternates` → alternation
- `Optional` → `?`
- `Repeat` → `*`
- `Identifier` → inlined reference to another lexer rule

Recursive lexer rule references are rejected to avoid infinite expansion.

## NFA Construction

Each lexer rule becomes an epsilon‑NFA via standard Thompson construction:

- Literal strings become chains of transitions.
- Alternation uses split/merge with epsilon edges.
- Concatenation wires accepts to the next fragment’s start.
- Optional uses an epsilon bypass.
- Star uses an epsilon loop.

A global start state has epsilon edges to each rule’s NFA start.
Accepting NFA states are annotated with `(ruleName, priority)`.

Rule priority is determined by rule order in the grammar:
earlier rules win on tie.

## DFA Construction (No Minimization)

Subset construction is used:

- Each DFA state is a set of NFA states (epsilon-closure).
- DFA transitions are built per rune based on NFA transitions.
- Accepting DFA states pick the best rule by priority.

DFA minimization is intentionally skipped.

## Range-Based Transitions

The tables schema uses inclusive rune ranges:

```
type RangeTransition struct {
    From rune
    To   rune
    Next int
}
```

During DFA build, per-rune transitions are merged into ranges when:
- runes are consecutive, and
- the target DFA state is the same.

This keeps tables compact (e.g., `0-9` as one range).

## Generated Lexer Behavior

The generated lexer:

- Scans forward, tracking the last accepting DFA state.
- Returns the longest match (same as the previous literal-only lexer).
- Uses a range-based transition lookup:
  - ranges are sorted by `From`
  - lookup stops early if `r < From`
- Ignores tokens whose rule name starts with `!` (they are lexed but not emitted).

## Tables JSON Schema

The `Tables` struct now includes range transitions and optional rule metadata:

```
type Tables struct {
    StartState  int
    Transitions map[int][]RangeTransition
    Actions     map[int]string
    Rules       map[string]string // optional regex-like form
    Metadata    map[string]string // optional
}
```

`Actions` maps accepting DFA states to token types (rule names).

## Notes and Future Work

- DFA minimization can be added later to reduce table size.
- Range lookup can be optimized (e.g., binary search) if needed.
- If the lexer-rule selection heuristic is too coarse, add explicit
  annotations or a `--lexer` flag to select rules.
