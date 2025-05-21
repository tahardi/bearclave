#!/bin/bash
# Run our enclave and enclave-proxy programs in the background. Capture their
# STDOUT and STDERR output and prepend each line with the program name for
# better readability as their outputs are intermingled.
/app/enclave-server-1 --config /app/config.yaml 2>&1 | awk '{ print "[enclave-server-1] " $0; fflush(); }' &
ENCLAVE_SERVER_1_PID=$!

/app/enclave-server-2 --config /app/config.yaml 2>&1 | awk '{ print "[enclave-server-2] " $0; fflush(); }' &
ENCLAVE_SERVER_2_PID=$!

/app/enclave-proxy --config /app/config.yaml 2>&1 | awk '{ print "[enclave-proxy] " $0; fflush(); }' &
PROXY_PID=$!

# Wait for either process to exit
trap 'kill $ENCLAVE_SERVER_1_PID $ENCLAVE_SERVER_2_PID $PROXY_PID; exit' TERM INT
wait -n $ENCLAVE_SERVER_1_PID $ENCLAVE_SERVER_2_PID $PROXY_PID
exit 1
