I have a long-term project to propose: making acombined lexer and LALR(1) parser generator in Go, named PGPG.

Here are the top-level ideas:

## Goals

* Implement a few basic algorithms.
* Reuse code whenever possible
  * Across multiple algorithms like LALR/LR
* Make good use of classes -- e.g. `lexer.match()` rather than global `match()` which are commonly used in intro-to-parsing textbooks.
* Be lucid above all else. Lexing/parsing is ubiquitous in the modern world, and forms a large part of our world. Yet sadly such tools are too often arcane and confusing. PGPG is transparent, inclusive, and explains itself openly.
* Offer choices.
  * Sometimes a parser-generator is overkill -- for simpler grammars, a hand-written lexer and a hand-written recursive-descent parser are quite satisfactory. PGPG offers reusable, easy-to-understand examples here.
  * Sometimes a hand-written lexer/parser is underkill -- yet parser-generators can be complex and intimidating. Here, too, PGPG offers reusable, easy-to-understand examples.
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
