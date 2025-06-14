# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := notee-hello-world
.DELETE_ON_ERROR:
.SUFFIXES:

################################################################################
# No TEE Demo
################################################################################
.PHONY: notee-hello-world
notee-hello-world:
	@make --no-print-directory notee-processes-start

# Make sure your process-compose port does not conflict with
# the ports used by enclave, proxy, or nonclave processes
# https://github.com/F1bonacc1/process-compose
PROCESS_COMPOSE_PORT=8081
.PHONY: notee-processes-start
notee-processes-start:
	@process-compose up --tui=false --port=${PROCESS_COMPOSE_PORT} -f .process-compose.yaml 2> /dev/null

.PHONY: notee-processes-stop
notee-processes-stop:
	@process-compose down --port=${PROCESS_COMPOSE_PORT} 2> /dev/null
	@kill -9 $$(lsof -t -i :${PROCESS_COMPOSE_PORT}) || true

################################################################################
# Build Binaries
################################################################################
enclave/bin/enclave: $(shell find ./enclave -type f -name '*.go')
	@cd ./enclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave

enclave-proxy/bin/enclave-proxy: $(shell find ./enclave-proxy -type f -name '*.go')
	@cd ./enclave-proxy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/enclave-proxy

nonclave/bin/nonclave: $(shell find ./nonclave -type f -name '*.go')
	@cd ./nonclave && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/nonclave

enclave: enclave/bin/enclave
enclave-proxy: enclave-proxy/bin/enclave-proxy
nonclave: nonclave/bin/nonclave

################################################################################
# AWS
################################################################################
aws_profile=personal
aws_region=us-east-2

.PHONY: aws-cli-login
aws-cli-login:
	aws sso login --profile $(aws_profile)

# AWS Nitro
################################################################################
aws_nitro_instance_name=tahardi-bearclave
aws_nitro_instance_id=$(shell aws ec2 describe-instances \
	--profile $(aws_profile) \
	--no-cli-pager \
	--filters "Name=tag:Name,Values=$(aws_nitro_instance_name)" \
	--query "Reservations[].Instances[].InstanceId" \
	--output text)
aws_nitro_instance_ip=$(shell aws ec2 describe-instances \
	--profile $(aws_profile) \
	--no-cli-pager \
	--filters "Name=tag:Name,Values=$(aws_nitro_instance_name)" \
	--query "Reservations[].Instances[].PublicDnsName" \
	--output text)
aws_nitro_enclave_config=configs/nitro-config.yaml
aws_nitro_enclave_image_tag=hello-world-enclave-nitro
aws_nitro_enclave_eif_path=./enclave/bin/enclave.eif
aws_nitro_enclave_cpu_count=2
aws_nitro_enclave_memory_size=512
aws_nitro_enclave_cid=4

.PHONY: aws-nitro-enclave-build-image
aws-nitro-enclave-build-image: enclave
	@docker build \
		-f ./nitro.Dockerfile \
		-t $(aws_nitro_enclave_image_tag) \
		--build-arg CONFIG_FILE=$(aws_nitro_enclave_config) \
		.

.PHONY: aws-nitro-enclave-build-eif
aws-nitro-enclave-build-eif: aws-nitro-enclave-build-image
	@nitro-cli build-enclave \
		--docker-uri $(aws_nitro_enclave_image_tag):latest \
		--output-file $(aws_nitro_enclave_eif_path)

.PHONY: aws-nitro-enclave-run-eif
aws-nitro-enclave-run-eif: aws-nitro-enclave-build-eif
	@nitro-cli run-enclave \
		--cpu-count $(aws_nitro_enclave_cpu_count) \
		--memory $(aws_nitro_enclave_memory_size) \
		--enclave-cid $(aws_nitro_enclave_cid) \
		--eif-path $(aws_nitro_enclave_eif_path)

.PHONY: aws-nitro-enclave-run-eif-debug
aws-nitro-enclave-run-eif-debug: aws-nitro-enclave-build-eif
	@nitro-cli run-enclave \
		--cpu-count $(aws_nitro_enclave_cpu_count) \
		--memory $(aws_nitro_enclave_memory_size) \
		--enclave-cid $(aws_nitro_enclave_cid) \
		--eif-path $(aws_nitro_enclave_eif_path) \
		--debug-mode \
		--attach-console

.PHONY: aws-nitro-enclave-describe-eif
aws-nitro-enclave-describe-eif: aws-nitro-enclave-build-eif
	@nitro-cli describe-eif --eif-path $(aws_nitro_enclave_eif_path)

.PHONY: aws-nitro-enclave-terminate-eifs
aws-nitro-enclave-terminate-eifs:
	@nitro-cli terminate-enclave --all

.PHONY: aws-nitro-proxy-run
aws-nitro-proxy-run:
	go run ./proxy/main.go --config $(aws_nitro_enclave_config)

.PHONY: aws-nitro-nonclave-run
aws-nitro-nonclave-run:
	go run ./nonclave/main.go \
		--config $(aws_nitro_enclave_config) \
		--host $(aws_nitro_instance_ip)

.PHONY: aws-nitro-instance-start
aws-nitro-instance-start:
	@aws ec2 start-instances \
		--profile $(aws_profile) \
		--no-cli-pager \
		--region=$(aws_region) \
		--instance-ids $(aws_nitro_instance_id)

.PHONY: aws-nitro-instance-stop
aws-nitro-instance-stop:
	@aws ec2 stop-instances \
		--profile $(aws_profile) \
		--no-cli-pager \
		--region=$(aws_region) \
		--instance-ids $(aws_nitro_instance_id)

.PHONY: aws-nitro-instance-ssh
aws-nitro-instance-ssh:
	if [ -z "$(aws_nitro_instance_ip)" ]; then \
		echo "Error: No public DNS name found for instance $(aws_nitro_instance_name)"; \
		exit 1; \
	fi; \
	echo "Updating ~/.ssh/config with Public DNS: $(aws_nitro_instance_ip)"; \
	sed -i.bak "/^Host ec2-bearclave$$/,/^Host /{s|^\([[:space:]]*Hostname[[:space:]]\).*|\1$(aws_nitro_instance_ip)|}" ~/.ssh/config; \
	ssh ec2-bearclave

################################################################################
# GCP AMD SEV-SNP
################################################################################
sev_gcp_project_id=bearclave
sev_gcp_zone=us-central1-a
sev_gcp_instance_name=instance-bearclave-sev-snp
sev_gcp_artifact_registry=us-east1-docker.pkg.dev/$(sev_gcp_project_id)/bearclave
sev_enclave_image_name=hello-world-enclave-sev
sev_enclave_image_tag=$(sev_gcp_artifact_registry)/$(sev_enclave_image_name)

.PHONY: sev-enclave-build-image
sev-enclave-build-image: enclave enclave-proxy
	@docker build \
		-f ./sev.Dockerfile \
		-t $(sev_enclave_image_name) \
		-t $(sev_enclave_image_tag) \
		--build-arg CONFIG_FILE=configs/sev-config.yaml \
		.

.PHONY: sev-enclave-push-image
sev-enclave-push-image: sev-enclave-build-image
	@docker push $(sev_enclave_image_tag)

.PHONY: sev-enclave-run-image
sev-enclave-run-image: sev-enclave-push-image
	@echo "Retrieving the most recent image SHA for $(sev_enclave_image_name)..."
	@sha=$(shell gcloud container images list-tags $(sev_enclave_image_tag) --sort-by="~TIMESTAMP" --limit=1 --format=json | jq -r '.[0].digest') && \
	echo "Most recent image SHA: $$sha" && \
	echo "Updating GCP instance to use the new image with SHA $$sha..." && \
	gcloud compute instances update-container $(sev_gcp_instance_name) \
		--zone $(sev_gcp_zone) \
		--container-image "$(sev_enclave_image_tag)@$$sha" && \
	echo "Restarting GCP instance $(sev_gcp_instance_name)..." && \
	gcloud compute instances reset $(sev_gcp_instance_name) --zone $(sev_gcp_zone)

.PHONY: sev-nonclave-run
sev-nonclave-run:
	@ip=$$(gcloud compute instances describe $(sev_gcp_instance_name) \
		--zone $(sev_gcp_zone) \
		--format='get(networkInterfaces[0].accessConfigs[0].natIP)') && \
	go run ./nonclave/main.go --platform sev --host $$ip

.PHONY: sev-instance-start
sev-instance-start:
	@gcloud compute instances start $(sev_gcp_instance_name) \
		--zone=$(sev_gcp_zone)

.PHONY: sev-instance-stop
sev-instance-stop:
	@gcloud compute instances stop $(sev_gcp_instance_name) \
		--zone=$(sev_gcp_zone)

.PHONY: sev-instance-ssh
sev-instance-ssh:
	@gcloud compute ssh $(sev_gcp_instance_name)

################################################################################
# GCP Intel TDX
################################################################################
tdx_gcp_project_id=bearclave
tdx_gcp_zone=us-central1-a
tdx_gcp_instance_name=instance-bearclave-tdx
tdx_gcp_artifact_registry=us-east1-docker.pkg.dev/$(tdx_gcp_project_id)/bearclave
tdx_enclave_image_name=hello-world-enclave-tdx
tdx_enclave_image_tag=$(tdx_gcp_artifact_registry)/$(tdx_enclave_image_name)

.PHONY: tdx-enclave-build-image
tdx-enclave-build-image: enclave enclave-proxy
	@docker build \
		-f ./tdx.Dockerfile \
		-t $(tdx_enclave_image_name) \
		-t $(tdx_enclave_image_tag) \
		--build-arg CONFIG_FILE=configs/tdx-config.yaml \
		.

.PHONY: tdx-enclave-push-image
tdx-enclave-push-image: tdx-enclave-build-image
	@docker push $(tdx_enclave_image_tag)

.PHONY: tdx-enclave-run-image
tdx-enclave-run-image: tdx-enclave-push-image
	@echo "Retrieving the most recent image SHA for $(tdx_enclave_image_name)..."
	@sha=$(shell gcloud container images list-tags $(tdx_enclave_image_tag) --sort-by="~TIMESTAMP" --limit=1 --format=json | jq -r '.[0].digest') && \
	echo "Most recent image SHA: $$sha" && \
	echo "Updating GCP instance to use the new image with SHA $$sha..." && \
	gcloud compute instances update-container $(tdx_gcp_instance_name) \
		--zone $(tdx_gcp_zone) \
		--container-image "$(tdx_enclave_image_tag)@$$sha" && \
	echo "Restarting GCP instance $(tdx_gcp_instance_name)..." && \
	gcloud compute instances reset $(tdx_gcp_instance_name) --zone $(tdx_gcp_zone)

.PHONY: tdx-nonclave-run
tdx-nonclave-run:
	@ip=$$(gcloud compute instances describe $(tdx_gcp_instance_name) \
		--zone $(tdx_gcp_zone) \
		--format='get(networkInterfaces[0].accessConfigs[0].natIP)') && \
	go run ./nonclave/main.go --platform tdx --host $$ip

.PHONY: tdx-instance-start
tdx-instance-start:
	@gcloud compute instances start $(tdx_gcp_instance_name) \
		--zone=$(tdx_gcp_zone)

.PHONY: tdx-instance-stop
tdx-instance-stop:
	@gcloud compute instances stop $(tdx_gcp_instance_name) \
		--zone=$(tdx_gcp_zone)

.PHONY: tdx-instance-ssh
tdx-instance-ssh:
	@gcloud compute ssh $(tdx_gcp_instance_name)

################################################################################
# Clean
################################################################################
.PHONY: clean
clean:
	rm -rf ./enclave/bin
	rm -rf ./enclave-proxy/bin
	rm -rf ./nonclave/bin
