# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := run-image
.DELETE_ON_ERROR:
.SUFFIXES:

test_name=tdx-test
gcp_project_id=bearclave
gcp_zone=us-central1-a
gcp_tdx_instance_name=instance-bearclave-tdx
gcp_artifact_registry=us-east1-docker.pkg.dev/$(gcp_project_id)/bearclave
image_tag=$(gcp_artifact_registry)/$(test_name)
image_sha=$(shell gcloud container images list-tags \
	$(image_tag) \
	--sort-by="~TIMESTAMP" \
	--limit=1 \
	--format=json \
	| jq -r '.[0].digest')

.PHONY: tdx-test
tdx-test: $(shell find . -type f -name '*.go')
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o ./bin/$(test_name)

.PHONY: build-image
build-image: tdx-test
	@docker build \
		-f ./tdx.Dockerfile \
		-t $(image_tag) \
		.

.PHONY: push-image
push-image: build-image
	@docker push $(image_tag)

.PHONY: run-image
run-image: push-image
	gcloud compute instances update-container $(gcp_tdx_instance_name) \
		--zone $(gcp_zone) \
		--container-image "$(image_tag)@$(image_sha)" && \
	gcloud compute instances reset $(gcp_tdx_instance_name) --zone $(gcp_zone)

.PHONY: instance-start
instance-start:
	@gcloud compute instances start $(gcp_tdx_instance_name) \
		--zone=$(gcp_zone)

.PHONY: instance-stop
instance-stop:
	@gcloud compute instances stop $(gcp_tdx_instance_name) \
		--zone=$(gcp_zone)

.PHONY: clean
clean:
	@rm -rf ./bin
	@docker rmi $(image_tag) 2>/dev/null || true
