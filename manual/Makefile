# ----------------------------------------------------------------
# This must remain the first target in this file, which is what 'make' with no
# arguments will run.
build:
	go build ./cmd/tryast
	go build ./cmd/trylex
	go build ./cmd/tryparse

check test:
	#go test github.com/johnkerl/pgpg/...
	go test ./...

clean:
	rm -f ./tryast
	rm -f ./trylex
	rm -f ./tryparse

# ----------------------------------------------------------------
# Formatting
# go fmt ./... finds Python files which we want to ignore.
fmt:
	-go fmt ./cmd/...
	-go fmt ./pkg/...

# ----------------------------------------------------------------
# Static analysis

# Needs first: go install honnef.co/go/tools/cmd/staticcheck@latest
# See also: https://staticcheck.io
staticcheck:
	staticcheck ./...

# ----------------------------------------------------------------
# For developers before pushing a commit.
dev:
	-make fmt
	make build
	make check
	@echo DONE

# ================================================================
# Go does its own dependency management, outside of make.
.PHONY: build check fmt staticcheck dev
