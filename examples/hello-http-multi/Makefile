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
# AWS - Targets for deploying to AWS-supported TEE platforms
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
aws_nitro_nonclave_config=configs/nonclave/nitro.yaml
aws_nitro_enclave_config=configs/enclave/nitro.yaml
aws_nitro_enclave_image_tag=hello-http-multi-enclave-nitro
aws_nitro_enclave_eif_path=./enclaves.eif
aws_nitro_enclave_cpu_count=2
aws_nitro_enclave_memory_size=512
aws_nitro_enclave_cid=4

.PHONY: aws-nitro-enclave-build-image
aws-nitro-enclave-build-image: enclave-server-1 enclave-server-2
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
	go run ./enclave-proxy/main.go --config $(aws_nitro_enclave_config)

.PHONY: aws-nitro-nonclave-run
aws-nitro-nonclave-run:
	go run ./nonclave/main.go \
		--config $(aws_nitro_nonclave_config) \
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
# GCP - Targets for deploying to GCP-supported TEE platforms
################################################################################
gcp_project_id=bearclave
gcp_zone=us-central1-a
gcp_artifact_registry=us-east1-docker.pkg.dev/$(gcp_project_id)/bearclave

# GCP AMD SEV-SNP
################################################################################
gcp_sev_instance_name=instance-bearclave-sev-snp
gcp_sev_instance_ip=$(shell gcloud compute instances describe \
	$(gcp_sev_instance_name) \
	--zone $(gcp_zone) \
	--format='get(networkInterfaces[0].accessConfigs[0].natIP)')
gcp_sev_nonclave_config=configs/nonclave/sev.yaml
gcp_sev_enclave_config=configs/enclave/sev.yaml
gcp_sev_enclave_image_name=hello-http-multi-enclave-sev
gcp_sev_enclave_image_tag=$(gcp_artifact_registry)/$(gcp_sev_enclave_image_name)
gcp_sev_enclave_image_sha=$(shell gcloud container images list-tags \
	$(gcp_sev_enclave_image_tag) \
	--sort-by="~TIMESTAMP" \
	--limit=1 \
	--format=json \
	| jq -r '.[0].digest')

.PHONY: gcp-sev-enclave-build-image
gcp-sev-enclave-build-image: enclave-server-1 enclave-server-2 enclave-proxy
	@docker build \
		-f ./sev.Dockerfile \
		-t $(gcp_sev_enclave_image_name) \
		-t $(gcp_sev_enclave_image_tag) \
		--build-arg CONFIG_FILE=$(gcp_sev_enclave_config) \
		.

.PHONY: gcp-sev-enclave-push-image
gcp-sev-enclave-push-image: gcp-sev-enclave-build-image
	@docker push $(gcp_sev_enclave_image_tag)

.PHONY: gcp-sev-enclave-run-image
gcp-sev-enclave-run-image: gcp-sev-enclave-push-image
	gcloud compute instances update-container $(gcp_sev_instance_name) \
		--zone $(gcp_zone) \
		--container-image "$(gcp_sev_enclave_image_tag)@$(gcp_sev_enclave_image_sha)" && \
	gcloud compute instances reset $(gcp_sev_instance_name) --zone $(gcp_zone)

.PHONY: gcp-sev-nonclave-run
gcp-sev-nonclave-run:
	go run ./nonclave/main.go --config $(gcp_sev_nonclave_config) \
 		--host $(gcp_sev_instance_ip)

.PHONY: gcp-sev-instance-start
gcp-sev-instance-start:
	@gcloud compute instances start $(gcp_sev_instance_name) \
		--zone=$(gcp_zone)

.PHONY: gcp-sev-instance-stop
gcp-sev-instance-stop:
	@gcloud compute instances stop $(gcp_sev_instance_name) \
		--zone=$(gcp_zone)

.PHONY: gcp-sev-instance-ssh
gcp-sev-instance-ssh:
	@gcloud compute ssh $(gcp_sev_instance_name)

# GCP Intel TDX
################################################################################
gcp_tdx_instance_name=instance-bearclave-tdx
gcp_tdx_instance_ip=$(shell gcloud compute instances describe \
	$(gcp_tdx_instance_name) \
	--zone $(gcp_zone) \
	--format='get(networkInterfaces[0].accessConfigs[0].natIP)')
gcp_tdx_nonclave_config=configs/nonclave/tdx.yaml
gcp_tdx_enclave_config=configs/enclave/tdx.yaml
gcp_tdx_enclave_image_name=hello-http-enclave-tdx
gcp_tdx_enclave_image_tag=$(gcp_artifact_registry)/$(gcp_tdx_enclave_image_name)
gcp_tdx_enclave_image_sha=$(shell gcloud container images list-tags \
	$(gcp_tdx_enclave_image_tag) \
	--sort-by="~TIMESTAMP" \
	--limit=1 \
	--format=json \
	| jq -r '.[0].digest')

.PHONY: gcp-tdx-enclave-build-image
gcp-tdx-enclave-build-image: enclave-server-1 enclave-server-2 enclave-proxy
	@docker build \
		-f ./tdx.Dockerfile \
		-t $(gcp_tdx_enclave_image_name) \
		-t $(gcp_tdx_enclave_image_tag) \
		--build-arg CONFIG_FILE=$(gcp_tdx_enclave_config) \
		.

.PHONY: gcp-tdx-enclave-push-image
gcp-tdx-enclave-push-image: gcp-tdx-enclave-build-image
	@docker push $(gcp_tdx_enclave_image_tag)

.PHONY: gcp-tdx-enclave-run-image
gcp-tdx-enclave-run-image: gcp-tdx-enclave-push-image
	gcloud compute instances update-container $(gcp_tdx_instance_name) \
		--zone $(gcp_zone) \
		--container-image "$(gcp_tdx_enclave_image_tag)@$(gcp_tdx_enclave_image_sha)" && \
	gcloud compute instances reset $(gcp_tdx_instance_name) --zone $(gcp_zone)

.PHONY: gcp-tdx-nonclave-run
gcp-tdx-nonclave-run:
	go run ./nonclave/main.go --config $(gcp_tdx_nonclave_config) \
		--host $(gcp_tdx_instance_ip)

.PHONY: gcp-tdx-instance-start
gcp-tdx-instance-start:
	@gcloud compute instances start $(gcp_tdx_instance_name) \
		--zone=$(gcp_zone)

.PHONY: gcp-tdx-instance-stop
gcp-tdx-instance-stop:
	@gcloud compute instances stop $(gcp_tdx_instance_name) \
		--zone=$(gcp_zone)

.PHONY: gcp-tdx-instance-ssh
gcp-tdx-instance-ssh:
	@gcloud compute ssh $(gcp_tdx_instance_name)

################################################################################
# Clean
################################################################################
.PHONY: clean
clean:
	rm -rf ./enclave-server-1/bin
	rm -rf ./enclave-server-2/bin
	rm -rf ./enclave-proxy/bin
	rm -rf ./nonclave/bin