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
	@-make --no-print-directory notee-start-processes
	@-make --no-print-directory notee-stop-processes

# Make sure your process-compose port does not conflict with
# the ports used by enclave, proxy, or nonclave processes
# https://github.com/F1bonacc1/process-compose
PROCESS_COMPOSE_PORT=8083
.PHONY: notee-start-processes
notee-start-processes:
	@-process-compose up --tui=false --port=${PROCESS_COMPOSE_PORT} -f .process-compose.yaml 2> /dev/null

.PHONY: notee-stop-processes
notee-stop-processes:
	@-process-compose down --port=${PROCESS_COMPOSE_PORT} 2> /dev/null

################################################################################
# Clean
################################################################################
.PHONY: clean
clean:
	rm -rf ./enclave/bin
	rm -rf ./enclave-proxy/bin
	rm -rf ./nonclave/bin