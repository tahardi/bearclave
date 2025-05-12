# TODOs

## Implementation
- **Replace outdated libraries**:
    - `hf/nsm` and `hf/nitrite`: These libraries are no longer regularly updated.  
      Write custom implementations to handle security updates effectively and  
      enable better testing capabilities.
    - `sev-guest` and `tdx-guest`: While actively maintained, consider replacing  
      them to gain full control over enclave code, improve testing, and define a  
      device interface for better generality.
    - `vsock`: Replace this outdated implementation with a custom version.

## IAM Configuration
- **AWS IAM**:
    - Reduce permission sets for developers.
    - Combine billing and developer roles.
- **GCP IAM**:
    - Understand roles, service accounts, and IAM structure in GCP.
    - Implement IAM best practices.

## Testing
- **Unit tests**:
    - Write verification tests using real attestation strings.
    - Add unit tests once a device interface is defined.

## Code Maintenance
- Rename `enclave-proxy` and `unsafe` for better clarity.
- Update the attestation interface to accept `userdata [64]byte`.

## Build Process
- Improve binary builds:
    - Use a multi-stage Dockerfile to streamline the process.
    - Ensure the build targets are optimized for Linux.

## Networking
- **EC2 Networking**:
    - Resolve configuration issues preventing `enclave-proxy` from being called  
      via the public EC2 IP.
- **Networking Lockdown**:
    - Secure instances on AWS and GCP to prevent open access to the world.
    - Consider using an API Gateway:
        - Implement authentication.
        - Configure instances to only accept connections from the gateway.