# Overview

A **Trusted Computing Base (TCB)** is the collection of hardware, firmware,
and software that secures a computing system. Normally, this might consist of
your hypervisor, operating system, and computer hardware. The larger the TCB,
the larger the likelihood that it contains bugs or vulnerabilities that
attackers can use to subvert system security. Thus, one of the primary
goals in secure system design is to reduce the size of the TCB.

**Trusted Execution Environments (TEEs)** reduce the TCB through a
combination of specialized hardware and software. They employ separate page
tables, special CPU instructions, customized firmware, and other such
mechanisms to isolate your code and data from the rest of the system. Instead
of trusting the entire OS, hypervisor, and hardware platform, you need only
trust the subset of specialized hardware and firmware used to realize TEEs.
Not only is your code protected from other applications, it is protected from
privileged software (e.g., OS, hypervisor) as well.

Previous TEE implementations (e.g., Intel SGX) isolated code at the _process_
level. Meaning, you could protect a single program from the rest of the system.
Modern TEEs (e.g., AWS Nitro Enclaves, AMD SEV-SNP, Intel TDX), however,
provide isolation at the _VM_ level, allowing you to protect an entire guest OS
and its applications from the untrusted OS and hypervisor.

Not only do TEEs provide isolation (i.e., confidentiality), they also provide a
mechanism for verifying the integrity and authenticity of the code executing
within. This process is known as **remote attestation**. When a remote client
connects to an application supposedly running within a TEE, the client can
request an attestation report. This report contains platform information
(e.g., CPU vendor, firmware version) and a measurement of the code currently
executing within the TEE. More importantly, this report is signed with a key
that is only available within the TEE context. By verifying the signature
and the report information, a client can determine whether they are
connected to the expected application running within a genuine TEE.

While TEEs provide stronger integrity and confidentiality guarantees than
traditional systems, they do so at the cost of performance and usability.
TEEs encrypt and authenticate data that crosses the CPU package boundary to
ensure that only trusted code can access it. This means that I/O intensive
workloads may see slowdowns of 5 to 20%. Additionally, certain system resources
are not available to code running in a TEE. For instance, TEEs cannot directly
access network interfaces or storage devices. To do so, they must go through
the untrusted OS and hypervisor. This may add to the complexity of the code, as
precautions must be taken to ensure any requests passed to the untrusted OS
are properly sanitized or otherwise protected.

In summary, TEEs provide stronger confidentiality and integrity guarantees than
traditional systems do, but at the cost of increased complexity and performance.
They isolate code and data from privileged software in a way that can be verified
and audited. They cannot protect against physical attacks, however, nor can they
guarantee access to system resources managed by the untrusted OS and hypervisor.

## AWS Nitro Enclaves

Amazon Web Services (AWS) provides a TEE platform called
[Nitro Enclaves](https://aws.amazon.com/ec2/nitro/nitro-enclaves/), which allows
you to create isolated, hardened, and constrained VMs on EC2 compute instances.
The Nitro Hypervisor ensures that the VM is protected from the Host OS and
applications, as well as from users with root and admin privileges. There are
several key differences between Nitro Enclaves and the other major TEE platforms.

The first difference is in the trust model. Users of Nitro Enclaves must place
complete trust in the cloud provider (i.e., Amazon). The hardware and software
that makes up Nitro Enclaves is designed, built, and maintained solely by Amazon.
While the high-level design and architecture of the system has been made
publicly available, no outside third parties have ever been granted access to
independently verify the security claims made by Amazon regarding Nitro
Enclaves.

The second difference is in the networking model. Applications running within
Nitro Enclaves do not have access to sockets. Instead, their only connection
to the "outside world" is through a _virtual_ socket interface managed by the
Nitro hypervisor. This means that traditional networking applications (e.g.,
HTTP servers) must be modified to send and receive data via virtual sockets.
Additionally, you need to run a Proxy on the (untrusted) EC2 Host instance to
forward communications between remote clients and the Nitro Enclave. **This is
the reason that Bearclave requires at least two applications:** a Proxy for
handling network traffic and the Enclave for running the actual application
logic.

The third difference relates to key management and persistent storage.
Applications running within Nitro Enclaves only have direct access to RAM,
which does not persist between Enclave instances. Enclaves can write
data to disk, or to some networked storage device, but they must do so via
untrusted channels (e.g., Host instance OS). While they can use standard
authenticated-encryption mechanisms to preserve the confidentiality and
integrity of their data, they still need a method for persisting the
encryption keys so future Enclave instances can decrypt and verify the data.
Platforms like AMD SEV-SNP and Intel TDX offer the ability to derive
cryptographic keys that are tied to a particular platform (i.e., CPU) or
Enclave (i.e., application). This means that applications can derive the
same keys across invocations so long as they are running on the same
hardware instance and with the same application code. On the other hand,
Nitro Enclaves must use the AWS Key Management Service (KMS) if they wish
to persist keys. This is another example of where the trust models differ
between Nitro Enclaves and other platforms.

The last difference relates to remote attestation. Nitro Enclave attestations
attest to the boot state of the Enclave VM (e.g., kernel image, cmd line args)
_and_ the state of the applications running in userspace (e.g., your
Enclave server). By default, other platforms such as AMD SEV-SNP and Intel TDX
only attest to the boot state of the protected VM. Meaning, you can verify that
the application is running within a VM on SEV-SNP or TDX, but you cannot
discern _what_ that application is. To include application measurements in
attestations, you need to configure the VM to run a virtual Trusted Platform
Module (vTPM).

In summary, Nitro Enclaves are fully isolated and hardened virtual machines.
They have no persistent storage, no interactive access, and no external networking.
Assuming you trust Amazon, you are assured that the Nitro Hypervisor protects
the CPU and memory of your enclave from users, applications, and privileged
software on the parent EC2 instance. The integrity and authenticity of your
applications can be verified by you, or by any third party, via Nitro
attestation reports.

## AMD Secure Encrypted Virtualization (SEV)

Beginning with the EPYC 7001 line of server CPUs, AMD has included support for
[Secure Encrypted Virtualization (SEV)](https://www.amd.com/en/developer/sev.html).
There have been several iterations of the SEV platform, including:
- Encrypted State (ES)
- Secure Nested Paging (SNP)
- Trusted I/O (TIO)
- Transparent Secure Memory Encryption (TSME)

This document will focus on SEV-SNP, as that is the particular flavor currently
supported by Bearclave.

Like the other major TEE platforms (i.e., AWS Nitro, Intel TDX), AMD SEV-SNP
lets you secure a VM from the host OS and hypervisor. It encrypts the VM's CPU
register and RAM contents using keys that are unique to each VM instance. These
keys are generated and managed by the AMD Secure Processor (ASP). The ASP is a
dedicated ARM-based microcontroller integrated into the AMD x86 CPUs. It
operates separately from the host OS and hypervisor, and is responsible for
the secure boot, attestation reports, and key management.

The ASP can derive keys tied to a specific CPU or VM instance. Depending on
the application's needs, these keys can be ephemeral---meaning they are
available during the VM's lifetime but do not persist between VM restarts---or
they can be persistent, allowing VMs to encrypt and store data to disk.

Unlike Nitro Enclaves, AMD SEV-SNP provides VMs with access to network sockets
so that standard Linux socket APIs and network stacks work without modification.
Since the hypervisor cannot read the VM's encrypted memory pages, the VM must
copy over the network buffer to a shared unprotected memory region. The
performance impact on modern applications such as NGINX is
[around 7%](https://www.amd.com/content/dam/amd/en/documents/epyc-business-docs/performance-briefs/confidential-computing-performance-sev-snp-google-n2d-instances.pdf)
compared to non-confidential VMs.

While you must trust AMD to properly design and build the SEV-SNP hardware and
firmware, you do not necessarily have to trust a cloud provider as well. Since
SEV-SNP-enabled CPUs are commercially available, you can buy and manage your
own SEV-SNP TEE machine if you so desire. That said, they are also available on
major cloud providers such as AWS and GCP. It is important to note that none of
the TEE implementations discussed in this document protect from actors with
physical access. So, you must trust the owner and operator of the TEE hardware,
whether that be you or some third party.

The ASP provides attestation reports that cover the VM's boot state and
configuration (e.g., kernel code, cmdline arguments, init program), but not
the userspace applications (i.e., your application code). If you want to include
application measurements in your AMD SEV-SNP attestation reports, then you need
to configure the VM to run a vTPM. This program is generally run at the highest
privilege level within the VM and "extends" Platform Configuration Registers
(PCRs) to create immutable records of your application's code.

In summary, AMD SEV-SNP allows you to isolate VMs from the host OS and
hypervisor. The AMD Secure Processor provides CPU and VM-specific keys
that can be used to persist data across VM restarts. The AMD SEV-SNP system
can be purchased and managed privately, or through cloud providers like AWS
and GCP. Application-level attestations are not provided by default but can
be achieved through the use of a virtual Trusted Platform Module.

## Intel Trusted Domain Extensions (TDX)

TODO: [concepts - tdx](https://taylor-a-hardin.atlassian.net/browse/BCL-51)
