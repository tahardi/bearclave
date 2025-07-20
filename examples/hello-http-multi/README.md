# Hello HTTP Multi
Bearclave adapts to different TEE platforms by requiring two primary programs:
- **Enclave Servers**: Securely run within the TEE. In this example, two 
separate HTTP servers ([enclave-server-1](./enclave-server-1/main.go) and 
[enclave-server-2](./enclave-server-2/main.go)) operate within the TEE
environment, each configured to handle requests on different routes.
- **Proxy**: Acts as an HTTP reverse proxy, routing client requests to the
appropriate enclave server based on the incoming request's route. It uses
traditional sockets or VSOCK, depending on the platform.

For AWS Nitro Enclaves, which use VSOCK interfaces, the proxy translates HTTP
requests into VSOCK communication and routes them to the correct enclave server.
On AMD SEV-SNP and Intel TDX, where networking is directly accessible within the
TEE, the proxy operates over standard networking.

### Workflow Example
This example demonstrates an HTTP-based setup with multiple enclave servers:
1. A **remote client** ("nonclave") sends an HTTP request to the proxy with a
route (`/v1` or `/v2`) in the URL and user data.
2. The **proxy** determines the appropriate enclave server to forward the
request to based on the specified route:
    - `/v1` -> **enclave-server-1**
    - `/v2` -> **enclave-server-2**

3. The **selected enclave server** processes the incoming HTTP request,
generates an attestation report, and includes it in the HTTP response.
4. The **proxy** relays the server's response back to the client.
5. The **nonclave** verifies the attestation report and extracts the attested
data.

### Key Points
- Each enclave server supports a specific route and processes requests
independently, enabling a multi-server enclave setup.
- The proxy transparently handles incoming requests and uses route-based
forwarding, simplifying TEE application design with multiple enclaves.
- The solution works across AWS Nitro, AMD SEV-SNP, and Intel TDX by managing
platform-specific differences within the proxy.

## Run the example locally
Bearclave supports a "No TEE" mode (`notee`) for running code locally. This
allows for quick iteration and saves on Cloud infrastructure costs. Assuming
you have installed the minimum set of dependencies listed in the
[top-level README](../../README.md), you can run this example locally with:

```bash
make

# You should see output similar to:
[enclave-proxy  ] time=2025-07-20T11:51:37.096-04:00 level=INFO msg="loaded config" configs/enclave/notee.yaml="&{Platform:notee Measurement: IPCs:map[] Servers:map[enclave-server-1:{CID:4 Port:8082 Route:/v1} enclave-server-2:{CID:5 Port:8083 Route:/v2}] Proxy:{Port:8080}}"
[enclave-proxy  ] time=2025-07-20T11:51:37.096-04:00 level=INFO msg="Proxy server started" addr=0.0.0.0:8080
[enclave-server-1       ] time=2025-07-20T11:51:37.136-04:00 level=INFO msg="loaded config" configs/enclave/notee.yaml="&{Platform:notee Measurement: IPCs:map[] Servers:map[enclave-server-1:{CID:4 Port:8082 Route:/v1} enclave-server-2:{CID:5 Port:8083 Route:/v2}] Proxy:{Port:8080}}"
[enclave-server-1       ] time=2025-07-20T11:51:37.136-04:00 level=INFO msg="Enclave server 1 started" addr=127.0.0.1:8082
[enclave-server-2       ] time=2025-07-20T11:51:37.149-04:00 level=INFO msg="loaded config" configs/enclave/notee.yaml="&{Platform:notee Measurement: IPCs:map[] Servers:map[enclave-server-1:{CID:4 Port:8082 Route:/v1} enclave-server-2:{CID:5 Port:8083 Route:/v2}] Proxy:{Port:8080}}"
[enclave-server-2       ] time=2025-07-20T11:51:37.149-04:00 level=INFO msg="Enclave server 2 started" addr=127.0.0.1:8083
[nonclave       ] time=2025-07-20T11:51:37.496-04:00 level=INFO msg="loaded config" configs/nonclave/notee.yaml="&{Platform:notee Measurement:Not a TEE platform. Code measurements are not real. IPCs:map[] Servers:map[enclave-server-1:{CID:0 Port:0 Route:/v1} enclave-server-2:{CID:0 Port:0 Route:/v2}] Proxy:{Port:8080}}"
[nonclave       ] time=2025-07-20T11:51:37.496-04:00 level=INFO msg="sending request to url" url=http://127.0.0.1:8080/v2
[enclave-server-2       ] time=2025-07-20T11:51:37.497-04:00 level=INFO msg="attesting userdata" userdata="Hello from enclave-server-2!"
[nonclave       ] time=2025-07-20T11:51:37.498-04:00 level=INFO msg="verified userdata" userdata="Hello from enclave-server-2!"
[nonclave       ] time=2025-07-20T11:51:37.498-04:00 level=INFO msg="sending request to url" url=http://127.0.0.1:8080/v1
[enclave-server-1       ] time=2025-07-20T11:51:37.498-04:00 level=INFO msg="attesting userdata" userdata="Hello from enclave-server-1!"
[nonclave       ] time=2025-07-20T11:51:37.499-04:00 level=INFO msg="verified userdata" userdata="Hello from enclave-server-1!"
```

## Run the example on AWS Nitro
Recall that AWS Nitro Enclave development requires building and deploying
from the EC2 instance itself. Assuming you have configured your AWS account
and the necessary cloud resources laid out in the
[AWS setup guide](../../docs/AWS.md), you can run the example with:

1. Login to the AWS cli
    ```bash
    make aws-cli-login 
    ```
2. Start your EC2 instance
    ```bash
    make aws-nitro-instance-start
   ```
3. Log into your EC2 instance
    ```bash
   make aws-nitro-instance-ssh 
   ```
4. Clone (if needed) bearclave and cd to example
    ```bash
   git clone git@github.com:tahardi/bearclave.git
   cd bearclave/examples/hello-http-multi
   ```
5. Build and run the enclave program
    ```bash
    make aws-nitro-enclave-run-eif
    
    # This will attach a debug console to the enclave program so you can see
    # its log. Note that this will change the code measurement, however, to
    # indicate the enclave is in debug mode (and thus may leak sensitive info)
    make aws-nitro-enclave-run-eif-debug
    ```
6. Run the proxy program
    ```bash
   make aws-nitro-proxy-run 
   ```
7. In a separate terminal window, run the nonclave program. Remember this is a
   remote client making an HTTP request so you can run it from your local machine
    ```bash
    make aws-nitro-nonclave-run
    ```
8. Remember to turn off your EC2 instance when you are finished. Otherwise, you
   will continue to incur AWS cloud charges.
    ```bash
    make aws-nitro-instance-stop 
    ```

## Run the example on GCP AMD SEV-SNP or Intel TDX
Unlike AWS Nitro, the GCP AMD and Intel programs are built and deployed from
your local machine. Assuming you have configured your GCP account and the
necessary cloud resources laid out in the [GCP setup guide](../../docs/GCP.md),
you can run the example with:

1. Start your Compute instance. Note that these steps demonstrate running the
   example on AMD SEV-SNP, but you can simply replace `sev` with `tdx` in the make
   commands as the process is exactly the same.
    ```bash
    make gcp-sev-instance-start
   ```
2. Build and deploy the enclave and proxy programs. Note that unlike AWS Nitro,
   the enclave and proxy are bundled together and both deployed within the TEE
   when run on SEV or TDX.
    ```bash
    make gcp-sev-enclave-run-image
    ```
3. Run the nonclave program
    ```bash
   make gcp-sev-nonclave-run
   ```
4. Remember to turn off your Compute instance when you are finished. Otherwise,
   you will continue to incur GCP cloud charges.
    ```bash
    make gcp-sev-instance-stop 
    ```