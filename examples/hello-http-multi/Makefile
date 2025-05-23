# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := notee-hello-http
.DELETE_ON_ERROR:
.SUFFIXES:

################################################################################
# No TEE Demo
################################################################################
.PHONY: notee-hello-http
notee-hello-http:
	@make --no-print-directory notee-start-processes

# Make sure your process-compose port does not conflict with
# the ports used by enclave, proxy, or nonclave processes
# https://github.com/F1bonacc1/process-compose
PROCESS_COMPOSE_PORT=8081
.PHONY: notee-start-processes
notee-start-processes:
	@process-compose up --tui=false --port=${PROCESS_COMPOSE_PORT} -f .process-compose.yaml 2> /dev/null

.PHONY: notee-stop-processes
notee-stop-processes:
	@process-compose down --port=${PROCESS_COMPOSE_PORT} 2> /dev/null
	@kill -9 $$(lsof -t -i :${PROCESS_COMPOSE_PORT}) || true

################################################################################
# Build Binaries
################################################################################
enclave-server-1/bin/enclave-server-1: $(shell find ./enclave-server-1 -type f -name '*.go')
	@cd ./enclave-server-1 && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave-server-1

enclave-server-2/bin/enclave-server-2: $(shell find ./enclave-server-2 -type f -name '*.go')
	@cd ./enclave-server-2 && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave-server-2

enclave-proxy/bin/enclave-proxy: $(shell find ./enclave-proxy -type f -name '*.go')
	@cd ./enclave-proxy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave-proxy

nonclave/bin/nonclave: $(shell find ./nonclave -type f -name '*.go')
	@cd ./nonclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/nonclave

enclave-server-1: enclave-server-1/bin/enclave-server-1
enclave-server-2: enclave-server-2/bin/enclave-server-2
enclave-proxy: enclave-proxy/bin/enclave-proxy
nonclave: nonclave/bin/nonclave

################################################################################
# AWS Nitro-Enclave
################################################################################
nitro_aws_profile=personal
nitro_aws_region=us-east-2
nitro_aws_instance_name=tahardi-bearclave

.PHONY: nitro-build-enclave-image
nitro-build-enclave-image: enclave-server-1 enclave-server-2
	@docker build \
		-f ./nitro.Dockerfile \
		-t hello-http-multi-enclave-nitro \
		--build-arg CONFIG_FILE=configs/nitro-config.yaml \
		.

.PHONY: nitro-build-enclave-eif
nitro-build-enclave-eif: nitro-build-enclave-image
	@nitro-cli build-enclave \
		--docker-uri hello-http-multi-enclave-nitro:latest \
		--output-file ./enclaves.eif

.PHONY: nitro-run-enclave-eif
nitro-run-enclave-eif: nitro-build-enclave-eif
	@nitro-cli run-enclave \
		--cpu-count 2 \
		--memory 512 \
		--enclave-cid 4 \
		--eif-path ./enclaves.eif \
		--debug-mode \
		--attach-console

.PHONY: nitro-describe-enclave-eif
nitro-describe-enclave-eif: nitro-build-enclave-eif
	@nitro-cli describe-eif --eif-path ./enclaves.eif

.PHONY: nitro-start-instance
nitro-start-instance:
	@instance_id=$$(aws ec2 describe-instances \
		--profile $(nitro_aws_profile) \
		--no-cli-pager \
		--filters "Name=tag:Name,Values=$(nitro_aws_instance_name)" \
		--query "Reservations[].Instances[].InstanceId" \
		--output text) && \
	aws ec2 start-instances \
		--profile $(nitro_aws_profile) \
		--no-cli-pager \
		--region=$(nitro_aws_region) \
		--instance-ids $$instance_id

.PHONY: nitro-stop-instance
nitro-stop-instance:
	@instance_id=$$(aws ec2 describe-instances \
		--profile $(nitro_aws_profile) \
		--no-cli-pager \
		--filters "Name=tag:Name,Values=$(nitro_aws_instance_name)" \
		--query "Reservations[].Instances[].InstanceId" \
		--output text) && \
	aws ec2 stop-instances \
		--profile $(nitro_aws_profile) \
		--no-cli-pager \
		--region=$(nitro_aws_region) \
		--instance-ids $$instance_id

.PHONY: nitro-ssh-instance
nitro-ssh-instance:
	@public_dns=$$(aws ec2 describe-instances \
		--profile $(nitro_aws_profile) \
		--no-cli-pager \
		--filters "Name=tag:Name,Values=$(nitro_aws_instance_name)" \
		--query "Reservations[].Instances[].PublicDnsName" \
		--output text); \
	if [ -z "$$public_dns" ]; then \
		echo "Error: No public DNS name found for instance $(nitro_aws_instance_name)"; \
		exit 1; \
	fi; \
	echo "Updating ~/.ssh/config with Public DNS: $$public_dns"; \
	sed -i.bak "/^Host ec2-bearclave$$/,/^Host /{s|^\([[:space:]]*Hostname[[:space:]]\).*|\1$$public_dns|}" ~/.ssh/config; \
	ssh ec2-bearclave

################################################################################
# GCP AMD SEV-SNP
################################################################################
sev_gcp_project_id=bearclave
sev_gcp_zone=us-central1-a
sev_gcp_instance_name=instance-bearclave-sev-snp
sev_gcp_artifact_registry=us-east1-docker.pkg.dev/$(sev_gcp_project_id)/bearclave
sev_enclave_image_name=hello-http-multi-enclave-sev
sev_enclave_image_tag=$(sev_gcp_artifact_registry)/$(sev_enclave_image_name)

.PHONY: sev-build-enclave-image
sev-build-enclave-image: enclave-server-1 enclave-server-2 enclave-proxy
	@docker build \
		-f ./sev.Dockerfile \
		-t $(sev_enclave_image_name) \
		-t $(sev_enclave_image_tag) \
		--build-arg CONFIG_FILE=configs/sev-config.yaml \
		.

.PHONY: sev-push-enclave-image
sev-push-enclave-image: sev-build-enclave-image
	@docker push $(sev_enclave_image_tag)

.PHONY: sev-run-enclave-image
sev-run-enclave-image: sev-push-enclave-image
	@echo "Retrieving the most recent image SHA for $(sev_enclave_image_name)..."
	@sha=$(shell gcloud container images list-tags $(sev_enclave_image_tag) --sort-by="~TIMESTAMP" --limit=1 --format=json | jq -r '.[0].digest') && \
	echo "Most recent image SHA: $$sha" && \
	echo "Updating GCP instance to use the new image with SHA $$sha..." && \
	gcloud compute instances update-container $(sev_gcp_instance_name) \
		--zone $(sev_gcp_zone) \
		--container-image "$(sev_enclave_image_tag)@$$sha" && \
	echo "Restarting GCP instance $(sev_gcp_instance_name)..." && \
	gcloud compute instances reset $(sev_gcp_instance_name) --zone $(sev_gcp_zone)

.PHONY: sev-run-nonclave
sev-run-nonclave:
	@ip=$$(gcloud compute instances describe $(sev_gcp_instance_name) \
		--zone $(sev_gcp_zone) \
		--format='get(networkInterfaces[0].accessConfigs[0].natIP)') && \
	go run ./nonclave/main.go --config ./configs/sev-config.yaml --host $$ip

.PHONY: sev-start-instance
sev-start-instance:
	@gcloud compute instances start $(sev_gcp_instance_name) \
		--zone=$(sev_gcp_zone)

.PHONY: sev-stop-instance
sev-stop-instance:
	@gcloud compute instances stop $(sev_gcp_instance_name) \
		--zone=$(sev_gcp_zone)

.PHONY: sev-ssh-instance
sev-ssh-instance:
	@gcloud compute ssh $(sev_gcp_instance_name)

################################################################################
# GCP Intel TDX
################################################################################
tdx_gcp_project_id=bearclave
tdx_gcp_zone=us-central1-a
tdx_gcp_instance_name=instance-bearclave-tdx
tdx_gcp_artifact_registry=us-east1-docker.pkg.dev/$(tdx_gcp_project_id)/bearclave
tdx_enclave_image_name=hello-http-multi-enclave-tdx
tdx_enclave_image_tag=$(tdx_gcp_artifact_registry)/$(tdx_enclave_image_name)

.PHONY: tdx-build-enclave-image
tdx-build-enclave-image: enclave-server-1 enclave-server-2 enclave-proxy
	@docker build \
		-f ./tdx.Dockerfile \
		-t $(tdx_enclave_image_name) \
		-t $(tdx_enclave_image_tag) \
		--build-arg CONFIG_FILE=configs/tdx-config.yaml \
		.

.PHONY: tdx-push-enclave-image
tdx-push-enclave-image: tdx-build-enclave-image
	@docker push $(tdx_enclave_image_tag)

.PHONY: tdx-run-enclave-image
tdx-run-enclave-image: tdx-push-enclave-image
	@echo "Retrieving the most recent image SHA for $(tdx_enclave_image_name)..."
	@sha=$(shell gcloud container images list-tags $(tdx_enclave_image_tag) --sort-by="~TIMESTAMP" --limit=1 --format=json | jq -r '.[0].digest') && \
	echo "Most recent image SHA: $$sha" && \
	echo "Updating GCP instance to use the new image with SHA $$sha..." && \
	gcloud compute instances update-container $(tdx_gcp_instance_name) \
		--zone $(tdx_gcp_zone) \
		--container-image "$(tdx_enclave_image_tag)@$$sha" && \
	echo "Restarting GCP instance $(tdx_gcp_instance_name)..." && \
	gcloud compute instances reset $(tdx_gcp_instance_name) --zone $(tdx_gcp_zone)

.PHONY: tdx-run-nonclave
tdx-run-nonclave:
	@ip=$$(gcloud compute instances describe $(tdx_gcp_instance_name) \
		--zone $(tdx_gcp_zone) \
		--format='get(networkInterfaces[0].accessConfigs[0].natIP)') && \
	go run ./nonclave/main.go --config ./configs/sev-config.yaml --host $$ip

.PHONY: tdx-start-instance
tdx-start-instance:
	@gcloud compute instances start $(tdx_gcp_instance_name) \
		--zone=$(tdx_gcp_zone)

.PHONY: tdx-stop-instance
tdx-stop-instance:
	@gcloud compute instances stop $(tdx_gcp_instance_name) \
		--zone=$(tdx_gcp_zone)

.PHONY: tdx-ssh-instance
tdx-ssh-instance:
	@gcloud compute ssh $(tdx_gcp_instance_name)

################################################################################
# Clean
################################################################################
.PHONY: clean
clean:
	rm -rf ./enclave-server-1/bin
	rm -rf ./enclave-server-2/bin
	rm -rf ./enclave-proxy/bin
	rm -rf ./nonclave/bin