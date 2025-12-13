# Overview

Follow the steps below to install and setup the tools necessary for
developing and running Bearclave applications. _Note that these steps, and the
Bearclave SDK, have only been tested on Ubuntu 24.04.3 LTS_

## Install & Setup (No TEE)

Bearclave provides a _No TEE_ mode that allows you to build and run TEE
applications on your local machine. Meaning you can develop, test, and debug
your applications without needing to have access to a TEE platform. While this
is not a true one-to-one replacement, it can be useful for speeding up
development cycles and reducing cloud costs.

1. Install [Golang](https://golang.org/doc/install) (v1.24.3 or higher) to build
and run Bearclave applications.
2. Install [Process Compose](https://github.com/F1bonacc1/process-compose)
(v1.78.0 or higher) to orchestrate and run applications in "No TEE" mode.

You now have the minimum set of tools required to build and run Bearclave
applications locally. Try them out with one of the examples in our
[examples](https://github.com/tahardi/bearclave-examples) repository. If you
wish to run applications on genuine TEE platforms, continue on to the AWS or
GCP setup guides below.

## Install & Setup (AWS)

TODO: [setup - aws](https://taylor-a-hardin.atlassian.net/browse/BCL-53)

## Install & Setup (GCP)

TODO: [setup - gcp](https://taylor-a-hardin.atlassian.net/browse/BCL-54)
