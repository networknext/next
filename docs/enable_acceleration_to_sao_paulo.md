<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Enable acceleration to Sao Paulo

Each buyer has a set of datacenters that are enabled for acceleration.

By default the test buyer is setup so that the following datacenters are accelerated:

# google.iowa.1
# google.iowa.2
# google.iowa.3

In this section we're going to enable acceleration for the google Sao Paulo datacenters:

# google.saopaulo.1
# google.saopaulo.2
# google.saopaulo.3

To do this, open the file "terraform/dev/relays/terraform.tfvars", and make the following change:

	(screenshot showing edit)

Commit the change:

	git commit -am "add sao paulo datacenters for test buyer"
	git push origin

Apply the change to the postgres database via terraform:

	cd terraform/dev/relays
	terraform apply

And then commit the database to make the changes active in the backend:

	cd ~/next
	next database
	next relays

You can verify this by running:

	next database

And you should now see that the test buyer is active for both "ohio" and "saopaulo" google datacenters:

	(screenshot)

Next, we will to spin up some google relays.

Up next: [Disable the raspberry clients](disable_the_raspberry_clients.md).
