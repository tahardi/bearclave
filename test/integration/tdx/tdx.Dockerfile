FROM alpine:latest
WORKDIR /app

COPY ./bin/tdx-test .
COPY ./tdx-run-test.sh .
RUN chmod +x ./tdx-run-test.sh

CMD ["/app/tdx-run-test.sh"]
