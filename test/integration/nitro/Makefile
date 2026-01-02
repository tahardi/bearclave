# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := run-eif-debug
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: nitro-test
nitro-test: $(shell find . -type f -name '*.go')
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o ./bin/nitro-test

.PHONY: build-docker
build-docker: nitro-test
	@docker build \
		-f ./nitro.Dockerfile \
		-t nitro-test \
		.

.PHONY: build-eif
build-eif: build-docker
	@nitro-cli build-enclave \
		--docker-uri nitro-test:latest \
		--output-file ./bin/nitro-test.eif

.PHONY: run-eif-debug
run-eif-debug: terminate-eif build-eif
	@nitro-cli run-enclave \
		--cpu-count 2 \
		--memory 512 \
		--enclave-cid 4 \
		--eif-path ./bin/nitro-test.eif \
		--debug-mode \
		--attach-console

.PHONY: terminate-eif
terminate-eif:
	@nitro-cli terminate-enclave --all

.PHONY: clean
clean:
	@rm -rf ./bin
	@docker rmi nitro-test 2>/dev/null || true
