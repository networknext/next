<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Operator guide to bare metal relays

This section describes how to use the Network Next terraform provider to configure bare metal relays.

Typically, bare metal means anything that's not cloud. But in this context, bare metal means any relays that you have manually provisioned without using terraform, and you can now SSH in and you want to turn them into relays in your relay fleet.

## 1. Create a new seller module in terraform

First we need to set up a module for the new seller in terraform.

Datapacket.com (https://datapacket.com) is an excellent bare metal provider for relays. But they don't have a terraform provider so you have to configure them manually.

Let's start by configuring a seller in terraform called "datapacket".

1. Create a new folder under `~/next/sellers` called "datapacket".
2. Copy the template `main.tf` from `~/next/sellers/bare_metal/main.tf` into your `sellers/datapacket` directory.

Open the copy of `main.tf` for editing.

You'll see:

```
# ----------------------------------------------------------------------------------------

variable "relays" { type = map(map(string)) }

locals {

  seller_name = "[Your seller]"

  seller_code = "[seller]"

  ssh_user = "root"

  datacenter_map = {

    "<seller>.cityname" = {
      latitude    = 10.00
      longitude   = 20.00
    }

  }

}

output "relays" {
  description = "All relays for [seller]"
  value = {
    for k, v in var.relays : k => zipmap( 
      [
        "relay_name", 
        "datacenter_name",
        "seller_name",
        "seller_code",
        "public_address", 
        "internal_address", 
        "internal_group", 
        "ssh_address", 
        "ssh_user",
      ], 
      [
        k,
        v.datacenter_name,
        local.seller_name,
        local.seller_code,
        v.public_address, 
        "", 
        0, 
        v.public_address, 
        local.ssh_user,
      ]
    )
  }
}

output "datacenters" {
  description = "All datacenters for [seller]"
  value = locals.datacenter_map
}

# --------------------------------------------------------------------------
```

Replace "[seller]" with "datapacket", and "[Your seller]" with "Datapacket" and save the file.

## 2. Define seller datacenters in terraform

Modify the datacenter map in `sellers/datapacket/main.tf` to add datacenters you want to use.

Go to https://datapacket.com and see what datacenters they have. At the time of writing, they have the following datacenters:

* stockholm
* copenhagen
* dublin
* london
* amsterdam
* warsaw
* kyiv
* london
* frankfurt
* brussels
* zurich
* prague
* bratislava
* zurich
* paris
* vienna
* bucharest
* milan
* zagreb
* sofia
* lisbon
* madrid
* marseille
* palermo
* athens
* istanbul
* telaviv
* johanesburg
* sydney
* singapore
* hongkong
* tokyo
* queretaro
* bogota
* saopaulo
* santiago
* vancouver
* seattle
* denver
* sanjose
* losangeles
* chicago
* toronto
* boston
* newyork
* losangeles
* dallas
* houston
* mcallen
* miami
* boston
* newyork
* ashburn
* atlanta

The first thing we need to do is add all these datacenters to a datacenter map, along with their latitude and logitude, so they become available to create relays in.

Go through each datacenter and add an entry in the datacenter map. You'll need to look up the approximate lat/long of each city. Take special care with signs on the lat/long values.

```
  datacenter_map = {

    "datapacket.losangeles" = {
      latitude    = 34.0522
      longitude   = -118.2437
    },

    "datapacket.chicago" = {
      latitude    = 41.8781
      longitude   = -87.6298
    },

  }
```

Once all the datacenters are added to `datapacket/main.tf` save the file.

## 3. Link seller to dev/prod relays terraform script

...

## 4. Provision relay

...

## 6. Link relay to terraform

...

## 7. Setup relay

...

## 8. Commit database and verify relay is online

...

[Back to main documentation](../README.md)
