7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup prerequisites for development environment

Before we can setup your dev environment, we need to setup some prerequites.

These include:

* A OpenVPN instance running on linode to secure access to your network next backend
* Three domain names for different parts of your network next instance
* A cloudflare account so you can point "dev.[domain]" to your three development load balancer instances
* Configuration of your network next environment, so that it is secure with its own set of keypairs that are unique to your company.

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

Once the linode is created, take note of the IP address of your linode. You'll need it later.

You can finish configuration by following these instructions: https://www.linode.com/docs/products/tools/marketplace/guides/openvpn/

_Basically, you just need to connect to the instance via the linode browser based SSH, trigger some setup, and then from this point forward you can configure the rest of the OpenVPN server using the OpenVPN admin interface running on your linode..._

Once the OpenVPN server is up and running, setup your OpenVPN client so you can access the VPN. While you are on the VPN your IP address will appear as the VPN address. Later on, we're going to secure the dev environment such that the REST APIs and portal are only accessible from your VPN IP address.

2. Create three domains

3. Setup a cloudflare account

4. Configure your network next environment
