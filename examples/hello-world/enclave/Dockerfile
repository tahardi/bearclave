FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY ./bin/enclave .

FROM scratch
WORKDIR /app
COPY --from=builder /app/enclave .
CMD ["/app/enclave --platform nitro --nonclaveCID 3"]