# PGPG

PGPG is a combined lexer and LALR(1) parser generator with an emphasis on
clarity, reuse, and approachable documentation. It targets Go first and is
designed to produce a language-independent intermediate representation (IR)
that can later be turned into language-specific code.

## Goals

- Implement a few core algorithms in a clear, reusable way.
- Reuse code across multiple algorithms (LR/LALR) and outputs.
- Use types and methods to keep code local and readable.
- Be lucid above all else: transparent, inclusive, and self-explanatory.
- Offer choices:
  - Hand-written lexer + recursive-descent parser for simpler grammars.
  - Generated lexer + parser for larger or more complex grammars.
  - Shared building blocks so you can mix hand-written and generated code.
  - One-shot generation or IR-based multi-step workflows.

## Languages

- Implementation: Go first; possibly Python and/or Rust later.
- Generator targets: Go first; possibly Python and/or Rust later.
- Prefer language-independent data structures when possible.

## Applications

- Self-education and experimentation.
- Promotion of parser-generation knowledge.
- Intended future use in [Miller](https://github.com/johnkerl/miller).

## High-level architecture

- `bnf` reader -> grammar AST.
- Grammar normalization:
  - Epsilon handling.
  - Left-factoring (optional).
  - Left-recursion handling (optional, for LL-style outputs).
- Lexer construction:
  - NFA -> DFA (subset construction).
  - DFA minimization (optional).
  - Longest-match + rule priority.
- Parser construction:
  - LR(0), SLR(1), LALR(1), then LR(1).
  - Table construction + conflict reporting.
- IR output:
  - JSON first, with stable schema + versioning.
- Code generation:
  - Go first, then Python/Rust.

## Tentative packages

- `grammar`: BNF reader, AST, validation.
- `lexer`: regex AST, NFA/DFA, minimized DFA, token streams.
- `parser`: LR items, FIRST/FOLLOW, parsing tables, diagnostics.
- `ir`: language-independent IR schema + JSON codec.
- `codegen`: language-specific output.
- `cli`: command entry points and workflows.

## Tentative CLI workflows

- `pgpg bnf-to-ir <input.bnf> > grammar.json`
- `pgpg ir-to-go <grammar.json> --out ./parser`
- `pgpg all-in-one <input.bnf> --lang go --out ./parser`

## Deliverables and milestones

- M1: grammar AST + BNF reader + JSON IR.
- M2: NFA/DFA lexer with tests.
- M3: LR(0)/SLR(1) parsing tables with diagnostics.
- M4: LALR(1) parsing tables + conflict reports.
- M5: Go codegen for lexer + parser.
- M6: Example projects (hand-written + generated).

## Notes on clarity

- Each algorithm has a narrative doc and a small example.
- Provide diagrams (state machines, item sets) alongside code.
- Avoid clever abstractions; prefer explicit data structures.

## Docs

- `docs/overview.md`
- `docs/algorithms/lexer-nfa-dfa.md`
- `docs/algorithms/lalr.md`

## Example

Generate IR from a tiny arithmetic grammar:

```
pgpg bnf-to-ir examples/arith.bnf > examples/arith.out.json
```

Generate a stub parser from a minimal IR with embedded tables:

```
pgpg ir-to-go examples/minimal.json --out /tmp/miniparser
```
