# UNDER CONSTRUCTION

This repository was recently restructured, but the documentation is not yet
updated.
Note that examples have been moved to a separate repository:
[bearclave-examples](https://github.com/tahardi/bearclave-examples).

# Bearclave: Simplifying Cloud-Based TEE Development

Trusted Execution Environments (TEEs) are a combination of specialized  
hardware and software designed to enhance the confidentiality and integrity of  
sensitive code and data. Curious developers often want to explore TEE technology  
but find themselves overwhelmed by steep learning curves and complicated  
requirements. These challenges include:

- Limited documentation that can be dense, scattered, or difficult to understand
- The need for specialized hardware that is either costly or requires complex setup
- A deep understanding of how TEEs function and how to configure them on cloud platforms

Bearclave is here to bridge that knowledge gap! This repository is tailored for  
developers who want to take their first steps into the world of cloud-based TEE 
application development. Bearclave provides all the necessary resources to go 
from zero to a working example, while keeping the process approachable and affordable.

---

### What Bearclave Offers
This repository offers a holistic, step-by-step guide to developing TEE  
applications, including:

- **Introduction to TEEs**: A high-level overview of TEEs and popular  
  platforms such as AWS Nitro, AMD SEV-SNP, and Intel TDX
- **Cloud Integration Guides**: Detailed instructions for configuring the cloud  
  resources needed to build and deploy TEE applications on AWS and GCP
- **Practical Code Examples**: Demonstrations of how to compile, deploy, and
  interact with real-world applications on different TEE platforms
- **Platform-Agnostic Code**: A framework for developing applications on various
  TEE platforms. Bearclave abstracts away platform-specific details so you can
  more easily write portable applications without worrying about the 
  idiosyncrasies of the underlying TEE implementation

Additionally, Bearclave includes a **"No TEE" development mode**, allowing you to  
develop and test your application without a TEE instance. This reduces costs  
significantly, making the barrier to entry even lower.

---

### A Note on Costs
Building and deploying TEE applications typically requires specialized hardware,  
which isn't free. Unless you own and manage the hardware yourself, you'll need to  
rent resources through cloud providers like AWS or GCP. Fortunately, these  
providers offer affordable, TEE-enabled instances starting at $0.17 to $0.40 per  
hour. Paired with Bearclave's "No TEE" mode, you can develop and test your  
applications for just a few dollars a month if you carefully manage your resources.

---

### Important Reminder
Bearclave is designed as an educational tool. While the repository provides  
practical examples and working code, it should not be considered production-ready.  
We encourage you to use it as a learning resource and adapt it for your unique  
production needs.

We hope Bearclave inspires you to explore the exciting world of Trusted Execution  
Environments and eases your journey into TEE-enabled cloud applications!

---

# Getting Started
Bearclave has only been tested on **Ubuntu 24.04 LTS**. Modifications to the
example Makefiles and Dockerfiles may be required if you wish to build and
deploy from other systems.

---

## TEE Overview
Please refer to [this](docs/TEE.md) document for a high-level overview of TEEs
and popular implementations such as AWS Nitro, AMD SEV-SNP, and Intel TDX.

---

## Install Dependencies
Please ensure that all tools are properly installed and added to your system's
`PATH` for seamless use.

- **[Golang (version 1.24.3 or higher):](https://go.dev/dl/)** this project
  is written in Go and is required for building and running the examples.

- **[Docker Engine:](https://docs.docker.com/engine/install/)** cloud-based TEE
  platforms require you to provide your application packaged as an OCI-compliant
  image. While you could use any OCI-compliant tool, the build and deploy
  commands in the examples assume you have `docker`.

- **[Process Compose:](https://github.com/F1bonacc1/process-compose)** a simple
  tool modeled after `docker-compose` for initializing non-containerized
  applications. This is a convenience for running the "No TEE" versions of the
  example applications, but is not strictly necessary for running on cloud-based
  TEEs.

These are the minimum set of tools required to build and run the Bearclave
examples locally in "No TEE" mode. If you wish to run on real TEE platforms,
follow the steps detailed in [Configure Cloud Resources](#configure-cloud-resources).

---

## Configure Cloud Resources
If you wish to use Nitro Enclaves, refer to the [AWS setup guide](docs/AWS.md).
If you wish to use AMD SEV-SNP or Intel TDX, refer to the 
[GCP setup guide](docs/GCP.md).

---

## Build and Deploy Examples
Note that examples have been moved to a separate repository:
[bearclave-examples](https://github.com/tahardi/bearclave-examples).
