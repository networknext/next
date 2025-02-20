<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Disable the raspberry clients

Open the file "terraform/dev/terraform.tfvars" and set "disable_raspberry" to true:

<img width="1363" alt="disable raspberry" src="https://github.com/user-attachments/assets/d861a1b2-84eb-4b88-b05f-86b9c480b652" />

```
git commit -am "disable raspberry"
git push origin
```

Tag the build to trigger a deploy:

```
git tag dev-003
git push origin dev-003
```

This process is how you deploy any code or configuration change to the dev backend.

1. Tag dev-[n+1]
2. Push the tag

The code at the tag is then automatically built and deployed to the dev environment.

While the deploy is running you can go to "Instance Groups" in google cloud, and watch the deploy -- notice that each instance template name with "002" is being replaced with "003":

	(screenshot showing deploy in progres....)

when the deploy completes, you'll see the various "raspberry_*" instance groups now have 0 VM instances each. They're disabled:

	(screenshot showing completed result)

Go to your portal and you see that the raspberry clients are no longer running.

	(screenshot)
