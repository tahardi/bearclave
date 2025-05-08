FROM golang:1.23-alpine AS builder
WORKDIR /app

ARG CONFIG_FILE=nitro-config.yaml
COPY ./${CONFIG_FILE} ./config.yaml
COPY ./enclave/bin/enclave .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/enclave .
COPY --from=builder /app/config.yaml .
CMD ["/app/enclave", "--config", "./config.yaml"]
