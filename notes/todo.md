# Top

* Multi-object:
  * Move to io.Reader (w/ string-handling option of course)
  * Then implement streaming proposal

* PASCAL-S
  * have more parsing-debug tools available in sample apps!
  * wup: sharp edge if this isn't first b/c first-found & it matches identifier
  * wup: ./generate.sh  && go build && ./astprint -e 'program foo'
  * wup: better with missing semicolons
  * wup: Root must come first
  * top-level and next-level makefiles for pgpge: build fmt clean [test]
    * `go get github.com/johnkerl/pgpg/lib@v0.2.0`
    * `go get github.com/johnkerl/pgpg/generators/go@v0.2.0`

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

* Dependabot
* Copilot reviews
* Prohibit any parent/child indices the same
* Port manuals to JavaScript
* Re-organize and grok the lexer/parser tables etc.
* D on ^D at CLI
* Allow separate `.bnf` files for lex and parse? In case of all parser bits identical (e.g. `PEMDAS*`).
* `go.work` narrative ..
* `taginfo.sh` ... make this crisper ...
* ??  `GOPROXY=direct go mod tidy`
* Code neatens:
  * Move types to tops of files
  * Dedupe
  * Etc.
* Doc `tryparse` with `-tokens`, `-states`, `-stack`.
* Scale-test everything for perf early on -- especially channel-switching
* DFA minimization
* Hand-written recursive-ascent impls for simple grammars?
* AST to-string factored out of printer
  * Use for UT
* Wrap lexers in LA1 and LA2 etc for lookahead level:
  * Hide direct calls to Scan``
  * `.First()`
  * `.Second()`
  * `.Advance()`

# Credits

* The Dragon book
* https://github.com/goccmack/gocc
* https://go.dev/talks/2011/lex.slide#1
* Cursor

# FF arithmetic REPL etc.

To port:

* PYFFL: port eval DSL/CLI from GOFFL
* PYFFL: port cmds from RUFFL

What to implement in a BNF/DSL vs. using Python for scripting:

* Desk calc
* Math tables by datatype & op
* Multiplicative order
* GCD, LCM, totient
* Factorization
* Irreducibility etc.

Borderline:

* Matrix I/O ...

RUFFL mains:

```
Arithmetic (EDs or residues):
-----------------------------
f-  f.  f+  fdiv  fexp  fmod
z-  z.  z+  zdiv  zexp  zmod
fm- fm. fm+ fmdiv fmexp
zm- zm. zm+ zmdiv zmexp
```

```
Euclidean domains:
------------------
fgcd     zgcd
fegcd    zegcd
flcm     zlcm
ftotient ztotient
ffactor  zfactor
frandom zrandom

fdeg
fdivisors

flowestirr
frandomirr
ftestirr
ftestprim

zdivisors
```

```
Residue rings/fields:
---------------------

fmlist   zmlist
fmtbl    zmtbl
fmrandom zmrandom
fmord    zmord
fmmaxord zmmaxord
fperiod
fmorbit  zmorbit

zmfindgen
```
