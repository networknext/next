<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Datacenter and relay naming conventions

A full production system has thousands of datacenters and hundreds of relays.

Let's keep them all organized with a clear naming convention :)

## 1. Datacenter naming convention

Each datacenter must be named:

* `[seller].[cityname].[number]`

For example:

* google.losangeles.1
* google.losangeles.2
* amazon.virginia.5

In cases where there is only one datacenter in a city for that seller, the number may be omitted.

* i3d.istanbul

When chosing a city name, prefer a cityname as already used for existing relays. This makes it easier to find other relays in the same city, which is a common operation when you are building your relay fleet.

# Relay naming convention

Relays must include the datacenter name, followed by a,b,c etc. for each relay in that datacenter.

* `[datacenter].[a|b|c...]`

For example:

* google.losangeles.2.a
* google.losangeles.2.b
* google.losangeles.2.c
* amazon.virginia.1.a
* amazon.virginia.1.b
* i3d.istanbul.a
* i3d.istanbul.b

In cases where there is only one relay in a datacenter, it is OK to omit the a,b,c if you wish, and then the relay is named the same as the datacenter:

* google.losangeles.1

[Back to main documentation](../README.md)
