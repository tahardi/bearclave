FROM alpine:latest

# Add tini for better process management
RUN apk add --no-cache tini bash

WORKDIR /app

ARG ENCLAVE_CONFIG_FILE=enclave/sev-config.yaml
ARG PROXY_CONFIG_FILE=enclave-proxy/sev-config.yaml

COPY ./${ENCLAVE_CONFIG_FILE} ./enclave-config.yaml
COPY ./${PROXY_CONFIG_FILE} ./proxy-config.yaml
COPY ./enclave/bin/enclave .
COPY ./enclave-proxy/bin/enclave-proxy .
COPY ./sev-run.sh .
RUN chmod +x ./sev-run.sh

# Use tini as the entry point
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/sev-run.sh"]