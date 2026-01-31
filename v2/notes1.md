I have a long-term project to propose: making a combined lexer and LALR(1)
parser generator in Go.

Core principles
* Clear and lucid for the reader and user.
* Extensible.
* Reuse code across algorithms and implementations.
* Provide both hand-written and generated options.

Project direction
* This does not read a `.bnf` file and directly generate Go code.
* It reads a `.bnf` file and produces an implementation-independent
  intermediate representation (IR).
* A separate tool takes that IR and generates language-specific code.

Target languages
* Initial implementation in Go.
* Initial generator target in Go.
* Future targets: Python and/or Rust.
* Prefer language-independent data structures when possible.

Applications
* Self-education and experimentation.
* Promotion of parser-generation knowledge.
* Future use in Miller.
* Potential for AST-level "sed/grep" tooling for languages like Go/Python/JS.

High-level architecture
* `bnf` reader -> grammar AST.
* Grammar normalization:
  * Epsilon handling.
  * Left-factoring (optional).
  * Left-recursion handling (optional, for LL-style outputs).
* Lexer construction:
  * NFA -> DFA (subset construction).
  * DFA minimization (optional).
  * Longest-match + rule priority.
* Parser construction:
  * LR(0), SLR(1), LALR(1) and (later) LR(1).
  * Table construction + conflict reporting.
* IR output:
  * JSON (first), with stable schema and versioning.
* Code generation:
  * Go (first), then Python/Rust.

Key packages (tentative)
* `grammar`: BNF parser, AST, validation.
* `lexer`: regex AST, NFA/DFA, minimized DFA, token streams.
* `parser`: LR items, FIRST/FOLLOW, parsing tables, conflict diagnostics.
* `ir`: language-independent IR schema + JSON encoder/decoder.
* `codegen`: language-specific code output.
* `cli`: command entry points and high-level workflows.

CLI workflows (tentative)
* `pgpg bnf-to-ir <input.bnf> > grammar.json`
* `pgpg ir-to-go <grammar.json> --out ./parser`
* `pgpg all-in-one <input.bnf> --lang go --out ./parser`

Deliverables & milestones
* M1: grammar AST + BNF reader + JSON IR.
* M2: NFA/DFA lexer with tests.
* M3: LR(0)/SLR(1) parsing tables with diagnostics.
* M4: LALR(1) parsing tables + conflict resolution reports.
* M5: Go codegen for lexer+parser.
* M6: Example projects (hand-written + generated).

Notes on clarity
* Every algorithm has a narrative doc and a small example.
* Provide diagrams (state machines, item sets) alongside code.
* Avoid clever abstractions; prefer explicit data structures.
