7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup prerequisites for development environment

Before we can setup your dev environment, we need to setup some prerequites.

These include:

* A OpenVPN instance running on linode to secure your access to your network next backend
* Three domain names for different parts of your network next instance
* A cloudflare account so you can point "dev.[domain]" to your three development load balancer instances
* Configuration of your network next environment, so that it is secure with its own set of keypairs that are unique to you.

Once these prerequisites are created, the actual setup of your development backend is easy, as it is performed automatically with terraform scripts.

1. Setup a VPN on Linode

Create a new linode account at https://linode.com (they have been acquired by Akamai)

Select "Marketplace" in the navigation menu on the left:

<img width="190" alt="image" src="https://github.com/networknext/next/assets/696656/3a2000b9-1847-48bc-9ed1-95673379069a">

Search for "OpenVPN", and you'll find an image you can install with the OpenVPN server:



2. Create three domains

3. Setup a cloudflare account

4. Configure your network next environment
