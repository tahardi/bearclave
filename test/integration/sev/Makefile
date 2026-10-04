# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := build
.DELETE_ON_ERROR:
.SUFFIXES:

GCP_PROJECT_ID ?= bearclave
GCP_REGISTRY ?= us-east1-docker.pkg.dev/$(GCP_PROJECT_ID)/bearclave
IMAGE_TAG ?= latest
image=$(GCP_REGISTRY)/sev-test:$(IMAGE_TAG)

bin/sev-test: $(shell find . -type f -name '*.go')
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o ./bin/sev-test

.PHONY: build
build: bin/sev-test
	@docker build \
		-f ./sev.Dockerfile \
		-t $(image) \
		.

.PHONY: push
push: build
	@docker push $(image)

.PHONY: image-ref
image-ref:
	@echo $(image)

.PHONY: clean
clean:
	@rm -rf ./bin
	@docker rmi $(image) 2>/dev/null || true
