# Overview

The collection of hardware, firmware, and software responsible for the security
of a system is often referred to as the **Trusted Computing Base (TCB)**.
Normally, this consists of your hypervisor, operating system, and computer.
The larger the TCB, the greater the potential for attackers to compromise your
system. Thus, one of the primary goals for any secure system is to reduce the size
of the TCB.

**Trusted Execution Environments (TEEs)** do this through a combination
of specialized hardware and software. They employ separate page tables,
special CPU instructions, customized firmware, and other such mechanisms to
isolate your code and data from the rest of the system. Instead of trusting the
entire OS and hypervisor, you need only trust the small subset of specialized
hardware and firmware that TEEs use to secure your code. 

The isolated space that TEEs create is known as an **Enclave**. Previous
implementations (e.g., Intel SGX) provided enclaves that isolated code at the
process level. This meant that you could protect a single program from the rest
of the system. There were a number of drawbacks to this approach. Most notably,
enclaves did not have access to system resources (e.g., network interfaces,
storage devices) without having to go through the untrusted OS or hypervisor.

TODO: [finish concepts overview](https://taylor-a-hardin.atlassian.net/browse/BCL-55)

## AWS Nitro Enclaves

TODO: [concepts - nitro](https://taylor-a-hardin.atlassian.net/browse/BCL-49)

## AMD Secure Encrypted Virtualization (SEV)

TODO: [concepts - sev](https://taylor-a-hardin.atlassian.net/browse/BCL-50)

## Intel Trusted Domain Extensions (TDX)

TODO: [concepts - tdx](https://taylor-a-hardin.atlassian.net/browse/BCL-51)
