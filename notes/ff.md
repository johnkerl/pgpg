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
