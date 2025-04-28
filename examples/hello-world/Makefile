# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := hello-world-unsafe
.DELETE_ON_ERROR:
.SUFFIXES:

enclave: enclave/bin/enclave

enclave/bin/enclave: $(shell find ./enclave -type f -name '*.go')
	@cd ./enclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave

nonclave: nonclave/bin/nonclave

nonclave/bin/nonclave: $(shell find ./nonclave -type f -name '*.go')
	@cd ./nonclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/nonclave

.PHONY: enclave-image
enclave-image: enclave
	@docker build -f ./enclave/Dockerfile -t hello-world-enclave ./enclave

.PHONY: hello-world-unsafe
hello-world-unsafe:
	@make --no-print-directory servers-up

# https://github.com/F1bonacc1/process-compose
PROCESS_COMPOSE_PORT=8082
.PHONY: servers-down
servers-down:
	@process-compose down --port=${PROCESS_COMPOSE_PORT}

.PHONY: servers-up
servers-up:
	@process-compose up --tui=false --port=${PROCESS_COMPOSE_PORT} -f .process-compose.yaml

.PHONY: clean
clean:
	rm -rf ./enclave/bin
	rm -rf ./nonclave/bin
