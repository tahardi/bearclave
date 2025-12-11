# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := pre-pr
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: pre-pr
pre-pr: tidy mock lint test-unit test-integration

# https://golangci-lint.run/welcome/install/#install-from-sources
# They do not recommend using golangci-lint via go tool directive
# as there are still bugs, but I want to try out go tool and work
# uses an old version of golangci-lint. So, I don't mind guinea
# pigging go tool and using a new version of golangci-lint in here
lint_modfile=modfiles/golangci-lint/go.mod
.PHONY: lint
lint:
	@go tool -modfile=$(lint_modfile) golangci-lint run --config .golangci.yaml

.PHONY: lint-fix
lint-fix:
	@go tool -modfile=$(lint_modfile) golangci-lint run --config .golangci.yaml --fix

mockery_modfile=modfiles/mockery/go.mod
.PHONY: mock
mock: tidy
	@go tool -modfile=$(mockery_modfile) mockery --config=.mockery.yaml

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: test-unit
test-unit: tidy test-unit-tee test-unit-internal

.PHONY: test-unit-tee
test-unit-tee:
	@go test -v -count=1 -race ./tee/...

.PHONY: test-unit-internal
test-unit-internal:
	@go test -v -count=1 -race ./internal/...

.PHONY: test-integration
test-integration:
	@go test -v -count=1 -race ./test/integration/...
