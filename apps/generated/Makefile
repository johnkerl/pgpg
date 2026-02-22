# ----------------------------------------------------------------
# Grammar specs are "name|TypePrefix" where TypePrefix is used to
# build FooLexer / FooParser type names.
LEX_SPECS=\
  sign-digit|SignDigit \
  pemdas|PEMDAS \
  pemdas_int|PEMDASInt \
  pemdas_float|PEMDASFloat \
  pemdas_mod|PEMDASMod \
  pemdas_plain|PEMDASPlain \
  statements|Statements \
  seng|SENG \
  lisp|LISP \
  json|JSON \
  json_plain|JSONPlain

PARSE_SPECS=\
  pemdas|PEMDAS \
  pemdas_int|PEMDASInt \
  pemdas_float|PEMDASFloat \
  pemdas_mod|PEMDASMod \
  pemdas_plain|PEMDASPlain \
  statements|Statements \
  seng|SENG \
  lisp|LISP \
  json|JSON \
  json_plain|JSONPlain

GO_GEN := ../go/generated
JSONS := ../jsons

LEXGENS=$(foreach spec,$(LEX_SPECS),$(GO_GEN)/pkg/lexers/$(firstword $(subst |, ,$(spec))).go)
PARSEGENS=$(foreach spec,$(PARSE_SPECS),$(GO_GEN)/pkg/parsers/$(firstword $(subst |, ,$(spec))).go)

all: dirs $(LEXGENS) $(PARSEGENS)

dirs:
	@mkdir -p $(JSONS)
	@mkdir -p $(GO_GEN)/pkg/lexers
	@mkdir -p $(GO_GEN)/pkg/parsers

# ----------------------------------------------------------------
define LEX_GO_RULE
$(GO_GEN)/pkg/lexers/$(1).go: $(JSONS)/$(1)-lex.json
	../../generators/go/lexgen-code -o $$@ -type $(2)Lexer $$<
endef

define LEX_JSON_RULE
$(JSONS)/$(1)-lex.json: ../bnfs/$(1).bnf
	../../generators/go/lexgen-tables -o $$@ $$<
endef

define PARSE_GO_RULE
$(GO_GEN)/pkg/parsers/$(1).go: $(JSONS)/$(1)-parse.json
	../../generators/go/parsegen-code -o $$@ -package parsers -type $(2)Parser $$<
endef

define PARSE_JSON_RULE
$(JSONS)/$(1)-parse.json: ../bnfs/$(1).bnf
	../../generators/go/parsegen-tables -o $$@ $$<
endef

$(foreach spec,$(LEX_SPECS),$(eval $(call LEX_GO_RULE,$(firstword $(subst |, ,$(spec))),$(word 2,$(subst |, ,$(spec))))))
$(foreach spec,$(LEX_SPECS),$(eval $(call LEX_JSON_RULE,$(firstword $(subst |, ,$(spec))))))
$(foreach spec,$(PARSE_SPECS),$(eval $(call PARSE_GO_RULE,$(firstword $(subst |, ,$(spec))),$(word 2,$(subst |, ,$(spec))))))
$(foreach spec,$(PARSE_SPECS),$(eval $(call PARSE_JSON_RULE,$(firstword $(subst |, ,$(spec))))))


# ----------------------------------------------------------------
clean:

genclean:
	rm -rf $(JSONS)
	rm -rf $(GO_GEN)/pkg/lexers
	rm -rf $(GO_GEN)/pkg/parsers

# ----------------------------------------------------------------
# Formatting
# go fmt ./... finds Python files which we want to ignore.
fmt:
	cd $(GO_GEN) && go fmt ./pkg/...

# ----------------------------------------------------------------
# Needs first: go install honnef.co/go/tools/cmd/staticcheck@latest
# See also: https://staticcheck.io
staticcheck:
	cd $(GO_GEN) && staticcheck ./...

# ================================================================
# Go does its own dependency management, outside of make.
.PHONY: build check fmt staticcheck dev clean
