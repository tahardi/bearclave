# Bearclave

- write your own implementations of `hf/nsm` and `hf/nitrite`
  - consider rewriting these as it looks like they are no
  longer supported

# Configuring an EC2 Instance for Nitro Development
Check out [this](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html#instance-hypervisor-type)
page for Nitro-enabled instances. Be careful to select an instance with enough 
vCPUs if you require more than 1 for your enclave or wish to run more than one
at once. At the time of this writing, I go with c5.xlarge because it has 4vCPUs
and costs around $0.17/hr.

Create an instance with
- AmazonLinux as the OS
- an ssh keypair for logging in
- "enable enclave" otherwise you won't be able to start enclaves

Grab it's public IP and add to your `~/.ssh/config` file for ease-of-login
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

Once logged into your EC2 instance you may need to install some tools and libs.
Check [this](https://docs.aws.amazon.com/enclaves/latest/user/nitro-enclave-cli-install.html)
page for nitro-related config and setup you need to do. An example setup is below.
Note that I created a fine-grained access token in github with content read
permissions and explicitly chose the repos I wanted to grant access to.
```bash
# Install git and nitro-cli tooling
sudo dnf install -y git aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel

# Add your user to the nitro enclave and docker groups (exit and login after)
sudo usermod -aG ne $USER
sudo usermod -aG docker $USER

# Start the nitro-enclave resource allocator service
sudo systemctl enable --now nitro-enclaves-allocator.service

# Start docker and tell it to start everytime the instance starts
sudo systemctl enable --now docker

# Install nix for running nix-shell
sh <(curl -L https://nixos.org/nix/install) --no-daemon
. /home/ec2-user/.nix-profile/etc/profile.d/nix.sh

# Configure git so you can clone private repos
git config --global url.https://<your-personal-access-token-here>@github.com/.insteadOf https://github.com/
git clone https://github.com/tahardi/bearclave.git

# Setup Go
wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
tar -xvf go.1.23.3.linux-amd64.tar.gz
sudo mv go /usr/local

# Add to .bash_profile (exit and logout or source .bash_profile)
export GOROOT=/usr/local/go
export GOPATH=$HOME
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

# AWS Notes
[setup aws cli](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html)
Configure aws cli to use your short-lived SSO Identity Center profile instead
of long-lived IAM creds and/or logging in directly under aws account

1. Run `aws configure sso` and provided it the following
```bash
# SSO session name
tahardi-dev-mac 
# SSO start URL
https://d-9a67642110.awsapps.com/start
# SSO region
us-east-2
# SSO registration scopes (this is the default)
sso:account:access
```
2. I then chose my "SystemAdministrator" role and it asked me some questions
about configuring a profile for logging into said role
```bash
# default client region
us-east-2
# cli default output format
json
# profile name
tahardi-ec2-mac 
```
3. You can edit these settings in `~/.aws.config`. To use this profile with
the aws cli specify `--profile tahardi-ec2-mac`
4. You may have to login first with
```bash
# sign into a profile. Caches creds and auto renews as needed
aws sso login --profile tahardi-ec2-mac 
# sign out
aws sso logout
# consider setting this so you don't have to constantly specify the profile flag
export AWS_PROFILE=tahardi-ec2-mac
```
5. EC2 commands
```bash
# start an instance - this is my tahard-bearclave instance ID
# specify a profile otherwise it tries to use what default
# aws is configured for
export AWS_PROFILE=tahardi-ec2-mac
aws sso login --profile tahardi-ec2-mac

export TAHARDI_BEARCLAVE_EC2_ID=i-01bdf23ce28366cb5
aws ec2 start-instances --instance-ids $TAHARDI_BEARCLAVE_EC2_ID

aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=tahardi-bearclave" \
    --query 'Reservations[*].Instances[*].{InstanceId: InstanceId, InstanceType: InstanceType, State: State.Name, PublicIp: PublicIpAddress, Name: Tags[?Key==`Name`].Value|[0]}' \
    --output json

aws ec2 stop-instances --instance-ids $TAHARDI_BEARCLAVE_EC2_ID
aws sso logout
```

6. Extract IP from running ec2 instance and update ssh config (on mac)
```bash
NEW_IP=$(aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=tahardi-bearclave" \
    --query 'Reservations[*].Instances[*].{PublicIp: PublicIpAddress}' \
    --output json | jq -r '.[0][0].PublicIp') && \
sed -i '.bak' -E "s/(Host ec2-bearclave[[:space:]]*$'\n'[[:space:]]*Hostname[[:space:]]*)[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/\1$NEW_IP/" ~/.ssh/config 
```

6. Extract IP from running ec2 instance and update ssh config (on linux)
```bash
# Get the IP and update ssh config
NEW_IP=$(aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=tahardi-bearclave" \
    --query 'Reservations[*].Instances[*].{PublicIp: PublicIpAddress}' \
    --output json | jq -r '.[0][0].PublicIp') && \
sed -i.bak -E "s/(Host ec2-bearclave\n[[:space:]]*Hostname[[:space:]]*)[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/\1$NEW_IP/" ~/.ssh/config
```

# GCP Notes
Below is a very rough draft of notes, tips, links compiled during my initial GCP exploration. I plan to clean them
up in future PRs as I continue to refine the sev and tdx implementations.

## TODOs
1. Figure out how to see enclave logging output---some sort of google confidential VM debug mode?
2. clean up these notes
3. Figure out how to allow for http calls to ec2 instance so i can call proxy
4. Figure out how to actually set up google cloud IAM and other things correctly at some point...

## Tutorial/Code links
[confidential space tutorial](https://cloud.google.com/confidential-computing/confidential-space/docs/create-your-first-confidential-space-environment#run_the_workload)
confidential spaces use confidential VMs. This tutorial at least has some code and shows deployment but its still
unclear to me how communicating with programs within the confidential VM works (if at all)

I *think* google has a VM instance and you just specify a docker image to pull and run in said instance. So, just like
how Nitro has their custom Linux VM for running "EIFs" google has their own custom VM though they do say you can
[define and run your own VM image](https://cloud.google.com/confidential-computing/confidential-vm/docs/create-custom-confidential-vm-images)
  - when creating the VM instance there is an option under "OS and Storage -> Container" that allows you to deploy
an image with the VM

[stet repo](https://github.com/GoogleCloudPlatform/stet)
related to confidential spaces. Has some code that may help

[confidential-space repo](https://github.com/GoogleCloudPlatform/confidential-space)
another confidential spaces repo. may or may not be helpful

[confidential space deploy workload doc](https://cloud.google.com/confidential-computing/confidential-space/docs/deploy-workloads)
[confidential spaces images](https://cloud.google.com/confidential-computing/confidential-space/docs/confidential-space-images)

[third party confidential space project](https://github.com/salrashid123/confidential_space)
this might be the most thorough example yet. Go through code to figure out how they are designing/building
confidential spaces apps

[confidential space codelab 1](https://codelabs.developers.google.com/confidential-space-pki#0)

[confidential space codelab 2](https://codelabs.developers.google.com/codelabs/confidential-space#0)

[ncc confidential spaces security review](chrome-extension://efaidnbmnnnibpcajpcglclefindmkaj/https://www.nccgroup.com/media/edukzwst/_ncc_group_googleinc_e004374_confidentialspacereport_public_v10.pdf)

# Confidential VM Setup & Deployment
- setup gcloud cli tool: currently using this in conjunction with the web console though I may ultimately move to
just the cli tool

- You can add a personal ssh key to the VM instance. Note that when you deploy your container as a confidential VM it
only allows you to log into the guest confidential VM in read-only mode---you cannot ssh into the host instance! For
this reason I don't find ssh'ing all that helpful right now
  - The default username is not "ubuntu". Mine was "taylor.antonio.hardin" I think it pulled from the email in my ssh
    pub key file that i uploaded

- Enable "Artifact Registry API" for storing your docker images
- enable http in network settings (also added 8080 port to the default http firewall rules)

Commands for pushing your image to google's Artifact Registry so you can pull them into your VM

```bash
gcloud init
gcloud auth login
gcloud config set account `403490793521-compute@developer.gserviceaccount.com`
gcloud auth configure-docker us-east1-docker.pkg.dev
#
gcloud artifacts repositories add-iam-policy-binding bearclave \
  --location=us-east1 \
  --member=serviceAccount:403490793521-compute@developer.gserviceaccount.com \
  --role=roles/artifactregistry.writer

# Why do I have to tag it this way? Push fails if I dont...
docker tag hello-world-enclave-sev us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev

# Is the tag what tells it where to push or is that bc I logged into the registry earlier?
docker push us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev
```

```bash
# Maybe how you create an AMD_SEV_SNP instance? According to AI assistant you can't actually
# create an sev-snp-enabled instance via the web GUI
gcloud compute instances create-with-container instance-bearclave-sev-snp \
    --zone=us-central1-a \
    --machine-type=n2d-standard-2 \
    --confidential-compute-type=SEV_SNP \
    --maintenance-policy=TERMINATE \
    --container-image=us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev@sha256:a824652361384b513af405e25c8f4c5c258d193c56b249ed70391befc0e2b43f \
    --project=bearclave \
    --tags=http-server \
    --scopes=cloud-platform \
    --shielded-secure-boot

# update the image for a running instance
gcloud compute instances update-container instance-bearclave-sev-snp \
    --zone=us-central1-a \
    --project=bearclave \
    --container-image=us-east1-docker.pkg.dev/bearclave/bearclave/hello-world-enclave-sev@sha256:cca4019c8909b92202de15d7b8d20b70129a231be02dbf8e72bd54a876e1725f

# God fucking dammit you have to mount `/dev/sev-guest` to the guest VM since that is not
# done by default... Christ almight some documentation would be very helpful
# Update your instance create command to do this
gcloud compute instances update-container instance-bearclave-sev-snp \
    --container-mount-host-path mount-path=/dev/sev-guest,host-path=/dev/sev-guest \
    --container-privileged \
    --zone=us-central1-a \
    --project=bearclave

gcloud compute instances create vm1 \
  --zone=us-central1-a \
  --confidential-compute \
  --shielded-secure-boot \
  --tags=tee-vm \
  --project $OPERATOR_PROJECT_ID \
  --maintenance-policy=TERMINATE \
  --scopes=cloud-platform  \
  --image-project=confidential-space-images \
  --image=confidential-space-231201 \
  --network=teenetwork \
  --service-account=operator-svc-account@$OPERATOR_PROJECT_ID.iam.gserviceaccount.com \
  --metadata ^~^tee-image-reference=$IMAGE_HASH~tee-restart-policy=Never~tee-container-log-redirect=true~tee-signed-image-repos=us-central1-docker.pkg.dev/$BUILDER_PROJECT_ID/repo1/tee

```


```golang
	//if len(userdata) > AMD_SEV_USERDATA_SIZE {
	//	return nil, fmt.Errorf(
	//		"userdata must be less than %d bytes",
	//		AMD_SEV_USERDATA_SIZE,
	//	)
	//}
	//attestation := []byte("in sev-snp attester: ")
	//attestation = append(attestation, userdata...)
	//return attestation, nil

	//sevQP, err := client.GetQuoteProvider()
	//if err != nil {
	//	return nil, fmt.Errorf("getting quote provider: %w", err)
	//}
	//
	//if !sevQP.IsSupported() {
	//	return nil, fmt.Errorf("SEV is not supported")
	//}
	//
	//var reportData [64]byte
	//copy(reportData[:], userdata)
	//attestation, err := sevQP.GetRawQuote(reportData)
	//if err != nil {
	//	return nil, fmt.Errorf("getting quote: %w", err)
	//}
	//return attestation, nil
```