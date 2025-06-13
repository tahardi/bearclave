# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := pre-pr
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: pre-pr
pre-pr: tidy mock lint test-unit test-examples

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
test-unit: tidy test-internal-unit

.PHONY: test-internal-unit
test-internal-unit:
	@go test -v -count=1 -race ./internal/...

.PHONY: test-examples
test-examples: \
	hello-world-example \
	hello-http-example \
	hello-http-multi-example

.PHONY: hello-world-example
hello-world-example:
	@make -C ./examples/hello-world/

.PHONY: hello-http-example
hello-http-example:
	@make -C ./examples/hello-http/

.PHONY: hello-http-multi-example
hello-http-multi-example:
	@make -C ./examples/hello-http-multi/

.PHONY: clean
clean:
	@make -C ./examples/hello-world/ clean
	@make -C ./examples/hello-http/ clean
	@make -C ./examples/hello-http-multi/ clean
