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

.PHONY: sev-build-enclave-image
sev-build-enclave-image: enclave enclave-proxy
	@docker build \
		-f ./sev.Dockerfile \
		-t hello-world-enclave-sev \
		-t us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev \
		--build-arg CONFIG_FILE=sev-config.yaml \
		.

.PHONY: sev-push-enclave-image
sev-push-enclave-image: sev-build-enclave-image
	@docker push us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev

.PHONY: nitro-build-enclave-image
nitro-build-enclave-image: enclave
	@docker build \
		-f ./nitro.Dockerfile \
		-t hello-world-enclave-nitro \
		--build-arg CONFIG_FILE=nitro-config.yaml \
		.

.PHONY: nitro-build-enclave-eif
nitro-build-enclave-eif: nitro-build-enclave-image
	@nitro-cli build-enclave \
		--docker-uri hello-world-enclave-nitro:latest \
		--output-file ./enclave/bin/enclave.eif

.PHONY: nitro-run-enclave-eif
nitro-run-enclave-eif: nitro-build-enclave-eif
	@nitro-cli run-enclave \
		--cpu-count 2 \
		--memory 512 \
		--enclave-cid 4 \
		--eif-path ./enclave/bin/enclave.eif \
		--debug-mode \
		--attach-console

.PHONY: nitro-describe-enclave-eif
nitro-describe-enclave-eif: nitro-build-enclave-eif
	@nitro-cli describe-eif --eif-path ./enclave/bin/enclave.eif

# TODO: nitro-cli describe-enclaves then extract EnclaveID field for use in console cmd
.PHONY: console-enclave-eif
console-enclave-eif:

.PHONY: hello-world-unsafe
hello-world-unsafe: enclave enclave-proxy nonclave
	@-make --no-print-directory servers-up
	@-make --no-print-directory servers-down

# https://github.com/F1bonacc1/process-compose
PROCESS_COMPOSE_PORT=8083
.PHONY: servers-down
servers-down:
	@-process-compose down --port=${PROCESS_COMPOSE_PORT} 2> /dev/null

.PHONY: servers-up
servers-up:
	@-process-compose up --tui=false --port=${PROCESS_COMPOSE_PORT} -f .process-compose.yaml 2> /dev/null

.PHONY: clean
clean:
	rm -rf ./enclave/bin
	rm -rf ./enclave-proxy/bin
	rm -rf ./nonclave/bin
