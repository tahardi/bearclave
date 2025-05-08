FROM golang:1.23-alpine AS builder
WORKDIR /app

ARG ENCLAVE_CONFIG_FILE=enclave/sev-config.yaml
ARG PROXY_CONFIG_FILE=enclave-proxy/sev-config.yaml
COPY ./${ENCLAVE_CONFIG_FILE} ./enclave-config.yaml
COPY ./${PROXY_CONFIG_FILE} ./proxy-config.yaml
COPY ./enclave/bin/enclave .
COPY ./enclave-proxy/bin/enclave-proxy .
COPY ./sev-run.sh .
RUN chmod +x ./sev-run.sh

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/enclave .
COPY --from=builder /app/enclave-proxy .
COPY --from=builder /app/sev-run.sh .
COPY --from=builder /app/enclave-config.yaml .
COPY --from=builder /app/proxy-config.yaml .
CMD ["/bin/sh", "/app/sev-run.sh"]
