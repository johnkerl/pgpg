# ----------------------------------------------------------------
# Build lib and generators; install generator binaries to bin/ for
# use by apps/go/generated (e.g. ../../../../go/bin/lexgen-tables).
# ----------------------------------------------------------------
BIN := bin

build:
	go build ./...
	@mkdir -p $(BIN)
	go build -o $(BIN)/lexgen-tables ./generators/cmd/lexgen-tables
	go build -o $(BIN)/lexgen-code   ./generators/cmd/lexgen-code
	go build -o $(BIN)/parsegen-tables ./generators/cmd/parsegen-tables
	go build -o $(BIN)/parsegen-code   ./generators/cmd/parsegen-code

test:
	go test ./...

clean:
	rm -rf $(BIN)

fmt:
	go fmt ./...

# Static analysis (requires: go install honnef.co/go/tools/cmd/staticcheck@latest)
staticcheck:
	staticcheck ./...

# Pre-push: fmt, build, test
dev: fmt build test
	@echo DONE

.PHONY: build test clean fmt staticcheck dev
