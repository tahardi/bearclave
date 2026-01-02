FROM alpine:latest
WORKDIR /app

COPY ./bin/nitro-test .
COPY ./nitro-run-test.sh .
RUN chmod +x ./nitro-run-test.sh

CMD ["/app/nitro-run-test.sh"]
