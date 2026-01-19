#!/bin/sh

OUTPUT_FILE=$(mktemp)
EXIT_FILE=$(mktemp)
trap 'rm -f $OUTPUT_FILE $EXIT_FILE' EXIT

# Run the test executable and capture output and exit code
/app/tdx-test -test.v > "$OUTPUT_FILE" 2>&1
echo $? > "$EXIT_FILE"

# Extract the output and exit status
TEST_OUTPUT=$(cat "$OUTPUT_FILE")
TEST_EXIT=$(cat "$EXIT_FILE")

# Keep the enclave alive by continuously printing output and status
while true; do
  clear
  echo "=== TDX Test Results ==="
  echo ""
  echo "$TEST_OUTPUT"
  echo ""
  echo "=== Exit Status: $TEST_EXIT ==="
  echo ""
  echo "Test completed. Press Ctrl+C to exit the enclave console."
  sleep 3600
done
