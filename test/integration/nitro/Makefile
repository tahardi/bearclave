# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := build
.DELETE_ON_ERROR:
.SUFFIXES:

NITRO_CLI ?= nitro-cli

bin/nitro-test: $(shell find . -type f -name '*.go')
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o ./bin/nitro-test

.PHONY: build-docker
build-docker: bin/nitro-test
	@docker build \
		-f ./nitro.Dockerfile \
		-t nitro-test \
		.

.PHONY: build
build: build-docker
	@$(NITRO_CLI) build-enclave \
		--docker-uri nitro-test:latest \
		--output-file ./bin/nitro-test.eif

.PHONY: run
run: terminate build
	@nitro-cli run-enclave \
		--cpu-count 2 \
		--memory 512 \
		--enclave-cid 4 \
		--eif-path ./bin/nitro-test.eif \
		--debug-mode \
		--attach-console

.PHONY: terminate
terminate:
	@nitro-cli terminate-enclave --all

.PHONY: clean
clean:
	@rm -rf ./bin
	@docker rmi nitro-test 2>/dev/null || true
