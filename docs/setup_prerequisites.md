7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup prerequisites for network next

Before you can setup network next, you need the following things:

* A OpenVPN instance
* A domain name
* A cloudflare account
* An AWS account with the (whatever tool AWS uses setup locally)
* A google account with the "gcloud" tool setup locally
* A linode account
* A semaphoreci startup account
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

You can finish the rest of the OpenVPN configuration by following this guide: https://www.linode.com/docs/products/tools/marketplace/guides/openvpn/

Once the OpenVPN server is up and running, setup your OpenVPN client so you can access the VPN. While you are on the VPN your IP address will appear as the VPN address. 

Later on, we're going to secure your network next instance such that the REST APIs, portal and relay are only accessible from your VPN IP address.

## 2. Register domain

Create a domain name using a domain name registrar, for example https://namecheap.com

This domain name will not be user facing, but will be visible to your organization when they go to the network next portal (eg. https://portal.[yourdomain])

##3. Setup a cloudflare account and import your new domains

We will use cloudflare to manage your domains. Create a new account at https://cloudflare.com and import your domain name.

This will take a few hours to a few days. Once the domains are managed by cloudflare you can proceed to the next step.


(more...)



#4. Increase google cloud quotas

When you create the dev environment, you'll likely be very close to the number of resources required to run it. This may stall out terraform, because it cannot allocate a VM because you are over quota.

To fix this, go to the "IAM & Admin" -> "Quotas" page in google cloud.

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
  
The requests above are aggressive and they will likely respond with lower numbers, accept these, then deploy the dev environment. Then, before you deploy the staging or production environments, request another aggressive quota increase. If they complain again, tell them you are doing a load test and these are the absolute numbers you need and this will likely get approved.

  
