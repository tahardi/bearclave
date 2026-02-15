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
infrastructure to run Bearclave applications on AWS Nitro Enclaves.

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

TODO: [setup - gcp](https://taylor-a-hardin.atlassian.net/browse/BCL-54)
