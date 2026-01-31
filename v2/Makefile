# PREFIX=/usr/local
PREFIX=$(HOME)
INSTALLDIR=$(PREFIX)/bin

# ================================================================
# General-use targets

# This must remain the first target in this file, which is what 'make' with no
# arguments will run.
build:
	go build ./cmd/pgpg

check: unit-test regression-test

# DESTDIR is for package installs; nominally blank when this is run interactively.
# See also https://www.gnu.org/prep/standards/html_node/DESTDIR.html
install: build
	mkdir -p $(DESTDIR)/$(INSTALLDIR)
	cp pgpg $(DESTDIR)/$(INSTALLDIR)/

# ================================================================
# Dev targets

# ----------------------------------------------------------------
# Formatting
# go fmt ./... finds experimental C files which we want to ignore.
fmt format:
	go fmt ./...

# ----------------------------------------------------------------
# Static analysis

# Needs first: go install honnef.co/go/tools/cmd/staticcheck@latest
# See also: https://staticcheck.io
staticcheck:
	staticcheck ./...

# ----------------------------------------------------------------
# Keystroke-savers

sure: build check

it: build check
so: install

# ================================================================
# Go does its own dependency management, outside of make.
.PHONY: build check install fmt format staticcheck
