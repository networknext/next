<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Modify set of google relays

Open the file `terraform/dev/relays/main.tf` and change the locals section underneath "GOOGLE RELAYS" to match this:

```terraform
# =============
# GOOGLE RELAYS
# =============

locals {

  google_credentials = "~/secrets/terraform-dev-relays.json"
  google_project     = file("~/secrets/dev-relays-project-id.txt")
  google_relays = {

    "google.saopaulo.1" = {
      datacenter_name = "google.saopaulo.1"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.saopaulo.2" = {
      datacenter_name = "google.saopaulo.2"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.saopaulo.3" = {
      datacenter_name = "google.saopaulo.3"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.dallas.1" = {
      datacenter_name = "google.dallas.1"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.dallas.2" = {
      datacenter_name = "google.dallas.2"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.dallas.3" = {
      datacenter_name = "google.dallas.3"
      type            = "n2-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.1" = {
      datacenter_name = "google.virginia.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.2" = {
      datacenter_name = "google.virginia.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.virginia.3" = {
      datacenter_name = "google.virginia.3"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.southcarolina.2" = {
      datacenter_name = "google.southcarolina.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.southcarolina.3" = {
      datacenter_name = "google.southcarolina.3"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.southcarolina.4" = {
      datacenter_name = "google.southcarolina.4"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },
  }
}
```

This changes the set of google cloud relays from:

* google.iowa.1.a
* google.iowa.1.b
* google.iowa.1.c
* google.iowa.2
* google.iowa.3
* google.iowa.6
* google.ohio.1
* google.ohio.2
* google.ohio.3
* google.virginia.1
* google.virginia.2
* google.virginia.3

To:

* google.saopaulo.1
* google.saopaulo.2
* google.saopaulo.3
* google.dallas.1
* google.dallas.2
* google.dallas.3
* google.virginia.1
* google.virginia.2
* google.virginia.3
* google.southcarolina.2
* google.southcarolina.3
* google.southcarolina.4

The good news is that terraform takes care of all changes for you, including removing old google cloud VMs and adding new ones. It's all automatic. 

First check in your work:

```console
git commit -am "change google relays"
git push origin
```

Then run terraform:

```console
cd terraform/dev/relays
terraform apply
```

Commit the database to the backend to make it active:

```console
cd ~/next
next database
next commit
```

Wait a few minutes for the relays to initialize...

Connect to the VPN, then setup the new google relays:

```console
next setup google
```

After about 5 minutes the new relays should be online:

```console
next relays
```

<img width="1208" alt="relays online" src="https://github.com/user-attachments/assets/a06ef47e-99cd-4952-98a3-2261e388d6f4" />

Now we can SSH into the test server again:

<img width="1482" alt="ssh into test server" src="https://github.com/user-attachments/assets/dde2a877-f1a3-4967-a212-0ad5049f86fe" />

Then run:

```console
sudo systemctl restart app && sudo journalctl -fu app
```

and you'll see that the test server now finds a server relay in Sao Paulo:

<img width="1245" alt="server finds relays" src="https://github.com/user-attachments/assets/703be26e-9a34-420f-aa37-550cc247ee97" />

This is great news. Acceleration is now _possible_.

But to actually find this acceleration, we need multiple different networks to optimize across for clients connecting to google.saopaulo.1

This means we need to spin up as many different relays in Sao Paulo that we can find from different hosting companies!

Up next: [Modify set of amazon relays](modify_set_of_amazon_relays.md).
