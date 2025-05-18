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
- client and server via vsock
  - vsock implements the net.Listener interface so you can pass it to a
  normal golang http.Server.Serve(vsockListener) and it should receive
  connections over vsock
  - you can make a custom `http.Transport` that dials vsock instead of
  tcp
```golang
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/mdlayher/vsock"
)

// vsockDialer is a custom dialer that supports vsock connections.
func vsockDialer(contextID uint32, port uint32) func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		// Use vsock.Dial to establish a connection
		conn, err := vsock.Dial(contextID, port)
		if err != nil {
			return nil, fmt.Errorf("failed to dial vsock: %v", err)
		}
		return conn, nil
	}
}

func main() {
	// Replace these with the appropriate vsock CID and port
	const enclaveCID = 3  // CID of the target Nitro Enclave
	const enclavePort = 5000 // vsock listening port in the enclave

	// Create an HTTP transport that uses the vsock dialer
	transport := &http.Transport{
		Dial: vsockDialer(enclaveCID, enclavePort),
	}

	// Create the HTTP client using the custom transport
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, // Optional: Timeout for the requests
	}

	// Make an HTTP request to the vsock server
	resp, err := client.Get("http://vsock")
	if err != nil {
		log.Fatalf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Print the response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to read response: %v", err)
	}
}
```

- **EC2 Networking**:
    - Resolve configuration issues preventing `enclave-proxy` from being called  
      via the public EC2 IP.
- **Networking Lockdown**:
    - Secure instances on AWS and GCP to prevent open access to the world.
    - Consider using an API Gateway:
        - Implement authentication.
        - Configure instances to only accept connections from the gateway.