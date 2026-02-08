# Test grammars

## AME

This is add and multiply of integers, with equal precedence.

GOCC grammar: [https://github.com/johnkerl/pgpg/blob/main/grammar-check/bnfs/ame.bnf](https://github.com/johnkerl/pgpg/blob/main/grammar-check/bnfs/ame.bnf)

## AMNE

This is add and multiply of integers, with unequal precedence.

GOCC grammar: [https://github.com/johnkerl/pgpg/blob/main/grammar-check/bnfs/amne.bnf](https://github.com/johnkerl/pgpg/blob/main/grammar-check/bnfs/amne.bnf)

## PEMDAS

Arithmetic with PEMDAS rules.

## VIC

Variables-and-integers calculator.

GOCC grammar: [https://github.com/johnkerl/pgpg/blob/main/grammar-check/bnfs/vic.bnf](https://github.com/johnkerl/pgpg/blob/main/grammar-check/bnfs/vic.bnf)

## VBC

Variables-and-booleans calculator.

Variable names, `AND`, `OR`, `NOT`, parentheses, assignments, evaluations.

## SENG

Simple English statements:

* subject/object/verb or imperative object/verb
* articles, adjectives, adverbs
* explicit terminal wordlists -- no morphological type-inference e.g. ends with `-ly` meaning it must be an adverb

GOCC grammar: [https://github.com/johnkerl/pgpg/blob/main/grammar-check/bnfs/seng.bnf](https://github.com/johnkerl/pgpg/blob/main/grammar-check/bnfs/seng.bnf)

Examples:

```
dog eats food
the dog eats the food
the brown dog tastes the very new food
take the old book
# _maybe_ conjunctions at higher and lower levels: _
take the old book and read it.
take the green and grey book.
```

## CSV/DKVP

[Comma-separated-value](https://miller.readthedocs.io/en/latest/file-formats/#csvtsvasvusvetc) and [DKVP](https://miller.readthedocs.io/en/latest/file-formats/#dkvp-key-value-pairs) files (Miller).

## EBNF Notes

This repo's EBNF dialect supports character ranges for lexer rules, using single-rune string literals. Example:

```
_lower ::= "a"-"z";
_upper ::= "A"-"Z";
_digit ::= "0"-"9";
id ::= ("_" | _lower | _upper) { "_" | _lower | _upper | _digit };
```

## Miller DSL

This is an ultimate goal. GOCC grammar: [https://github.com/johnkerl/miller/blob/main/internal/pkg/parsing/mlr.bnf](https://github.com/johnkerl/miller/blob/main/internal/pkg/parsing/mlr.bnf).
