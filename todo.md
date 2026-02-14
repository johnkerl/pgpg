# ASTGEN THING

* `pemdas-eval` with `-v`
* Prohibit any parent/child indices the same
* Port manuals to JavaScript
* Hex literals into some grammars ...
* Simple calculator:
  * Prompt string or no
  * Handle stdin
  * Types: int, float; then modular int, f2polys ...

# Top of list

* Look into custom-AST/custom-traversal options:
  * Impl-lang AST/arbitrary in `.bnf`.
  * DSL AST-building in `.bnf`.
  * Callbacks somehow ... .
* More BNF to match manual grammars.
* Code neatens:
  * Move types to tops of files
  * Dedupe
  * Etc.
* Desk-calculator grammar
  * `kbc` basically: int, float, hex, ibase/obase
  * Also a modular-arithmetic option perhaps
  * Assign to variables; print; target REPL
* JSON, CSV, TSV, DKVP, XTAB grammars
  * See what happens with repeated top-level blocks in the same input stream
* Config-file grammar?
* Manual LISP parser?
* Manual JSON parser?
* Doc `tryparse` with `-tokens`, `-states`, `-stack`.

# Use-cases for me personally

* JSON, CSV, TSV, DKVP, XTAB I/O.
* Config-file grammar---for Miller.
* `kbc` and `spffl`; desk-calculator REPL.
* LISP interpreter.

# Lexer

* [x] Abstract/interface class and datatypes
* [x] Make sure there's file/line/column info
* [x] CLIs for:
  * [x] Run a given lexer on given input text and dump out the sequence of tokens
* Hand-written lexer impls for simple grammars
  * [x] Canned lexer from a fixed list of strings
  * [x] Rune lexer: every rune is its own token
  * [x] Line lexer: every line is its own token
  * [x] Word lexer: delimit by whitespace
  * [x] Make an argv1-switching lex-runner w/ from-text or from-file for various lexers
  * [x] AM for AME and AMNE
  * [x] SENG (lexicon-driven) -- maybe SENG lexer layered atop word-lexer?
    * Prepositional phrases?
  * [x] VIC
  * [x] VBC
  * [x] EBNF
  * [ ] LISP grammar -- minimum viable product to "do things"
  * [ ] Scale-test everything for perf early on -- especially channel-switching
  * Mods needed:
    * [x] Standardize to match/accept/backUp/etc standard names as much as possible
    * [x] Explicit EOF and error tokens
    * [x] Userspace type-codes -- how to handle
      * Some from the inside going out, e.g. hand-written context
      * Some from the outside coming in, e.g. PG context
* [ ] Make sure impls can do _full_ faithful reconstruct of source -- including retention of intervening whitespace
  * [ ] Note that either `Token` struct will have two strings -- payload, and payload+whitespace ...
  * [ ] ... or, there should be a "produce ignore-tokens" option
* [x] Config-driven autogen
* [ ] Make sure grammar -> lexer build can be done either offline or online (the latter without need for process restart)

# Lexer-generator

* Maybe: auto‑include operator literals from parser rules:
  * Scan all grammar rules
  * Collect any literal terminals (e.g., "+", "-", "(", ")", "==", etc.).
  * Add those literals into the lexer rule set automatically, even if no explicit lexer rule exists
* DFA minimization

# AST

* [x] Adapt from Miller
* [x] Unit-test in isolation

# Parser

* [x] Abstract/interface class
* [x] Hand-written recursive-descent impls for simple grammars
* [ ] Hand-written recursive-ascent impls for simple grammars
* [x] Connect to AST populate
* [ ] CLIs for:
  * [x] Online grammar + string -> AST + pass/fail
  * [x] Offline grammar -> intermediate representation
  * [ ] Precomputed grammar + string -> AST + pass/fail
  * [ ] Linked-in grammar + string -> AST + pass/fail
* [x] Iterate on PGs per se

# Infra:

* [ ] AST to-string factored out of printer
  * [ ] Use for UT
* ame/amne UT
* Wrap lexers in LA1 and LA2 etc for lookahead level:
  * Hide direct calls to Scan``
  * `.First()`
  * `.Second()`
  * `.Advance()`

# [ ] Where (which `README.md` etc) to note this is all UTF-8

# [ ] Credits

* The Dragon book
* https://github.com/goccmack/gocc
* https://go.dev/talks/2011/lex.slide#1
* Cursor

# Needs CI!

* For current Go
* For TBD Python
