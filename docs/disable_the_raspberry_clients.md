<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Disable the raspberry clients

Open the file "terraform/dev/terraform.tfvars" and set "disable_raspberry" to true:

	(image of config highlighting raspberry disable true)

Commit your changes to the git repository:

git commit -am "disable raspberry"
git push origin

Tag the build to trigger a deploy:

	git tag dev-003
	git push origin dev-003

This process is how you deploy any code or configuration change to the dev backend.

Tag dev-[n+1]

Push the tag

The code at the tag is automatically built and deployed to the dev environment.

While the deploy is running you can go to "Instance Groups" in google cloud, and watch deploy -- notice that each instance template name with "002" is being replaced with "003"

	(screenshot showing deploy in progres....)

and when the deploy completes, you'll see the various "raspberry_*" instance groups now have 0 VM instances each.

	(screenshot showing completed result)

Go to the portal and you see that the raspberry clients are no longer running.

	(screenshot)

Up next: [Run your own client and server](run_your_own_client_and_server.md).
