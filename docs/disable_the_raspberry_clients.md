<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Disable the raspberry clients

The clients in the portal are test clients called "raspberry clients", because when we first developed Network Next, we ran hundreds of test clients around the world on raspberry pi's in people's homes. 

As we move forward, we don't want these fake clients running anymore. We only want _real clients_ in the portal.

To do this, open the file `terraform/dev/terraform.tfvars` and set "disable_raspberry" to true:

<img width="1363" alt="disable raspberry" src="https://github.com/user-attachments/assets/d861a1b2-84eb-4b88-b05f-86b9c480b652" />

Commit your changes to the git repository:

```console
git commit -am "disable raspberry"
git push origin
```

Then push a tag to trigger a deploy to dev:

```console
git tag dev-003
git push origin dev-003
```

This process is how you deploy any code or configuration change to the dev backend.

1. Commit the change and push to origin
2. Tag dev-[n+1]
3. Push the tag to origin

The code at the tag is then automatically built and deployed to the dev environment with terraform.

While the deploy is running you can go to "Instance Groups" in google cloud, and watch the deploy. 

<img width="946" alt="during deploy" src="https://github.com/user-attachments/assets/5e0a5b4f-7d72-44a1-86c1-11caeb83caa3" />

Notice that each instance template name with "002" is being replaced with "003".

When the deploy completes, you'll see the various "raspberry_*" instance groups now have 0 VM instances each.

The raspberry test server, clients and backend are now disabled.

<img width="946" alt="after deploy" src="https://github.com/user-attachments/assets/685d0e7c-9f0d-4419-be7d-5c51692c8fdc" />

Go to your portal and you see that the raspberry clients are no longer running.

<img width="1438" alt="no sessions" src="https://github.com/user-attachments/assets/1fb2f92e-1473-4de6-b413-9cec94f1dd79" />

Up next: [Connect a client to the test server](connect_a_client_to_the_test_server.md).
