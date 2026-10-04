#!/bin/sh
OUTPUT_FILE=$(mktemp)
/app/sev-test -test.v > "$OUTPUT_FILE" 2>&1
TEST_EXIT=$?
while true; do
  cat "$OUTPUT_FILE"
  echo "BEARCLAVE_TEST_EXIT=$TEST_EXIT"
  sleep 30
done
