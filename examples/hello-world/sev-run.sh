#!/bin/bash
# Run our enclave and proxy in the background and capture both their STDOUT and STDERR output.
# Prepend each line of output with the program name for better readability as both the
# enclave and proxy output will intermingle in the terminal. Tell sed not to buffer output.
/app/enclave --config /app/enclave-config.yaml 2>&1 | sed -u 's/^/[enclave] /' &
ENCLAVE_PID=$!

/app/enclave-proxy --config /app/proxy-config.yaml 2>&1 | sed -u 's/^/[enclave-proxy] /' &
PROXY_PID=$!

# Wait for either process to exit
trap 'kill $ENCLAVE_PID $PROXY_PID; exit' TERM INT
wait -n $ENCLAVE_PID $PROXY_PID
exit 1
