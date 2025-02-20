<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Spin up other relays in Sao Paulo

Acceleration comes from having many different networks and paths to get from the client to the server.

To support this, we want to have around 10-20 different relays providers in each city, from as many different hosting companies as possible.

This creates a massive diversity in different networks, which is how Network Next is able to find acceleration for you.

Let's see what other sellers have datacenters in Sao Paulo:

```console
next datacenters saopaulo
```

You'll see:

<img width="747" alt="datacenters in sao paulo" src="https://github.com/user-attachments/assets/f29a1426-1035-4d14-a083-cd3f5ea60554" />

I recommend that you set up at least the following bare metal providers in Sao Paulo, in addition to the google, amazon and akamai relays already set up.

* datapacket.saopaulo
* gcore.saopaulo
* i3d.saopaulo
* latitude.saopaulo.1
* latitude.saopaulo.2
* phoenixnap.saopaulo
* serversdotcom.saopaulo
* zenlayer.saopaulo 

The sellers above are _manual_, meaning there is no terraform provider support automatically creating and destroying relays for you. You have to do this manually, and then link the relays to terraform by IP address.

As an example, let's work through this process for latitude.saopaulo.1 and latitude.saopaulo.2

First, sign in to https://latitude.sh and setup an SSH key with the public key set to the contents of ~/secrets/next_ssh.pub and call this SSH key "Network Next".

Then, create two hourly bare metals machines in sao paulo, and sao paulo 2 datacenters respectively, making sure to set them as Ubuntu, and for them to use the "Network Next" SSH key, and hourly billing.

Next, open up terraform/dev/relays/main.tf and change the latitude section to:

```terraform
# ===============
# LATITUDE RELAYS
# ===============

locals {

  latitude_relays = {

    "latitude.saopaulo.1" = {
      datacenter_name = "latitude.saopaulo.1"
      public_address  = "189.1.174.29"
    },

    "latitude.saopaulo.2" = {
      datacenter_name = "latitude.saopaulo.2"
      public_address  = "103.88.232.173"
    },

  }
}
```

Where you replace the IP addresses with the public IPs of the two latitude.sh bare metal machines you created.

Check in your work:

```console
git commit -am "add latitude relays in sao paulo"
git push origin
```

And then apply your changes to the database via terraform:

```console
cd terraform/dev/relays
terraform apply
```

And commit your changes to the backend:

```console
cd ~/next
next database
next commit
```

Run next relays and you should see the latitude relays added to the list, but they are currently offline:

<img width="1228" alt="next relays latitude offline" src="https://github.com/user-attachments/assets/e5809908-944c-485e-a17d-695612f7e99b" />

Now you can ssh in to the latitude relays:

```console
next ssh latitude.saopaulo.1
```

Inside the SSH session, copy and paste the contents of the file "scripts/init_relay.sh" to the console and press enter.

Do the same thing for latitude.saopaulo.2 relay.

Next, connect to the VPN and run:

```console
next setup latitude
```

Wait a few minutes, and you should see the latitude relays are now online:

<img width="1228" alt="next relays latitude online" src="https://github.com/user-attachments/assets/2204c446-58eb-4b55-98af-1c3659a415f8" />

This is the general process for setting up manual relays in Network Next.

1. Create the relay manually on the provider website or portal
2. Do whatever you need to do to make it so the relay can be SSH'd into with ~/secrets/next_ssh.pub key
3. SSH into the relay via `next ssh [relayname]` and run the scripts/init_relay.sh manually
4. Connect to the VPN
5. Run `next setup [relayname]` locally
6. Verify the new relay is up by running `next relays`
7. Wait 5 minutes before the relay is ready to carry traffic

ps. If you cannot SSH in for step #3, it is possible your provider is using a different username for SSH login by default than root. 

To fix this, edit the file: "terraform/[sellername]/main.tf" and adjust the `ssh_user` variable.

Here is how I made that change for latitude already:
   
<img width="879" alt="latitude ssh user" src="https://github.com/user-attachments/assets/25933124-7948-4c70-8fc0-71c77d05aab9" />

If you need to do this for a supplier, please let me know so I can make the change in the root Network Next git repository for you.

Up next: [Spin up relays near you](spin_up_relays_near_you.md).
