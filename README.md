# pgpg

PGPG is the Pretty Good Parser Generator. As of Feburary 2023 it is very much a sketch and a work in progress.

## Goals

* Implement a few basic algorithms.
* Reuse code whenever possible
  * Across multiple algorithms like LALR/LR
* Make good use of classes -- e.g. `lexer.match()` rather than global `match()` which are commonly used in intro-to-parsing textbooks.
* Be lucid above all else. Lexing/parsing is ubiquitous in the modern world, and forms a large part of our world. Yet sadly such tools are arcane and confusing. PGPG is transparent, inclusive, and explains itself openly.
* Offer choices.
  * Sometimes a parser-generator is overkill -- for simpler grammars, a hand-written lexer and a hand-written recursive-descent parser are quite satisfactory. PGPG offers reusable, easy-to-understand examples here.
  * Sometimes a hand-written lexer/parser is underkill -- yet parser-generators can be complex and intimidatig. Here, too, PGPG offers reusable, easy-to-understand examples.
  * PGPG offers classes that reduce code-duplication for various lex/parse implementations: you can reuse what you want, and hand-write what you want.
  * PGPG offers grammar-to-parser all in one process invocation, or parser-generate to language-independent storage (probably JSON), or traditional parser-generate directly to implementation-language code.

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
* I'd love to get some code of ad-hoc code-sed/code-grep/etc functionality going for, say, Go, Python, JS, etc. wherein program text would be treated as a stream where the "sedding" and "grepping" would be done at the abstract-syntax-tree level. Do note however that Python and JavaScript offer library access to their own parsers, with tree-walk-visitor callbacks which are simply splendid.

## Development

WIP.

* `go test github.com/johnkerl/pgpg/...`
