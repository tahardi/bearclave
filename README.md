# Cloud-Based TEE Development

**Trusted Execution Environments (TEEs)** use specialized hardware and software to
provide stronger confidentiality and integrity guarantees than what is afforded
by traditional computing systems. For this reason, curious developers often
want to explore TEE technology for securing sensitive workloads, but find
themselves overwhelmed by steep learning curves and complicated requirements.
The _Bearclave_ project is a collection of code and documentation that aims to
address these challenges.

## What's Included?

- **Breakdowns** of TEE concepts and popular platforms, including AWS Nitro,
AMD SEV-SNP, and Intel TDX.
- **Guides** on building and deploying TEE-based applications to AWS and GCP.
- **Modules** for developing platform-agnostic Golang TEE applications.
- **Examples** demonstrating how to write, build, and deploy real-world
TEE-based applications.

## A Note on Costs

Running cloud-based TEE applications is not free. AWS and GCP TEE
compute instances typically cost between $0.20 to $0.50 per hour.
Fortunately, Bearclave provides a _No TEE_ mode that allows you to develop
and test applications locally. Using this mode, you should be able to
prototype and test TEE applications for just a few dollars a month.

## Getting Started

Use the links below to get started with the Bearclave project:
- [TEE Concepts]()
- [Install & Setup]()
- [Examples](github.com/tahardi/bearclave-examples)
