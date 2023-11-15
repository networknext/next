7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Create Google Cloud Projects with Terraform

Network Next runs across multiple projects in google cloud.

There is a "Development" project that runs the development backend, a "Staging" project for load testing, and a "Production" project for your production environment.

Relays for the development backend are created in the "Development Relays" project. The production environment relays go in "Production Relays". Staging project doesn't have any relays, and runs simulated servers and relays to make load testing more efficient.

All of these projects pull data and configuration from google cloud storage buckets in the "Storage" project. The configuration of all the permissions, service accounts and cloud storage buckets across all these projects is highly complex, so we have created a terraform script to do it all for you automatically.

You need to perform this step only once.

At the console, run this:

```console
cd ~/next/terraform/projects
terraform init
terraform apply
```

This step can take around 5-10 minutes to complete.

Once the projects are created, go to https://console.google.com and click on the project selector at the top-left of the screen:

![image](https://github.com/networknext/next/assets/696656/0ecc1ac6-f315-4348-95cc-63ee8669d25b)

You should now be able to see under the tab "ALL" the projects have been created in google cloud:

<img width="739" alt="image" src="https://github.com/networknext/next/assets/696656/888f0cd6-6d77-4372-b7ab-a2345d81bbeb">

Inside these projects service accounts have been created with appropriate permissions as needed by terraform to complete the per-project setup, and the google cloud storage buckets have been created under "Storage" project, with some files uploaded to them.

The "Development", "Staging" and "Production" projects are at this point just a skeleton. No backend services are running yet.

(next step)
