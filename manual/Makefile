# ----------------------------------------------------------------
check test:
	#go test github.com/johnkerl/pgpg/manual/...
	go test ./...

# ----------------------------------------------------------------
# Formatting
# go fmt ./... finds Python files which we want to ignore.
fmt:
	-go fmt ./go/pkg/...

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
	make check
	@echo DONE

# ================================================================
# Go does its own dependency management, outside of make.
.PHONY: build check fmt staticcheck dev
