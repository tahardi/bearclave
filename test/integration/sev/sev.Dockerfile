FROM alpine:latest
WORKDIR /app

COPY ./bin/sev-test .
COPY ./sev-run-test.sh .
RUN chmod +x ./sev-run-test.sh

CMD ["/app/sev-run-test.sh"]
