# Bearclave: Simplifying Cloud-Based TEE Development

Trusted Execution Environments (TEEs) are a fascinating fusion of specialized  
hardware and software designed to enhance the confidentiality and integrity of  
sensitive code and data. Curious developers often want to explore TEE technology  
but soon find themselves overwhelmed by steep learning curves and complicated  
requirements. The challenges include:

- Limited documentation that can be dense, scattered, or difficult to understand.
- The need for specialized hardware that is either costly or requires complex setup.
- A deep understanding of how TEEs function and interact with cloud platforms.

Bearclave is here to bridge that knowledge gap! This repository is tailored for  
developers—whether individuals or small teams—who want to take their first steps  
into the world of cloud-based TEE application development. Bearclave provides all  
the necessary resources to go from zero to a working example, while keeping the  
process approachable and affordable.

---

## What Bearclave Offers:
This repository offers a holistic, step-by-step guide to developing TEE  
applications, including:

- **Introduction to TEEs**: A high-level overview of TEEs and popular  
  implementations such as AWS Nitro, AMD SEV-SNP, and Intel TDX.
- **Cloud Integration Guides**: Detailed instructions for configuring the required  
  cloud resources on AWS and GCP.
- **Practical Code Examples**: Implementations for performing basic TEE operations  
  like remote attestation.
- **Deployment Simplified**: Code for compiling and deploying applications to cloud-  
  based TEEs.

Additionally, Bearclave includes a **"No TEE" development mode**, allowing you to  
develop and test your application without a TEE instance. This reduces costs  
significantly, making the barrier to entry even lower.

---

## A Note on Costs:
Building and deploying TEE applications typically requires specialized hardware,  
which isn't free. Unless you own and manage the hardware yourself, you'll need to  
rent resources through cloud providers like AWS or GCP. Fortunately, these  
providers offer affordable, TEE-enabled instances starting at $0.17 to $0.40 per  
hour. Paired with Bearclave's "No TEE" mode, you can develop and test your  
applications for just a few dollars a month if you carefully manage your instances.

---

## Important Reminder:
Bearclave is designed as an educational tool. While the repository provides  
practical examples and working code, it should not be considered production-ready.  
We encourage you to use it as a learning resource and adapt it for your unique  
production needs.

We hope Bearclave inspires you to explore the exciting world of Trusted Execution  
Environments and eases your journey into TEE-enabled cloud applications!

## Getting Started

### 1. Install Dependencies

Bearclave has been tested on **Ubuntu 24.04.1 LTS**. To get started, install the
following dependencies:

- **Golang (version 1.23.3 or higher):**  
  Install Go from the official website:  
  [https://go.dev/dl/](https://go.dev/dl/)

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
