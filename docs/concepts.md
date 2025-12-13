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

Previous implementations (e.g., Intel SGX) isolated code at the _process_ level.
Meaning, you could protect a single program from the rest of the system.
Modern TEEs (e.g., AWS Nitro Enclaves, AMD SEV-SNP, Intel TDX), however,
provide isolation at the _VM_ level. So, you can protect an entire guest OS
and its applications from the rest of the system.

Not only do TEEs provide isolation, they also provide a mechanism for proving
the integrity and authenticity of isolated code to outside parties. This
process is known as **remote attestation**. TEEs generate a report containing
platform information (e.g., CPU vendor, firmware version) and a measurement of
the isolated code. This report is cryptographically signed with a private
key available only to the TEE. The public component is made widely available
so that anyone can verify the signature and information contained within the
report.

While TEEs provide stronger integrity and confidentiality guarantees than
traditional systems, they do so at the cost of performance and usability.
TEEs encrypt and authenticate data that crosses the CPU boundary to ensure
that only trusted code can access it. This means that I/O intensive workloads
may see slowdowns of 5 to 20%. Additionally, certain system resources are not
available to code running in a TEE. For instance, TEEs cannot directly access
network interfaces or storage devices. To do so, they must go through the
untrusted OS and hypervisor. This may add to the complexity of the code, as
precautions must be taken to ensure any requests passed to the untrusted OS
are properly sanitized or otherwise protected.

In summary, TEEs provide stronger security guarantees than traditional systems
at the cost of increased complexity and performance. They _can_ isolate code and
data from privileged software in a way that is verifiable and auditable. They
_cannot_ protect against physical attacks, nor guarantee access to
system resources managed by the untrusted OS and hypervisor.

## AWS Nitro Enclaves

TODO: [concepts - nitro](https://taylor-a-hardin.atlassian.net/browse/BCL-49)

## AMD Secure Encrypted Virtualization (SEV)

TODO: [concepts - sev](https://taylor-a-hardin.atlassian.net/browse/BCL-50)

## Intel Trusted Domain Extensions (TDX)

TODO: [concepts - tdx](https://taylor-a-hardin.atlassian.net/browse/BCL-51)
