* Markdown defs of simple grammars
* Lexer
  * Abstract/interface class and datatypes
  * Make sure there's file/line/column info
  * Make sure impls can do _full_ faithful reconstruct of source -- including retention of intervening whitespace
    * Note `Token` struct will have two strings -- payload, and payload+whitespace
  * Hand-written lexer impls for simple grammars
  * Config-driven autogen
  * Make sure grammar -> lexer build can be done offline or online (the latter without need for process restart)
* AST
  * Adapt from Miller
* Parser
  * Abstract/interface class
  * Hand-written recursive-descent impls for simple grammars
  * Connect to AST populate
  * CLIs for:
    * Online grammar + string -> AST + pass/fail
    * Offline grammar -> intermediate representation
    * Precomputed grammar + string -> AST + pass/fail
    * Linked-in grammar + string -> AST + pass/fail
  * Iterate on PGs per se
* Grammars
  * AME:  `+` and `*` of numbers with equal precedence
  * AMNE: `+` and `*` of numbers with unequal precedence
  * SENG: very basic English subject/verb/object, with optional adjective/adverb/etc, with explicit terminal wordlists (no morphological type inference); imperatives?
  * SIMP: some simple imperative command language?
  * VPEMDAS: variable assignments; evaluations
  * CSV/DKVP -- including double-quote behaviors
  * Miller DSL

* Where (which `README.md` etc) to note this is all UTF-8
* Have a root Makefile
* What things to move from internal to external (everything?)
