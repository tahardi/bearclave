# https://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := hello-world
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: hello-world
hello-world:
	@make --no-print-directory servers-up

# https://github.com/F1bonacc1/process-compose
PROCESS_COMPOSE_PORT=8082
.PHONY: servers-down
servers-down:
	@process-compose down --port=${PROCESS_COMPOSE_PORT}

.PHONY: servers-up
servers-up:
	@process-compose up --tui=false --port=${PROCESS_COMPOSE_PORT} -f .process-compose.yaml
