FROM alpine:latest

# Add tini for better process management
RUN apk add --no-cache tini bash

WORKDIR /app

ARG CONFIG_FILE=configs/nitro-config.yaml
COPY ./${CONFIG_FILE} ./config.yaml
COPY ./enclave-server-1/bin/enclave-server-1 .
COPY ./enclave-server-2/bin/enclave-server-2 .
COPY ./nitro-run.sh .
RUN chmod +x ./nitro-run.sh

# Use tini as the entry point
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/nitro-run.sh"]
