<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Enable acceleration to Sao Paulo

Each buyer has a set of datacenters that are enabled for acceleration.

By default the test buyer is setup so that the following datacenters are accelerated:

* google.iowa.1
* google.iowa.2
* google.iowa.3
* google.iowa.6

In this section we're going to enable acceleration for these datacenters as well:

* google.saopaulo.1
* google.saopaulo.2
* google.saopaulo.3

First, open the file `terraform/dev/relays/terraform.tfvars` and make the following change:

<img width="851" alt="add sao paulo datacenters for test buyer" src="https://github.com/user-attachments/assets/83634f32-4b90-42d9-ab00-5db3cbc18874" />

Commit the change:

```console
git commit -am "add sao paulo datacenters for test buyer"
git push origin
```

Apply the change via terraform:

```console
cd ~/next/terraform/dev/relays
terraform apply
```

Commit the database to make the changes active in the backend:

```console
cd ~/next
next database
next relays
```

Verify the change by running:

```console
next database
```

And you should now see that the test buyer is active for both "ohio" and "saopaulo" google datacenters:

<img width="422" alt="destination datacenters" src="https://github.com/user-attachments/assets/b85f883f-84e9-4a86-8c56-349716661c9f" />

Next, we will spin up some google relays in Sao Paulo.

Up next: [Modify set of google relays](modify_set_of_google_relays.md).
