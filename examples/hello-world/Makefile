# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := hello-world-unsafe
.DELETE_ON_ERROR:
.SUFFIXES:

enclave/bin/enclave: $(shell find ./enclave -type f -name '*.go')
	@cd ./enclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave

enclave-proxy/bin/enclave-proxy: $(shell find ./enclave-proxy -type f -name '*.go')
	@cd ./enclave-proxy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave-proxy

nonclave/bin/nonclave: $(shell find ./nonclave -type f -name '*.go')
	@cd ./nonclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/nonclave

enclave: enclave/bin/enclave
enclave-proxy: enclave-proxy/bin/enclave-proxy
nonclave: nonclave/bin/nonclave

.PHONY: enclave-image
enclave-image-nitro: enclave
	@docker build \
		-f ./enclave/Dockerfile \
		-t hello-world-enclave-nitro \
		--build-arg CONFIG_FILE=nitro-config.yaml \
		./enclave

enclave-image-sev: enclave enclave-proxy
	@docker build \
		-f ./sev.Dockerfile \
		-t hello-world-enclave-sev \
		--build-arg CONFIG_FILE=enclave/sev-config.yaml \
		.

enclave-eif: enclave-image-nitro
	@nitro-cli build-enclave \
		--docker-uri hello-world-enclave-nitro:latest \
		--output-file ./enclave/bin/enclave.eif

.PHONY: run-enclave-eif
run-enclave-eif: enclave-eif
	@nitro-cli run-enclave \
		--cpu-count 2 \
		--memory 512 \
		--enclave-cid 4 \
		--eif-path ./enclave/bin/enclave.eif \
		--debug-mode

.PHONY: describe-enclave-eif
describe-enclave-eif: enclave-eif
	@nitro-cli describe-eif --eif-path ./enclave/bin/enclave.eif

# TODO: nitro-cli describe-enclaves then extract EnclaveID field for use in console cmd
.PHONY: console-enclave-eif
console-enclave-eif:

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
	rm -rf ./enclave-proxy/bin
	rm -rf ./nonclave/bin
