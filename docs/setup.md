# Overview

Follow the steps below to install and setup the tools necessary for
developing and running Bearclave applications. _Note that these steps, and the
Bearclave SDK, have only been tested on Ubuntu 24.04.3 LTS_

## Install & Setup (No TEE)

Bearclave provides a _No TEE_ mode that allows you to build and run TEE
applications on your local machine. Meaning you can develop, test, and debug
your applications without needing to have access to a TEE platform. While this
is not a true one-to-one replacement for a TEE platform, it can be useful for
speeding up your development cycle and reducing your cloud costs.

1. Install [Golang](https://golang.org/doc/install) version 1.24.3 or higher.
This is required to build and run Bearclave applications.
2. Install [Process Compose](https://github.com/F1bonacc1/process-compose)
version 1.78.0 or higher. While not strictly required, it is used in
the [Examples](https://github.com/tahardi/bearclave-examples) repository to run
and orchestrate applications in "No TEE" mode.

You now have the minimum set of tools required to build and run Bearclave
applications locally. If you wish to run applications on genuine TEE platforms, 
follow the AWS or GCP setup instructions below.

## Install & Setup (AWS)

TODO: [setup - aws](https://taylor-a-hardin.atlassian.net/browse/BCL-53)

## Install & Setup (GCP)

TODO: [setup - gcp](https://taylor-a-hardin.atlassian.net/browse/BCL-54)
