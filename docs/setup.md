# Overview

Follow the steps below to install and setup the tools necessary for
developing and running Bearclave applications.

> Note that these steps, and the Bearclave SDK, have only been tested on
> Ubuntu 24.04.3 LTS

## Install & Setup (No TEE)

Bearclave provides a _No TEE_ mode that allows you to build and run TEE
applications on your local machine. Meaning you can develop, test, and debug
your applications without needing to have access to a TEE platform. While this
is not a true one-to-one replacement, it can be useful for speeding up
development cycles and reducing cloud costs.

1. Install [Golang](https://golang.org/doc/install) (v1.24.3 or higher) to build
and run Bearclave applications.
2. Install [Process Compose](https://github.com/F1bonacc1/process-compose)
(v1.78.0 or higher) to orchestrate and run applications in "No TEE" mode.

You now have the minimum set of tools required to build and run Bearclave
applications locally. Try them out with one of the examples in our
[examples](https://github.com/tahardi/bearclave-examples) repository. If you
wish to run applications on genuine TEE platforms, continue on to the AWS or
GCP setup guides below.

## Install & Setup (AWS)

Amazon Web Services (AWS) has its own proprietary TEE platform known as AWS
Nitro Enclaves. Follow the steps below to set up the necessary tools and
infrastructure to develop Bearclave applications on AWS Nitro Enclaves.

1. Create an [AWS Account](https://aws.amazon.com/). Note that this is your
"root" account and should only be used to configure Billing and IAM roles.
2. Install and configure the [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).
The Makefile targets in the [Bearclave Examples](https://github.com/tahardi/bearclave-examples)
repository assume that the AWS CLI is installed and configured to use a role
with sufficient permissions to manage EC2 instances (see the Makefiles for details).
3. Install the [Terraform CLI](https://developer.hashicorp.com/terraform/install)
version 1.14.3 or higher.
4. Clone the [Bearclave TF](https://github.com/tahardi/bearclave-tf) repository.
5. Use the `aws-nitro-enclaves/` module to create an AWS Nitro Enclaves enabled
EC2 instance with the necessary dependencies to run Bearclave applications.
    ```bash
    git clone https://github.com/tahardi/bearclave-tf.git
    cd bearclave-tf/modules/aws-nitro-enclaves/
    terraform init
    terraform plan
    terraform apply
    ```
6. Follow the steps in the [Bearclave Examples](https://github.com/tahardi/bearclave-examples)
repository to build and run an example application on the newly provisioned
instance.

## Install & Setup (GCP)

Google Cloud Platform (GCP) provides compute instances that support the
AMD SEV-SNP and Intel TDX TEE platforms. Follow the steps below to set up the
necessary tools and infrastructure to develop Bearclave applications on
AMD SEV-SNP and Intel TDX.

1. Create a [Google Account](https://accounts.google.com/). If you use GMail,
Drive, or any other similar Google service, then you already have an account;
feel free to use that account. At the time of this writing, Google offers
$300 in free credits to new users.
2. Install and configure the [GCP CLI](https://docs.cloud.google.com/sdk/docs/install-sdk).
The Makefile targets in the [Bearclave Examples](https://github.com/tahardi/bearclave-examples)
repository assume that the GCP CLI is installed and configured to use a role
with sufficient permissions to manage Compute VMs and the Artifact Registry 
(see the Makefiles for details).
3. Install [Docker Engine](https://docs.docker.com/engine/install/) to containerize
and deploy Bearclave applications.
4. Create an image repository in the Artifact Registry and authorize docker to
push images to it. Note that some of these values may be out-of-date. Check the
[Bearclave Examples](https://github.com/tahardi/bearclave-examples) repository
to see what values are used in the Makefile for pushing images.
    ```bash
    gcloud artifacts repositories create bearclave \
    --repository-format=docker \
    --location=us-east1 \
    --description="Docker repository for bearclave images" \
    --project=bearclave
    gcloud auth configure-docker us-east1-docker.pkg.dev
    ```
5. Install the [Terraform CLI](https://developer.hashicorp.com/terraform/install)
version 1.14.3 or higher.
6. Clone the [Bearclave TF](https://github.com/tahardi/bearclave-tf) repository.
7. Use the `gcp-sev-snp/` and `gcp-tdx/` modules to create AMD SEV-SNP and
Intel TDX enabled compute instances with the necessary dependencies to run
Bearclave applications.
    ```bash
    git clone https://github.com/tahardi/bearclave-tf.git
    cd bearclave-tf/modules/gcp-sev-snp/
    terraform init
    terraform plan
    terraform apply
    
    cd ../gcp-tdx/
    terraform init
    terraform plan
    terraform apply
    ```
8. Follow the steps in the [Bearclave Examples](https://github.com/tahardi/bearclave-examples)
   repository to build and run an example application on the newly provisioned
   instances.
