# Overview

Follow the steps below to install and setup the tools necessary for
developing and running Bearclave applications. _Note that these steps, and the
Bearclave SDK, have only been tested on Ubuntu 24.04.3 LTS_

## Install & Setup (No TEE)

Bearclave provides a _No TEE_ mode that allows you to write, run, and test
your Bearclave applications without the need for a Trusted Execution Environment.
Meaning you can locally develop, test, and iterate on code before deploying it
to a TEE-enabled platform.

1. Install [Golang](https://golang.org/doc/install) version 1.24.3 or higher.
This is required to build and run Bearclave applications.
2. Install [Process Compose](https://github.com/F1bonacc1/process-compose)
version 1.78.0 or higher. While not strictly required, it is used in
the [Examples](https://github.com/tahardi/bearclave-examples) repository to run
and orchestrate applications in "No TEE" mode.

You now have the minimum set of tools required to build and run Bearclave
applications locally. If you wish to run Bearclave applications on genuine
TEE platforms, follow either of the guides below.

## Install & Setup (AWS)

TODO: [setup - aws](https://taylor-a-hardin.atlassian.net/browse/BCL-53)

## Install & Setup (GCP)

TODO: [setup - gcp](https://taylor-a-hardin.atlassian.net/browse/BCL-54)
