# Bearclave

Bearclave simplifies the process of developing, deploying, and managing
applications that run within Trusted Execution Environments (TEEs) such as AWS
Nitro Enclaves, AMD SEV-SNP, and Intel TDX. This repository provides a
comprehensive set of tools, interfaces, and examples to help developers fully
leverage modern TEE platforms.

---

## What Are Trusted Execution Environments (TEEs)?

Trusted Execution Environments (TEEs) represent a significant advancement in
secure computing. Unlike conventional computing environments running on a
standard operating system and hardware, TEEs combine specialized software and
hardware to deliver enhanced confidentiality and integrity for data and code.

TEE key features include:
- **Isolation:** TEEs isolate a virtual machine (VM) from both userland
  applications and the host operating system (OS). This ensures that neither
  can read or modify the VM's data and code.
- **Hardware-backed security:** Modern TEEs combine hardware support (e.g.,
  instruction-set architecture changes or separate chips) with specialized
  software (e.g., hypervisors) to establish a secure computing environment.
- **Popular platforms:** Examples include AWS Nitro Enclaves, AMD SEV-SNP,
  and Intel TDX.

While programs running in TEEs benefit from unparalleled security guarantees
such as code integrity and data confidentiality, they face limitations in
typical functionalities like networking and persistent storage. Bearclave
bridges this gap by expanding the capabilities of TEE-based applications.

---

## What Is Bearclave?

Bearclave is a set of tools and modules designed to help developers build and 
deploy TEE-based Golang applications. Here's what it
offers:

### Key Features:
1. **Low-level Interfaces:**
  - Bearclave provides interfaces for core TEE functionalities such as
    attestation, verification, and secure local communication.
  - Implementations are included for supported TEE platforms: AWS Nitro, AMD
    SEV-SNP, and Intel TDX.

2. **High-level SDK:**
  - The repository builds on the low-level interfaces to provide an SDK that
    adds higher-level functionality, including networking, storage, and
    deployment utilities. This makes it easier to build practical applications
    for TEEs.

3. **Deployment Examples:**
  - Examples and tools demonstrate the full lifecycle of TEE applications,
    including:
    - Registering and configuring cloud resources.
    - Packaging applications for deployment.
    - Debugging and testing applications running within TEEs.

4. **Platform-Specific Guides:**
  - Detailed documentation to help developers configure and deploy on various
    TEE platforms, including AWS and GCP.

---

## Why Bearclave?

1. **Holistic Approach:** Bearclave not only showcases how to write secure
   TEE-compatible programs but also provides a complete deployment and
   debugging workflow for real-world usage scenarios.
2. **Web-Scale Compatibility:** The repository equips developers to build
   practical TEE applications that overcome networking and storage limitations
   of these environments.
3. **Focus on Usability:** Bearclave reduces the complexity of adopting TEEs by
   exposing intuitive, platform-agnostic abstractions for low-level tasks and
   operational workflows.

---

## Getting Started

### 1. Install Dependencies

Bearclave has been tested on **Ubuntu 24.04.1 LTS**. To get started, install the
following dependencies:

- **Golang (version 1.23.3 or higher):**  
  Install Go from the official website:  
  [https://go.dev/dl/](https://go.dev/dl/)

- **golangci-lint:**  
  Install golangci-lint for running code linting checks:  
  [https://golangci-lint.run/usage/install/](https://golangci-lint.run/usage/install/)

- **Docker:**  
  Install Docker for managing containerized applications:  
  [https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/)

- **process-compose:**  
  Install process-compose for managing process groups via YAML files:  
  [https://github.com/F1bonacc1/process-compose](https://github.com/F1bonacc1/process-compose)

Please ensure that all tools are properly installed and added to your system's
`PATH` for seamless setup.

### 2. Clone the Repository

Clone the Bearclave repository and navigate to its directory:
```bash
git clone https://github.com/tahardi/bearclave.git && cd bearclave
```

### 3. Build and Deploy Examples

Follow the deployment examples in the `Makefile` for your platform of choice
(e.g., AWS or GCP). Refer to the [Platform-Specific Guides](#additional-resources)
for details on platform-specific configurations.

---

## Additional Resources

For platform-specific details and examples, refer to the following:
- **[Amazon Web Services (AWS)](AWS.md):** Guide for deploying enclaves on AWS
  Nitro Enclaves.
- **[Google Cloud Platform (GCP)](GCP.md):** Notes and insights on deploying
  SEV-SNP and TDX workloads on GCP.

---

Bearclave is an ongoing project dedicated to making modern secure computing
environments practical and accessible to all developers. Feedback and
contributions are always welcome!
