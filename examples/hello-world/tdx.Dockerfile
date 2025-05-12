FROM alpine:latest

# Add tini for better process management
RUN apk add --no-cache tini bash

WORKDIR /app

ARG ENCLAVE_CONFIG_FILE=enclave/tdx-config.yaml
ARG PROXY_CONFIG_FILE=enclave-proxy/tdx-config.yaml

COPY ./${ENCLAVE_CONFIG_FILE} ./enclave-config.yaml
COPY ./${PROXY_CONFIG_FILE} ./proxy-config.yaml
COPY ./enclave/bin/enclave .
COPY ./enclave-proxy/bin/enclave-proxy .
COPY ./tdx-run.sh .
RUN chmod +x ./tdx-run.sh

# Use tini as the entry point
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/tdx-run.sh"]
