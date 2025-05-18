# GCP AMD SEV-SNP and Intel-TDX: Overview and Development Setup Guide
The Google Cloud Platform (GCP) provides compute instances that support the
AMD SEV-SNP and Intel-TDX TEE platforms. Unlike AWS Nitro Enclaves, these
platforms are not tied to GCP specifically---you could spin up similar instances
on Azure or buy the hardware yourself. That said, there may be setup and
development idiosyncrasies introduced by GCP (not the TEE platform itself).
Thus, while the `sev` and `tdx` code in this repository _should_ run on Azure
instances, it is important to note that this has not been tested and there
may be Azure-specific deployment hiccups not accounted in this document and/or
code.

## What is AMD SEV-SNP?

**TODO**

---

## What is Intel-TDX? 

**TODO**

---

## Setting Up Your Environment for AMD SEV-SNP and Intel-TDX Development on GCP

To start developing you will need to launch and configure GCP compute instances
configured for SEV-SNP and TDX. Note that **THIS WILL COST MONEY**. In fact,
developing on any cloud-based TEE platform is going to cost money. That said,
the smallest sev and tdx-enabled instances only run between $0.20-0.40/hr.
Coupled with the Bearclave "No TEE" platform implementation, you should be able
to develop locally and minimize the time needed to test and run on actual TEE
hardware.

---

### Step 1: Create and Configure Your Google Cloud Account

1. **Create a Google Cloud Account**
   At the time of writing this document (May 18, 2025), Google gives $300 in
   free cloud credits (expires after 3 months) for 
   [creating a new account](https://cloud.google.com/free?hl=en).
   This is way more than you need to try out TEEs and well worth the trouble.

2. **Create a Project** Create a project to organize cloud resources under
   (e.g., `bearclave-test`)

3. **Locate Service Account** Locate the name of your GCP service account.
   This is what is used by `gcloud` to authenticate and interact with
   cloud resources.

4. **Enable Artifact Registry API** Your Confidential VM "workloads" should be
   packaged as docker images and uploaded to the GCP Artifact Registry because
   this is where the compute instance looks for the specified image when
   starting up.

---

### Step 2: Install and Configure the GCloud Command Line Interface (CLI)

I do not think it is possible to properly configure an SEV or TDX-enabled
compute instance purely through the web UI. At least, I have been unable to
figure out how to do so. Thus, I recommend installing the `gcloud` cli tool
for creating, configuring, and destroying compute instances (and other cloud
related resources).

1. **Install and Configure GCloud CLI**
   Install and set up the CLI using the
   [GCloud CLI guide](https://cloud.google.com/sdk/docs/install).

2. **Setup `gcloud`**
   Run the following command to sign in to your google cloud account with
   `gcloud` and select a default project for it to associate with.
   ```bash
    gcloud init
    # I think this is only needed if you want to switch accounts?
    gcloud auth login
    
    # e.g., 403490793521-compute@developer.gserviceaccount.com
    gcloud config set account <your-google-cloud-service-account>
   
    # Configure docker to work with GCP Artifact Registry
    # e.g., us-east1-docker.pkg.dev
    gcloud auth configure-docker <artifact-registry>
    
    # Add a policy allowing your service account to push to the registry
    # under your specified project
    # e.g., bearclave
    gcloud artifacts repositories add-iam-policy-binding <project-id> \
    --location=us-east1 \
    --member=serviceAccount:<your-google-cloud-service-account> \
    --role=roles/artifactregistry.writer
   ```

---

### Step 3: Create SEV and TDX Compute Instances

1. **Create an SEV-enabled Compute Instance** At the time of this writing, the
    `n2d-standard-*` instances provide SEV-SNP an run between $0.08-$0.34/hr.

    ```bash
    # Usage
    gcloud compute instances create-with-container <instance-name> \
        --project= \
        --zone= \
        --machine-type= \
        --confidential-compute-type=SEV_SNP \
        --container-privileged \
        --container-image= \
        --container-mount-host-path mount-path=/dev/sev-guest,host-path=/dev/sev-guest \
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
        # Specify the artifact registry string for the image of the confidential
        # workload that you wish to run inside the SEV-SNP TEE
        --container-image=us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-tdx@sha256:73267a52b7e026cf63e1e8d680af8985f7cb9d252a09175ea8bf024069e01221 \
        # You must mount the `/dev/sev-guest` device to the Confidential VM so
        # that your workload can generate attestations
        --container-mount-host-path mount-path=/dev/sev-guest,host-path=/dev/sev-guest \
        # Add this tag to enable external http traffic to your Confidential VM
        --tags=http-server \
        --scopes=cloud-platform \
        --maintenance-policy=TERMINATE \
        --shielded-secure-boot
    ```

2. **Create an TDX-enabled Compute Instance** At the time of this writing, the
   `c3-standard-*` instances provide TDX and run between $0.20-$0.40/hr.

    ```bash
    # Usage
    gcloud compute instances create-with-container <instance-name> \
        --project= \
        --zone= \
        --machine-type= \
        --confidential-compute-type=TDX \
        --container-privileged \
        --container-image= \
        --container-mount-host-path mount-path=/dev/tdx-guest,host-path=/dev/tdx-guest \
        --container-mount-host-path mount-path=/sys/kernel/config,host-path=/sys/kernel/config \     
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
        # Specify the artifact registry string for the image of the confidential
        # workload that you wish to run inside the TDX TEE
        --container-image=us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-tdx@sha256:73267a52b7e026cf63e1e8d680af8985f7cb9d252a09175ea8bf024069e01221 \
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
    ```
   
---

## Other useful `gcloud` commands

```bash
# If you don't tag with the full registry storage path docker push will default
# to docker hub and not store in the Google Artifact Registry
docker tag hello-world-enclave-sev us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev
docker push us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev

# update the image for a running instance
gcloud compute instances update-container instance-bearclave-sev-snp \
    --zone=us-central1-a \
    --project=bearclave \
    --container-image=us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev@sha256:d7214f758098275b254228345d46d2071b76b3c06aab5f198de1e58097370ba1

# Log into your instance and view output of your confidential workload
gcloud compute ssh instance-bearclave-sev-snp
docker logs -f $(docker ps -q)

# Start/Stop instances
gcloud compute instances start instance-bearclave-sev-snp \
    --zone=us-central1-a
gcloud compute instances stop instance-bearclave-sev-snp \
    --zone=us-central1-a
```

## Tutorial/Code links
- [confidential space tutorial](https://cloud.google.com/confidential-computing/confidential-space/docs/create-your-first-confidential-space-environment#run_the_workload)
confidential spaces use confidential VMs. This tutorial at least has some code
and shows deployment but its still unclear to me how communicating with programs
within the confidential VM works (if at all)
- [confidential-space repo](https://github.com/GoogleCloudPlatform/confidential-space)
another confidential spaces repo. may or may not be helpful
- [confidential space deploy workload doc](https://cloud.google.com/confidential-computing/confidential-space/docs/deploy-workloads)
- [confidential spaces images](https://cloud.google.com/confidential-computing/confidential-space/docs/confidential-space-images)
- [confidential space codelab 1](https://codelabs.developers.google.com/confidential-space-pki#0)
- [confidential space codelab 2](https://codelabs.developers.google.com/codelabs/confidential-space#0)
- [ncc confidential spaces security review](chrome-extension://efaidnbmnnnibpcajpcglclefindmkaj/https://www.nccgroup.com/media/edukzwst/_ncc_group_googleinc_e004374_confidentialspacereport_public_v10.pdf)
- [monitor and debug workloads](https://cloud.google.com/confidential-computing/confidential-space/docs/monitor-debug)
- [redirect to serial](https://cloud.google.com/confidential-computing/confidential-space/docs/deploy-workloads#tee-container-log-redirect)
- https://cloud.google.com/sdk/gcloud/reference/compute/instances/create-with-container
- https://cloud.google.com/compute/all-pricing?hl=en
- [third party confidential space project](https://github.com/salrashid123/confidential_space)
this might be the most thorough example yet. Go through code to figure out how
they are designing/building confidential spaces apps
