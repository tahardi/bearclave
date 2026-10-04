# Integration Tests

These tests need real TEE hardware. `make test-integration` only checks that they compile.

## Build

- Nitro: `make test-integration-nitro` builds `test/integration/nitro/bin/nitro-test.eif`. Set `NITRO_CLI` to run
  nitro-cli from a Docker image instead of the host.
- SEV: `make test-integration-sev` builds and pushes the image.
- TDX: `make test-integration-tdx` builds and pushes the image.

Set `GCP_REGISTRY` and `IMAGE_TAG` to change where SEV and TDX images go. Run
`make -C test/integration/<sev|tdx> image-ref` to print the full image reference.

## Run

- Nitro: on a Nitro host, run `make -C test/integration/nitro run`.
- SEV and TDX: deploy the pushed image with the bearclave-tf GCP module, then read the output with
  `gcloud compute instances get-serial-port-output`.

## Result

Each test prints the full `go test -v` output, then `BEARCLAVE_TEST_EXIT=<exit code>`. Both repeat every 30 seconds.
Exit code 0 means the tests passed.
