# Bearclave: TEE Development Made Easy

**Trusted Execution Environments (TEEs)** use specialized hardware and software to
provide stronger confidentiality and integrity guarantees than what is afforded
by traditional computing systems. For this reason, curious developers often
want to explore TEE technology for securing sensitive workloads, but find
themselves overwhelmed by steep learning curves and complicated requirements. 
The _Bearclave_ project is a collection of code and documentation that aims to
address these challenges.

## What's Included?

- **Breakdowns** of TEE concepts and platforms, including AWS Nitro Enclaves,
AMD SEV-SNP, and Intel TDX.
- **Guides** on building and deploying TEE-based applications to AWS and GCP.
- **Modules** for developing platform-agnostic Golang TEE applications.
- **Examples** demonstrating how to write, build, and deploy real-world
TEE-based applications.

> Running cloud-based TEE applications is not free. AWS and GCP TEE
compute instances typically cost between $0.20 to $0.50 per hour.
Fortunately, Bearclave provides a _NoTEE_ mode that allows you to develop
and test applications locally. By limiting your cloud usage with _NoTEE_ you
should be able to prototype and test TEE applications for just a few dollars
a month.

## Project Structure

- [**Bearclave**](https://github.com/tahardi/bearclave) an SDK for developing
TEE-based applications in Go.
- [**Bearclave Examples**](https://github.com/tahardi/bearclave-examples) a
collection of TEE-based applications demonstrating how to use the
Bearclave SDK.
- [**Bearclave TF**](https://github.com/tahardi/bearclave-tf) a collection of
Terraform modules for deploying Bearclave applications to AWS and GCP.
- [**Bearchain**](https://github.com/tahardi/bearchain) a (soon-to-be) collection
of TEE-related blockchain smart contracts.
- [**PluckMD**](https://github.com/tahardi/pluckmd) a handy tool for inserting
code into Markdown documents.

## Getting Started

1. Check out our documentation to learn about fundamental
TEE concepts ([Here](./docs/concepts.md)).
2. Get your environment [setup](./docs/setup.md) for developing with the
Bearclave SDK.
3. Try running our [Hello, World!](https://github.com/tahardi/bearclave-examples/hello-world)
example application.
