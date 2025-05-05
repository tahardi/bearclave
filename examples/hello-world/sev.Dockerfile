FROM golang:1.23-alpine AS builder
WORKDIR /app

ARG CONFIG_FILE=enclave/sev-config.yaml
COPY ./enclave/bin/enclave .
COPY ./gateway/bin/gateway .
COPY ./sev-run.sh .
RUN chmod +x ./sev-run.sh
COPY ./${CONFIG_FILE} ./config.yaml

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/enclave .
COPY --from=builder /app/gateway .
COPY --from=builder /app/sev-run.sh .
COPY --from=builder /app/config.yaml .
CMD ["/app/sev-run.sh"]
