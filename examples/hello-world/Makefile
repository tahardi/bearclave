# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := unsafe-hello-world
.DELETE_ON_ERROR:
.SUFFIXES:

# Unsafe Demo
.PHONY: unsafe-hello-world
unsafe-hello-world:
	@-make --no-print-directory unsafe-start-processes
	@-make --no-print-directory unsafe-stop-processes

# Make sure your process-compose port does not conflict with
# the ports used by enclave, proxy, or nonclave processes
# https://github.com/F1bonacc1/process-compose
PROCESS_COMPOSE_PORT=8083
.PHONY: unsafe-start-processes
unsafe-start-processes:
	@-process-compose up --tui=false --port=${PROCESS_COMPOSE_PORT} -f .process-compose.yaml 2> /dev/null

.PHONY: unsafe-stop-processes
unsafe-stop-processes:
	@-process-compose down --port=${PROCESS_COMPOSE_PORT} 2> /dev/null

# Build Binaries
enclave/bin/enclave: $(shell find ./enclave -type f -name '*.go')
	@cd ./enclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave

enclave-proxy/bin/enclave-proxy: $(shell find ./enclave-proxy -type f -name '*.go')
	@cd ./enclave-proxy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave-proxy

nonclave/bin/nonclave: $(shell find ./nonclave -type f -name '*.go')
	@cd ./nonclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/nonclave

enclave: enclave/bin/enclave
enclave-proxy: enclave-proxy/bin/enclave-proxy
nonclave: nonclave/bin/nonclave

# GCP AMD SEV-SNP
gcp_project_id=bearclave
gcp_zone=us-central1-a
gcp_instance_name=instance-bearclave-sev-snp
gcp_artifact_registry=us-east1-docker.pkg.dev/$(gcp_project_id)/bearclave
sev_enclave_image_name=hello-world-enclave-sev
sev_enclave_image_tag=$(gcp_artifact_registry)/$(sev_enclave_image_name)

.PHONY: sev-build-enclave-image
sev-build-enclave-image: enclave enclave-proxy
	@docker build \
		-f ./sev.Dockerfile \
		-t $(sev_enclave_image_name) \
		-t $(sev_enclave_image_tag) \
		--build-arg CONFIG_FILE=sev-config.yaml \
		.

.PHONY: sev-push-enclave-image
sev-push-enclave-image: sev-build-enclave-image
	@docker push $(sev_enclave_image_tag)

.PHONY: sev-run-enclave-image
sev-run-enclave-image: sev-push-enclave-image
	@echo "Retrieving the most recent image SHA for $(sev_enclave_image_name)..."
	@sha=$(shell gcloud container images list-tags $(sev_enclave_image_tag) --sort-by="~TIMESTAMP" --limit=1 --format='value(digest)') && \
	echo "Most recent image SHA: $$sha" && \
	echo "Updating GCP instance to use the new image with SHA $$sha..." && \
	gcloud compute instances update-container $(gcp_instance_name) \
		--zone $(gcp_zone) \
		--container-image "$(sev_enclave_image_tag)@$$sha" && \
	echo "Restarting GCP instance $(gcp_instance_name)..." && \
	gcloud compute instances reset $(gcp_instance_name) --zone $(gcp_zone)

# AWS Nitro-Enclave
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

# Clean
.PHONY: clean
clean:
	rm -rf ./enclave/bin
	rm -rf ./enclave-proxy/bin
	rm -rf ./nonclave/bin
