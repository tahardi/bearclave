#!/bin/bash
/app/enclave --config /app/enclave-config.yaml 2>&1 &
ENCLAVE_PID=$!

/app/enclave-proxy --config /app/proxy-config.yaml 2>&1 &
PROXY_PID=$!

# Handle process termination
trap 'kill $ENCLAVE_PID $PROXY_PID; exit' TERM INT

# Wait for either process to exit
wait -n $ENCLAVE_PID $PROXY_PID

# If we get here, one of the processes died, so exit with error
exit 1
