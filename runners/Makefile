# ----------------------------------------------------------------
# This must remain the first target in this file, which is what 'make' with no
# arguments will run.
build:
	@set -e; \
	for d in ./cmd/*; do \
		go build "$$d"; \
	done

check test:
	go test ./...

clean:
	@set -e; \
	for d in ./cmd/*; do \
		rm -f "./$$(basename "$$d")"; \
	done

# ----------------------------------------------------------------
# Formatting
# go fmt ./... finds Python files which we want to ignore.
fmt:
	-go fmt ./cmd/...

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
