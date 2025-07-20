# Amazon Web Services (AWS) Setup Guide
Amazon Web Services (AWS) provides compute instances that support the
AMD SEV-SNP and AWS Nitro TEE platforms, but Bearclave currently only supports
AWS Nitro. Follow the steps below to sign up for an AWS account and configure
the cloud resources required to develop on AWS Nitro Enclaves.

---

### Configure AWS Cloud
1. **Create an [AWS Account](https://aws.amazon.com/)** Note that this will be
   your "root" account and should only be used to setup billing and your user(s).
   I suggest looking for tutorials on AWS IAM Identity Center. Personally, I
   attached Billing and SystemAdministrator policies to my "developer" user.
   This allows me to login under a "role" with limited permissions instead of 
   as root. Note that your SSO page for logging in under these roles can be
   found in the "Settings Summary" on your IAM Identity Center page and should
   look something like `https://<subdomain>.awsapps.com/start`

2. **Setup Billing** You will need a valid billing method to deploy the
   TEE-enabled EC2 instances.

---

### Install and Configure the AWS CLI Tool
The Makefile targets in `examples/` require the `aws` cli tool for starting,
stopping, and sshing into your EC2 instances.

1. **Install the [`aws` CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)**

2. **Setup SSO** Configure SSO so `aws` can login under your SystemAdministrator
   role. Again, you can find your SSO login URL on IAM Identity Center under
   "Settings Summary".
   ```bash
   aws configure sso
   ```
   Input the following details:
   - SSO session name: `bearclave`
   - Start URL: `https://<subdomain>.awsapps.com/start`
   - SSO Region: `us-east-2`
   - SSO Scopes: Default (`sso:account:access`)

3. **Define an AWS CLI Profile**
   During the setup, specify:
   - Default region: `us-east-2`
   - Output format: `json`
   - Profile name: `personal`

4. **Authenticate with SSO**
   Sign in using:
   ```bash
   # personal is the profile named used in `examples/`. If you use a different
   # profile name you will need to update the Makefile targets to work
   aws sso login --profile personal
   ```

---

### Configure an EC2 Instance for Nitro Enclave development
The AWS Nitro developer workflow differs from AMD and Intel on GCP in that we
build and deploy the enclave program on the EC2 instance (as opposed to doing
it locally). Follow these steps to launch a Nitro-enabled EC2 instance from the
web console and install necessary tools for deploying Nitro applications:

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

4. **SSH into your EC2 instance**
   ```bash
   ssh ec2-nitro
   ```

5. **Install Git and Nitro CLI**
   ```bash
   sudo dnf install -y git aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel
   ```

6. **Grant Necessary Permissions**
   Add your user to the required groups for Nitro and Docker:
   ```bash
   sudo usermod -aG ne $USER
   sudo usermod -aG docker $USER
   ```
   Log out and log back in for these changes to apply.

7. **Enable Nitro Enclaves Allocator Service**
   Enable and start the Nitro Enclaves Allocator Service:
   ```bash
   sudo systemctl enable --now nitro-enclaves-allocator.service
   ```

8. **Enable Docker**
   Start Docker and configure it to run on instance startup:
   ```bash
   sudo systemctl enable --now docker
   ```

9. **Git Configuration for Private Repos (Optional)**
   If accessing private GitHub repositories, configure Git:
   ```bash
   git config --global url.https://<YOUR-TOKEN>@github.com/.insteadOf \
       https://github.com/
   git clone https://github.com/<YOUR-REPO>.git
   ```

10. **Install Go (Optional for Applications)** 
    ```bash
    wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
    tar -xvf go1.24.3.linux-amd64.tar.gz
    sudo mv go /usr/local
    ```
    Add Go to your environment variables:
    ```bash
    echo 'export GOROOT=/usr/local/go' >> ~/.bash_profile
    echo 'export GOPATH=$HOME' >> ~/.bash_profile
    echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.bash_profile
    source ~/.bash_profile
    ```
