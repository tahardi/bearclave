# Cloud-Based TEE Development Made Easy

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

## Project Structure

```text
.
├── bearclave/          # Low-level primitives for TEE features (attestation, networking, timing)
│   ├── docs/           # Documentation on TEE concepts and platforms
│   ├── internal/       # Platform-specific implementations of low-level primitives
│   ├── mocks/          # Mocks for testing
│   ├── modfiles/       # Mod files for go tools (e.g., golangci-lint) used in the project
│   ├── tee/            # Platform-agnostic abstractions for enclave HTTP clients and servers
│   └── ...
└── ...
```

## A Note on Costs

Running cloud-based TEE applications is not free. AWS and GCP TEE
compute instances typically cost between $0.20 to $0.50 per hour.
Fortunately, Bearclave provides a _NoTEE_ mode that allows you to develop
and test applications locally. By limiting your cloud usage with _NoTEE_ you
should be able to prototype and test TEE applications for just a few dollars
a month.

## Getting Started

- [TEE Concepts](./docs/concepts.md)
- [Install & Setup](./docs/setup.md)
- [Examples](https://github.com/tahardi/bearclave-examples)
