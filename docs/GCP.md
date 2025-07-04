# Google Cloud Platform (GCP) Setup Guide
Google Cloud Platform (GCP) provides compute instances that support the
AMD SEV-SNP and Intel TDX TEE platforms. Follow the steps below to sign up for
a Google Cloud account and configure the cloud resources required to develop
on AMD SEV-SNP and Intel TDX.

---

### Configure Google Cloud
1. **Create a [Google Account](https://accounts.google.com/)** If you use GMail,
  Drive, or any other similar Google service, then you already have an account;
  feel free to use that account.

2. **Setup Billing** Navigate to the Google Cloud
  [Console](https://console.cloud.google.com/). At the time of this writing 
  (June, 2025), Google offers $300 in credits for new Google Cloud users. Sign
  up if it is still available. Regardless, you will need to add a valid Billing
  method to your account to use TEE-enabled cloud resources.

---

### Install and Configure the GCloud CLI Tool
I do not think it is possible to properly configure an SEV or TDX-enabled
compute instance purely through the web UI. At least, I have been unable to
figure out how to do so. Thus, we will use the `gcloud` cli tool for
creating, configuring, and destroying compute instances (and other cloud-
related resources).

1. **Install the [`gcloud` CLI]((https://cloud.google.com/sdk/docs/install))**

2. **Initialize `gcloud`** login to your google account with `gcloud` and
configure default settings.
    ```bash
    # This will walk you through logging into your google account and then
    # setting defaults for when you use `gcloud`. Note that you are free to
    # choose your own defaults, but the commands in `examples/` assume:
    #   region: us-central1
    #   zone: us-central1-a
    #   project-id: bearclave
    #
    # If you choose different defaults you will have to edit the examples Makefiles
    gcloud init
    ```

---

### Create an Image Repository
To run your application on GCP TEE-enabled compute instances, you first need to
package it as an OCI-compliant image and upload it to the GCP
[Artifact Registry](https://console.cloud.google.com/artifacts).

1. **Create an image repository** Create a new repo in your Artifact Registry
  for storing images. Note that `examples/` assumes the repo is located in
  `us-east1` and called `bearclave`.
   ```bash
    gcloud artifacts repositories create bearclave \
    --repository-format=docker \
    --location=us-east1 \
    --description="Docker repository for bearclave images" \
    --project=bearclave
    ```
2. **Configure Docker** Update your `docker` config to use `gcloud` for
  authentication when interacting with the Artifact Registry.
   ```bash
    gcloud auth configure-docker us-east1-docker.pkg.dev
   ```

---

### Create SEV and TDX Compute Instances
1. **Allow HTTP Traffic** Update the default firewall policy to allow for
  http traffic on port 8080 as that is what is used in `examples/`.

    ```bash
    gcloud compute firewall-rules update default-allow-http \
    --target-tags=http-server \
    --allow=tcp:8080
    ```

2. **Create an SEV-enabled Compute Instance** At the time of this writing, the
  `n2d-standard-*` instances provide SEV-SNP and run between $0.08-$0.34/hr.
  Note that you need to create the instance with a valid image. Thus, we
  initially deploy with a simple `hello-world` image and will replace it later
  with one of the TEE applications in `examples/`.

    ```bash
    # Usage
    gcloud compute instances create-with-container <instance-name> \
    --project= \
    --zone= \
    --machine-type= \
    --confidential-compute-type=SEV_SNP \
    --container-privileged \
    # Specify the artifact registry string for the image of the confidential
    # workload that you wish to run inside the SEV-SNP TEE
    --container-image= \
    # You must mount the `/dev/sev-guest` device to the Confidential VM so
    # that your workload can generate attestations
    --container-mount-host-path mount-path=/dev/sev-guest,host-path=/dev/sev-guest \
    # Add this tag to enable external http traffic to your Confidential VM
    --tags=http-server \
    --scopes=cloud-platform \
    --maintenance-policy=TERMINATE \
    --shielded-secure-boot
    
    # Example (n2d-standard-2 $0.08/hr, n2d-standard-4 $0.17/hr, n2d-standard-8 $0.34/hr)
    gcloud compute instances create-with-container instance-bearclave-sev \
    --project=bearclave \
    --zone=us-central1-a \
    --machine-type=n2d-standard-8 \
    --confidential-compute-type=SEV_SNP \
    --container-privileged \
    --container-image=hello-world \
    --container-mount-host-path mount-path=/dev/sev-guest,host-path=/dev/sev-guest \
    --tags=http-server \
    --scopes=cloud-platform \
    --maintenance-policy=TERMINATE \
    --shielded-secure-boot
    ```

3. **Create an TDX-enabled Compute Instance** At the time of this writing, the
   `c3-standard-*` instances provide TDX and run between $0.20-$0.40/hr.

    ```bash
    # Usage
    gcloud compute instances create-with-container <instance-name> \
    --project= \
    --zone= \
    --machine-type= \
    --confidential-compute-type=TDX \
    --container-privileged \
    # Specify the artifact registry string for the image of the confidential
    # workload that you wish to run inside the TDX TEE
    --container-image= \
    # You must mount the `/dev/tdx-guest` device to the Confidential VM so
    # that your workload can generate attestations
    --container-mount-host-path mount-path=/dev/tdx-guest,host-path=/dev/tdx-guest \
    # You also need to mount this so tdx can access /sys/kernel/config/tdx/report
    --container-mount-host-path mount-path=/sys/kernel/config,host-path=/sys/kernel/config \
    # Add this tag to enable external http traffic to your Confidential VM     
    --tags=http-server \
    --scopes=cloud-platform \
    --maintenance-policy=TERMINATE \
    --shielded-secure-boot
   
    # Example (c3-standard-4 $0.20/hr, c3-standard-8 $0.40/hr)
    gcloud compute instances create-with-container instance-bearclave-tdx \
    --project=bearclave \
    --zone=us-central1-a \
    --machine-type=c3-standard-4 \
    --confidential-compute-type=TDX \
    --container-privileged \
    --container-image=hello-world \
    --container-mount-host-path mount-path=/dev/tdx-guest,host-path=/dev/tdx-guest \
    --container-mount-host-path mount-path=/sys/kernel/config,host-path=/sys/kernel/config \
    --tags=http-server \
    --scopes=cloud-platform \
    --maintenance-policy=TERMINATE \
    --shielded-secure-boot
    ```
