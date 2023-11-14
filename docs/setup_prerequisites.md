7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup prerequisites for network next

Before you can setup network next, you need the following things:

* A OpenVPN instance
* A domain name
* A cloudflare account
* A google cloud account
* An AWS account
* A linode account
* A semaphoreci account
* Terraform installed on your development machine

Once these prerequisites are met, actually setting up your network next instance is easy as it is all performed with terraform scripts.

## 1. Setup a VPN on Linode

Create a new linode account at https://linode.com

Select "Marketplace" in the navigation menu on the left and search for "OpenVPN":

<img width="1548" alt="Screenshot 2023-08-07 at 12 15 48 PM" src="https://github.com/networknext/next/assets/696656/d89d006c-91f7-478f-923b-a0dc8ee6c0a8">

Select the OpenVPN item and fill out the configuration

<img width="1337" alt="Screenshot 2023-08-07 at 12 17 07 PM" src="https://github.com/networknext/next/assets/696656/34730500-68bc-4849-a270-12ef5122c1d5">

I recommend setting Ubuntu 20.04 as the base image and selecting a location nearest to you for lowest latency.

The "Dedicated 4GB" instance should be sufficient to run your VPN. It's around $40 USD per-month as of 2023.

Create the instance with label "openvpn".

Once the linode is created, take note of the IP address of your linode. You'll need it later:

<img width="1445" alt="Screenshot 2023-08-07 at 12 25 53 PM" src="https://github.com/networknext/next/assets/696656/9029c207-1b39-4ac4-8e7a-fc4566936d7b">

Follow this guide to finish the rest of the OpenVPN configuration: https://www.linode.com/docs/products/tools/marketplace/guides/openvpn/

Once the OpenVPN server is up and running, setup your OpenVPN client so you can access the VPN. 

I recommend OpenVPN Connect: http://openvpn.net/client/

Later on, we're going to secure your network next instance such that the REST APIs, portal and relay are only accessible from your VPN IP address.

## 2. Register domain

Create a domain name using a domain name registrar, for example https://namecheap.com

This domain name will not be player facing, but it will be visible to your organization when they go to the network next portal (eg. https://portal.[yourdomain])

## 3. Import your new domain into cloudflare

We will use cloudflare to manage your domain. Create a new account at https://cloudflare.com and import your domain name.

This will take a few hours to a few days. It's OK to work on other steps in this list in parallel.

## 4. Create a google cloud account

Go to https://cloud.google.com/ and sign up to create a new google account. It must have an organization associated with it.

Once the account is setup, click on this link and create a test project in the google cloud console:

[https://console.google.com](https://developers.google.com/workspace/guides/create-project)

<img width="587" alt="image" src="https://github.com/networknext/next/assets/696656/6cfdf331-3856-4e53-93d0-9b2b56902cbd">

## 5. Install "gcloud" utility to manage Google Cloud

Install the "gcloud" utility from here: https://cloud.google.com/sdk/docs/install

On the command line, initialize gcloud with your admin email address and test account to a configuration called "test":

`gcloud init`

Follow the prompts in the command line to create a new configuration called "test" that points at your "Test" project.

Verify that gcloud is properly configured by running the following command:

`gcloud compute regions list`

You should see a table of available google cloud regions printed to your console.

```console
gaffer@batman docs % gcloud compute regions list
NAME                     CPUS    DISKS_GB  ADDRESSES  RESERVED_ADDRESSES  STATUS  TURNDOWN_DATE
asia-east1               0/2400  0/204800  0/2300     0/700               UP
asia-east2               0/2400  0/204800  0/2300     0/700               UP
asia-northeast1          0/2400  0/204800  0/2300     0/700               UP
asia-northeast2          0/2400  0/204800  0/2300     0/700               UP
asia-northeast3          0/2400  0/204800  0/2300     0/700               UP
asia-south1              0/2400  0/204800  0/2300     0/700               UP
asia-south2              0/500   0/204800  0/575      0/575               UP
asia-southeast1          0/2400  0/204800  0/2300     0/700               UP
asia-southeast2          0/2400  0/204800  0/2300     0/700               UP
australia-southeast1     0/2400  0/204800  0/2300     0/700               UP
australia-southeast2     0/500   0/204800  0/575      0/575               UP
europe-central2          0/500   0/204800  0/575      0/575               UP
europe-north1            0/2400  0/204800  0/2300     0/700               UP
europe-southwest1        0/500   0/204800  0/575      0/575               UP
europe-west1             0/2400  0/204800  0/2300     0/700               UP
europe-west10            0/500   0/204800  0/575      0/575               UP
europe-west12            0/500   0/204800  0/575      0/575               UP
europe-west2             0/2400  0/204800  0/2300     0/700               UP
europe-west3             0/2400  0/204800  0/2300     0/700               UP
europe-west4             0/2400  0/204800  0/2300     0/700               UP
europe-west6             0/2400  0/204800  0/2300     0/700               UP
europe-west8             0/500   0/204800  0/575      0/575               UP
europe-west9             0/500   0/204800  0/575      0/575               UP
me-central1              0/500   0/204800  0/575      0/575               UP
me-central2              0/500   0/204800  0/575      0/575               UP
me-west1                 0/500   0/204800  0/575      0/575               UP
northamerica-northeast1  0/2400  0/204800  0/2300     0/700               UP
northamerica-northeast2  0/500   0/204800  0/575      0/575               UP
southamerica-east1       0/2400  0/204800  0/2300     0/700               UP
southamerica-west1       0/500   0/204800  0/575      0/575               UP
us-central1              0/2400  0/204800  0/2300     0/700               UP
us-east1                 0/2400  0/204800  0/2300     0/700               UP
us-east4                 0/2400  0/204800  0/2300     0/700               UP
us-east5                 0/500   0/204800  0/575      0/575               UP
us-south1                0/500   0/204800  0/575      0/575               UP
us-west1                 0/3000  0/204800  0/2300     0/700               UP
us-west2                 0/2400  0/204800  0/2300     0/700               UP
us-west3                 0/2400  0/204800  0/2300     0/700               UP
us-west4                 0/2400  0/204800  0/2300     0/700               UP
```

## 6. Request increase for google cloud quotas

By default your google cloud account will have very low limits on the resources you can use in Google Cloud.

To fix this, go to the "IAM & Admin" -> "Quotas" page in google cloud: https://console.cloud.google.com/iam-admin/quotas

<img width="1988" alt="image" src="https://github.com/networknext/next/assets/696656/cb20b82b-3768-47e5-af0a-dab609ed1657">

Request increases to the following quotas:

* In use IP addresses -> 256
* CPUs (US-Central) -> 1024
* VM Instances (US-Central) -> 1024
* CPUs (All Regions) -> 2048
* Target HTTP proxies -> 256
* Target URL maps -> 256
* Networks -> 64
* In use IP addresses global -> 256
* Static IP addreses global -> 256
* Health checks -> 256
* Regional managed instance groups (US-Central) -> 256
  
The requests above are moderately aggressive and they will likely respond with lower numbers, accept the lower numbers. This process will likely take several days to complete. 

## 7. Setup an AWS account

Setup a new AWS account at https://aws.amazon.com

Once the account is created, download and follow the instructions to setup the aws command line tool here: https://aws.amazon.com/cli/

Verify that the aws command line tool is setup correctly by running this command:

`aws ec2 describe-regions --all-regions`

You should see a json output to the console listing all AWS regions available to your account:

```console
gaffer@batman docs % aws ec2 describe-regions --all-regions
{
    "Regions": [
        {
            "Endpoint": "ec2.ap-south-2.amazonaws.com",
            "RegionName": "ap-south-2",
            "OptInStatus": "opted-in"
        },
        {
            "Endpoint": "ec2.ap-south-1.amazonaws.com",
            "RegionName": "ap-south-1",
            "OptInStatus": "opt-in-not-required"
        },
        {
            "Endpoint": "ec2.eu-south-1.amazonaws.com",
            "RegionName": "eu-south-1",
            "OptInStatus": "opted-in"
        },
        {
            "Endpoint": "ec2.eu-south-2.amazonaws.com",
            "RegionName": "eu-south-2",
            "OptInStatus": "opted-in"
        },
        {
            "Endpoint": "ec2.me-central-1.amazonaws.com",
            "RegionName": "me-central-1",
            "OptInStatus": "opted-in"
        },
        {
            "Endpoint": "ec2.il-central-1.amazonaws.com",
            "RegionName": "il-central-1",
            "OptInStatus": "not-opted-in"
        },
        {
            "Endpoint": "ec2.ca-central-1.amazonaws.com",
            "RegionName": "ca-central-1",
            "OptInStatus": "opt-in-not-required"
        },
        {
            "Endpoint": "ec2.eu-central-1.amazonaws.com",
            "RegionName": "eu-central-1",
            "OptInStatus": "opt-in-not-required"
        },
        (etc...)
```
## 8. Export linode API token

Log in to your linode account, then click on your user profile in the top right corner and select "API Tokens":

<img width="467" alt="image" src="https://github.com/networknext/next/assets/696656/67c66d2d-ffae-4833-b54c-c2c43c99f06d">

Click on the "Create a Personal Access Token" button and create a new token called "Terraform" with read/write access to all resources that never expires.

<img width="467" alt="image" src="https://github.com/networknext/next/assets/696656/04b0a9f4-5824-441f-8890-1b3b617fb70a">

Download this API token and save it under your home directory as: ~/secrets/terraform-akamai.txt

This API token will be used when terraform creates relays under linode.

## 9. Setup cloudflare API token

Log in to your cloudflare account and click on the user icon in the top right, and select "My Profile".

<img width="385" alt="image" src="https://github.com/networknext/next/assets/696656/71af13fb-c778-455c-a983-9006eeb8d3ff">

In the profile page, click on "API Tokens" in the left side menu:

<img width="279" alt="image" src="https://github.com/networknext/next/assets/696656/f66fdc97-f7ef-4cc9-8c36-9444945c30f9">

Click on the "Create Token" button:

<img width="1092" alt="image" src="https://github.com/networknext/next/assets/696656/1c86a3ca-241d-4e66-a201-c3835bb923c2">

Select the "Edit Zone DNS" template:

<img width="745" alt="image" src="https://github.com/networknext/next/assets/696656/832b916c-00c7-4c6b-b189-1400c29e817d">

Configure the token to apply to all zones, or if you want to be specific, point it at the zone corresponding to your domain name:

<img width="934" alt="image" src="https://github.com/networknext/next/assets/696656/3fa33270-cd87-472f-b561-e890d12a0120">

Create the token then copy the text and save the text to ~/secrets/terraform-cloudflare.txt

This API token will be used when your domain name is pointed to backend services running in google cloud.

## 10. Setup semaphore ci

Navigate to https://semaphoreci.com and sign up for a new account.

Click "Sign up for a free 15 day trial" and log in with your github account.

Click on the "Create new / Add new organization" button:

<img width="309" alt="image" src="https://github.com/networknext/next/assets/696656/c2451bc5-6770-40fa-ae20-56e34bddad41">

Finish the account setup and select the "Startup Plan".

Once your account is created, log in and select "Create new" from the top menu:

<img width="585" alt="image" src="https://github.com/networknext/next/assets/696656/7c41fb2a-fd28-49a5-88d0-eaef168c013c">

Create a new semaphore project pointing at your forked "next" repository.

Use the existing semaphoreci configuration in the repository.

## 11. Install terraform

Follow the instructions here to install terraform:

https://developer.hashicorp.com/terraform/install

## 12. Final checks

1. You have OpenVPN setup and you have recorded the VPN IP address somewhere
2. You have created a new domain and it is imported and managed by cloudflare
3. You have a working google account with "Test" project and 'gcloud' is working locally
4. You have requested increased quotas in google cloud and these quotas have been approved
5. You have a working AWS account and 'aws' is working locally
6. You have exported a linode API token to ~/secrets/terraform-akamai.txt
7. You have exported a cloudflare API token to ~/secrets/terraform-cloudflare.txt
8. You have backed up your secrets somewhere
9. You have setup a semaphoreci.com account
10. You have installed terraform on your dev machine

Once all these prerequisites are met, you can proceed to the next section: [docs/configure_network_next.md](Configure Network Next).
