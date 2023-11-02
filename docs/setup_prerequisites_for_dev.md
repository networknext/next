7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup prerequisites for development environment

Before we can setup your dev environment, we need to setup some prerequites.

These include:

* A OpenVPN instance running on linode to secure access to your network next backend
* Three domain names for different parts of your network next instance
* A cloudflare account so you can point "dev.[domain]" to three different components in your dev backend.
* Increased quotes so your google cloud account can create enough resources to run the environment.

Once these prerequisites are created, the actual setup of your development backend is easy, as it is performed with terraform scripts.

1. Setup a VPN on Linode

Create a new linode account at https://linode.com (they have been acquired by Akamai)

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

Later on, we're going to secure the dev environment such that the REST APIs and portal are only accessible from your VPN IP address.

2. Create three domains

Create three silly domain names of your choice using a domain name registrar, for example https://namecheap.com

Each of these domains will be used for different components of your network next backend instance.

For example:

* losangelesfreewaysatnight.com -> relay backend load balancer
* spacecats.net -> server backend load balancer
* virtualgo.net -> REST API and portal

Try to make sure that registration for each of these domains is anonymized and protected. Ideally, an attacker wouldn't be able to discover the full set of domains or link them together in any way.

3. Setup a cloudflare account and import your new domains

We will use cloudflare to manage your domains. Create a new account at https://cloudflare.com and import your domain names.

This will take a few hours to a few days. Once the domains are managed by cloudflare you can proceed to the next step.

Next step: [setup your dev environment](setup_dev_environment.md)

4. Increase google cloud quotas

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

  
