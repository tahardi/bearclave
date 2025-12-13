# Simplifying Cloud-Based TEE Development

**Trusted Execution Environments (TEEs)** use specialized hardware and software to
provide stronger confidentiality and integrity guarantees than what is afforded
by traditional computing systems. For this reason, curious developers often
want to explore TEE technology for securing sensitive workloads, but find
themselves overwhelmed by steep learning curves and complicated requirements.
The **Bearclave** project is a collection of code and documentation that aims to
address those challenges.

## What's Included?

**Introduction to TEEs** An overview of TEEs and popular platforms, including
AWS Nitro, AMD SEV-SNP, and Intel TDX  
**How-to Guides** Steps for building and deploying applications to TEE
platforms on cloud providers like AWS and GCP  
**SDK** A library that allows you to develop platform-agnostic Golang
applications without having to worry about the underlying TEE implementation  
**Examples** Practical code examples that demonstrate how to build, deploy,
and interact with real-world applications on different TEE platforms

---

## A Note on Costs

Running cloud-based TEE applications is not free. AWS and GCP TEE-capable
compute instances typically cost between $0.20 to $0.50 per hour.
Fortunately, Bearclave provides a _No TEE_ mode that allows you to develop
and test applications locally. Using this mode, you should be able to
prototype and test TEE applications for just a few dollars a month.

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
