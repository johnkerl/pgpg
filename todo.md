* [x] Markdown defs of simple grammars
* Lexer
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
    * [ ] VIC
    * [ ] Scale-test everything for perf early on -- especially channel-switching
    * Mods needed:
      * [ ] Standardize to match/accept/backUp/etc standard names as much as possible
      * [x] Explicit EOF and error tokens
      * [ ] Userspace type-codes -- how to handle
        * Some from the inside going out, e.g. hand-written context
        * Some from the outside coming in, e.g. PG context
  * [ ] Make sure impls can do _full_ faithful reconstruct of source -- including retention of intervening whitespace
    * [ ] Note that either `Token` struct will have two strings -- payload, and payload+whitespace ...
    * [ ] ... or, there should be a "produce ignore-tokens" bool flag
  * [ ] Config-driven autogen
  * [ ] Make sure grammar -> lexer build can be done either offline or online (the latter without need for process restart)
* AST
  * [ ] Adapt from Miller
  * [ ] Unit-test in isolation
* Parser
  * [ ] Abstract/interface class
  * [ ] Hand-written recursive-descent impls for simple grammars
  * [ ] Connect to AST populate
  * [ ] CLIs for:
    * [ ] Online grammar + string -> AST + pass/fail
    * [ ] Offline grammar -> intermediate representation
    * [ ] Precomputed grammar + string -> AST + pass/fail
    * [ ] Linked-in grammar + string -> AST + pass/fail
  * [ ] Iterate on PGs per se

* [ ] Where (which `README.md` etc) to note this is all UTF-8
* [ ] Credits
  * The Dragon book
  * goccmack/gocc
  * https://go.dev/talks/2011/lex.slide#1

* Needs CI!
  * For current Go
  * For TBD Python
