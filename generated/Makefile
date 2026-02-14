# ----------------------------------------------------------------
# Grammar specs are "name|TypePrefix" where TypePrefix is used to
# build FooLexer / FooParser type names.
LEX_SPECS=sign-digit|SignDigit pemdas|PEMDAS pemdas_plain|PEMDASPlain statements|Statements seng|SENG lisp|LISP json|JSON json_plain|JSONPlain
PARSE_SPECS=pemdas|PEMDAS pemdas_plain|PEMDASPlain statements|Statements seng|SENG lisp|LISP json|JSON json_plain|JSONPlain

LEXGENS=$(foreach spec,$(LEX_SPECS),go/pkg/lexers/$(firstword $(subst |, ,$(spec))).go)
PARSEGENS=$(foreach spec,$(PARSE_SPECS),go/pkg/parsers/$(firstword $(subst |, ,$(spec))).go)

all: dirs $(LEXGENS) $(PARSEGENS)

dirs:
	@mkdir -p jsons
	@mkdir -p go/pkg/lexers
	@mkdir -p go/pkg/parsers

# ----------------------------------------------------------------
define LEX_GO_RULE
go/pkg/lexers/$(1).go: jsons/$(1)-lex.json
	../generators/go/lexgen-code -o $$@ -type $(2)Lexer $$<
endef

define LEX_JSON_RULE
jsons/$(1)-lex.json: bnfs/$(1).bnf
	../generators/go/lexgen-tables -o $$@ $$<
endef

define PARSE_GO_RULE
go/pkg/parsers/$(1).go: jsons/$(1)-parse.json
	../generators/go/parsegen-code -o $$@ -package parsers -type $(2)Parser $$<
endef

define PARSE_JSON_RULE
jsons/$(1)-parse.json: bnfs/$(1).bnf
	../generators/go/parsegen-tables -o $$@ $$<
endef

$(foreach spec,$(LEX_SPECS),$(eval $(call LEX_GO_RULE,$(firstword $(subst |, ,$(spec))),$(word 2,$(subst |, ,$(spec))))))
$(foreach spec,$(LEX_SPECS),$(eval $(call LEX_JSON_RULE,$(firstword $(subst |, ,$(spec))))))
$(foreach spec,$(PARSE_SPECS),$(eval $(call PARSE_GO_RULE,$(firstword $(subst |, ,$(spec))),$(word 2,$(subst |, ,$(spec))))))
$(foreach spec,$(PARSE_SPECS),$(eval $(call PARSE_JSON_RULE,$(firstword $(subst |, ,$(spec))))))


# ----------------------------------------------------------------
# Formatting
# go fmt ./... finds Python files which we want to ignore.
fmt:
	-go fmt ./go/pkg/...

# ----------------------------------------------------------------
# Needs first: go install honnef.co/go/tools/cmd/staticcheck@latest
# See also: https://staticcheck.io
staticcheck:
	staticcheck ./...

# ================================================================
# Go does its own dependency management, outside of make.
.PHONY: build check fmt staticcheck dev
