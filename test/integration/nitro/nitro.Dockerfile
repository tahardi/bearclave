FROM alpine:latest
WORKDIR /app

COPY ./bin/nitro-test .

CMD ["/app/nitro-test"]
