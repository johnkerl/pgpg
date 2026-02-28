# Top

* Multi-object:
  * Make sure truly streaming

* PASCAL-S

* UX findings from PASCAL-S:
  * Have more parsing-debug tools available in sample apps
  * Write up: Sharp edge if this isn't first b/c first-found & it matches identifier
  * Write up: `./generate.sh  && go build && ./astprint -e 'program foo'`
  * Write up: Better error messages with missing semicolons
  * Write up: Root must come first

* Error out on this:
```
$ rlwrap ./apps/go/pemdas-eval -mode f2poly -mod-poly 1f -l
```

* Perf-analyze the JSON lexer

# Language ideas

## Programming languages

* PEMDAS & VIC obv
* Miller/millerish ...
* LISP/ish
* PASCAL-S and/or PL/0
  * BNFs
  * Manual recursive descent

## Data languages

* CSV
* DKVP
* DKVPX
* JSON obv
* Try to get streaming reads over multiple objects

## Utility languages

* Config-file language for Miller -- ? Presumably there are standard config-file packages already out there.
* Something shell-like -- ?

# Tools to-do

* CI:
  * Copilot reviews
* UX:
  * No D on ^D at CLI
* Efficiency (memory and/or CPU):
  * Scale-test everything for perf early on -- especially channel-switching
  * DFA minimization
* Docs:
  * Doc `tryparse` with `-tokens`, `-states`, `-stack`.
  * ??  `GOPROXY=direct go mod tidy`
* Code-neatens:
  * Move types to tops of files
  * Dedupe
  * Re-organize and grok the lexer/parser tables etc.
  * Prohibit any parent/child indices the same
  * Allow separate `.bnf` files for lex and parse? In case of all parser bits identical (e.g. `PEMDAS*`).
  * AST to-string factored out of printer
    * Use for UT
  * [Maybe] Wrap lexers in LA1 and LA2 etc for lookahead level:
    * Hide direct calls to Scan``
    * `.First()`
    * `.Second()`
    * `.Advance()`
* Implmementation languages:
  * Catch more things up to Python
  * Catch more things up to JS
* Misc:
  * Hand-written recursive-ascent impls for simple grammars?

# Credits

* The Dragon book
* https://github.com/goccmack/gocc
* https://go.dev/talks/2011/lex.slide#1
* Cursor
