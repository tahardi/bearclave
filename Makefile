# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := pre-pr
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: pre-pr
pre-pr: tidy lint test-unit

.PHONY: lint
lint:
	@golangci-lint run --config .golangci.yaml

.PHONY: lint-fix
lint-fix:
	@golangci-lint run --config .golangci.yaml --fix

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: test-unit
test-unit: tidy test-internal-unit

# TODO: Add '-race' back once you've removed BoltDB
.PHONY: test-internal-unit
test-internal-unit:
	@go test -v -count=1 ./internal/...

.PHONY: test-examples
test-examples: hello-world-example

.PHONY: hello-world-example
hello-world-example:
	@make -C ./examples/hello-world/

.PHONY: clean
clean:
	rm -rf ./chains/
