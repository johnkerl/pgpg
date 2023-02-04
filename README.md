# pgpg

PGPG is the Pretty Good Parser Generator. As of Feburary 2023 it is very much a sketch and a work in progress.

## Goals

* Implement a few basic algorithms.
* Reuse code whenever possible
  * Across multiple algorithms like LALR/LR
  * And also make classes that reduce code-duplication even for hand-written recursive-descent parsers
* Make good use of classes -- e.g. `lexer.match()` rather than global `match()`.
* Be lucid above all else. Lexing/parsing is ubiquitous in the modern world, and forms a large part of our world. Yet sadly such tools are arcane and confusing. PGPG is transparent, inclusive, and explains itself openly.

## Languages

* Implementation initially in Go
  * Maybe Python and/or Rust later
  * Aim for non-clever abstraction and concept reuse
  * Try to use language-independent data structures when possible
* Generator target initially Go
  * Maybe Python and/or Rust later
  * Try to use language-independent data structures when possible

## Applications

* Self-education and experimentation
* Promotion of parser-generation knowledge
* I would like to ultimately use this in [Miller](https://github.com/johnkerl/miller)
* I'd love to get some code of ad-hoc code-sed/code-grep/etc functionality going for, say, Go, Python, JS, etc. wherein program text would be treated as a stream where the "sedding" and "grepping" would be done at the abstract-syntax-tree level

## Development

WIP.

* `go test github.com/johnkerl/pgpg/...`
