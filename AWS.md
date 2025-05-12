# AWS Nitro Enclaves: Overview and Development Setup Guide

## What are AWS Nitro Enclaves?

AWS Nitro Enclaves is a feature of Amazon EC2 that helps create isolated,
highly secure environments for processing sensitive data. Examples include
personally identifiable information, intellectual property, and cryptographic
material. Nitro Enclaves is built on the AWS Nitro System, which provides
advanced security and performance capabilities.

Nitro Enclaves differ from other TEE platforms like Intel TDX and AMD SEV-SNP.
While those platforms are not tied to specific cloud providers, Nitro Enclaves
are proprietary to AWS and are only available on EC2 and EKS. They provide
specialized Linux VMs to run containerized applications in isolation, with no
network or storage access. Communication is solely via the Virtual Socket
(VSOCK) interface, allowing traffic to be proxied through the host.

---

## Setting Up Your Environment for Nitro Enclave Development

To start developing with Nitro Enclaves, you must launch and configure an EC2
instance that supports Nitro Enclaves. Note that THIS WILL COST MONEY. In fact,
developing on any cloud-based TEE platform is going to cost money. That said,
the smallest nitro-enabled instances only run around $0.17/hr. Coupled with the
Bearclave "unsafe" platform, you should be able to develop and test your enclave
application for $5–10/month as long as you are diligent about turning off your
instances when done.

---

### Step 1: Choosing and Launching an EC2 Instance
Follow these steps to launch a Nitro-enabled EC2 instance from the web console:
1. **Select a Nitro-enabled Instance**
   Visit the [Nitro-enabled instances guide](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html#instance-hypervisor-type)
   to select a supported instance type. Make sure the instance has enough vCPUs
   for the enclave and the host. The instance type I generally go for is:
   - `c5.xlarge`: 4vCPUs (~$0.17/hr).

2. **Configure the Instance**
   - Select **Amazon Linux** as the operating system.
   - Add an **SSH keypair** for secure access.
   - Be sure to enable **Nitro Enclaves** under advanced options.

3. **Configure SSH Access**
   After launching the instance, note the public IP and add it to your SSH
   configuration for quicker access.

#### Example SSH Configuration
```bash
Host ec2-nitro
    Hostname ec2-3-82-190-249.compute-1.amazonaws.com
    # AmazonLinux uses "ec2-user" and Ubuntu uses "ubuntu" as login usernames
    User ec2-user 
    IdentitiesOnly yes
    # Tell ssh which key to use when logging into the instance
    IdentityFile ~/.ssh/ec2-key--tahardi-bearclave.pem
    # These two will keep your ssh session alive even if you are not
    # active at your computer. Useful if you are running long tests
    ServerAliveInterval 300
    ServerAliveCountMax 2
```

---

### Step 2: Install Required Tools and Libraries

After logging into your EC2 instance, execute the following commands to install
Nitro CLI and other necessary tools.

1. **Install Git and Nitro CLI**
   ```bash
   sudo dnf install -y git aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel
   ```

2. **Grant Necessary Permissions**
   Add your user to the required groups for Nitro and Docker:
   ```bash
   sudo usermod -aG ne $USER
   sudo usermod -aG docker $USER
   ```
   Log out and log back in for these changes to apply.

3. **Enable Nitro Enclaves Allocator Service**
   Enable and start the Nitro Enclaves Allocator Service:
   ```bash
   sudo systemctl enable --now nitro-enclaves-allocator.service
   ```

4. **Enable Docker**
   Start Docker and configure it to run on instance startup:
   ```bash
   sudo systemctl enable --now docker
   ```

5. **Git Configuration for Private Repos (Optional)**
   If accessing private GitHub repositories, configure Git:
   ```bash
   git config --global url.https://<YOUR-TOKEN>@github.com/.insteadOf \
       https://github.com/
   git clone https://github.com/<YOUR-REPO>.git
   ```

6. **Install Go (Optional for Applications)**
   ```bash
   wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
   tar -xvf go1.23.3.linux-amd64.tar.gz
   sudo mv go /usr/local
   ```
   Add Go to your environment variables:
   ```bash
   echo 'export GOROOT=/usr/local/go' >> ~/.bash_profile
   echo 'export GOPATH=$HOME' >> ~/.bash_profile
   echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.bash_profile
   source ~/.bash_profile
   ```

---

### Step 3: Configure AWS Command Line Interface (CLI)

The AWS CLI simplifies EC2 instance management. Use Single Sign-On (SSO) to
authenticate securely with short-lived credentials.

1. **Install and Configure AWS CLI**
   Install and set up the CLI using the 
   [AWS CLI guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html).

2. **Setup SSO**
   Run the following command to configure SSO access:
   ```bash
   aws configure sso
   ```
   Input the following details:
   - SSO session name: `tahardi-dev-mac`
   - Start URL: `https://d-9a67642110.awsapps.com/start`
   - SSO Region: `us-east-2`
   - SSO Scopes: Default (`sso:account:access`)

3. **Define an AWS CLI Profile**
   During the setup, specify:
   - Default region: `us-east-2`
   - Output format: `json`
   - Profile name: `tahardi-ec2-mac`

4. **Authenticate with SSO**
   Sign in using:
   ```bash
   aws sso login --profile tahardi-ec2-mac
   export AWS_PROFILE=tahardi-ec2-mac
   ```

---

### Step 4: Managing Your EC2 Instances

Use these commands to start, describe, or stop your instances via AWS CLI.

1. **Start Instance**
   ```bash
   aws ec2 start-instances --instance-ids <INSTANCE-ID>
   ```

2. **Describe Instance Details**
   ```bash
   aws ec2 describe-instances --filters \
       "Name=tag:Name,Values=<INSTANCE-NAME>" \
       --query 'Reservations[*].Instances[*].{PublicIp:PublicIpAddress}' \
       --output json
   ```

3. **Stop Instance**
   ```bash
   aws ec2 stop-instances --instance-ids <INSTANCE-ID>
   ```

4. **Update Your SSH Configuration Automatically**
   If the EC2 public IP changes, update your SSH configuration as follows:

   **macOS:**
   ```bash
   NEW_IP=$(aws ec2 describe-instances --filters \
       "Name=tag:Name,Values=<INSTANCE-NAME>" \
       --query 'Reservations[*].Instances[*].PublicIpAddress' \
       --output text) && \
   sed -i '.bak' "s/^.*Hostname.*$/    Hostname ${NEW_IP}/" ~/.ssh/config
   ```

   **Linux:**
   ```bash
   NEW_IP=$(aws ec2 describe-instances --filters \
       "Name=tag:Name,Values=<INSTANCE-NAME>" \
       --query 'Reservations[*].Instances[*].PublicIpAddress' \
       --output text) && \
   sed -i.bak "s/^.*Hostname.*$/    Hostname ${NEW_IP}/" ~/.ssh/config
   ```

---

With this setup, you're ready to build and deploy applications with AWS Nitro
Enclaves. Learn more about advanced enclave configuration in the [Nitro
Enclaves documentation](https://docs.aws.amazon.com/enclaves/latest/user/nitro-enclaves.html).
