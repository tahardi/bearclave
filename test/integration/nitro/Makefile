# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := build-eif
.DELETE_ON_ERROR:
.SUFFIXES:

bin=./bin
prog=nitro-test
eif=$(prog).eif
dockerfile=./nitro.Dockerfile
image_tag=$(prog)

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
run-eif-debug: build-eif
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
	rm -rf ./bin
	docker rmi nitro-test || true
