7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Create Google Cloud Projects with Terraform

Network Next runs across multiple projects in google cloud:

* Development
* Development Relays
* Staging
* Production
* Production Relays
* Storage

"Development" is a low cost development only backend used for testing. "Staging" for load testing. It has been tested to scale up to 1000 relays, 2.5M servers and 25M sessions (CCU). "Production" contains your production backend. Out of the box it can handle 1M CCU and can scale up to 25M+.

Google cloud relays for the development environment are created in the "Development Relays" project. Google cloud relays for production environment relays go in "Production Relays" project. Staging doesn't have any relays because it runs a simulation of relay and server load, so it's possible to load test the system for a lower cost than actually running 25 million clients connections.

All of these projects pull data and configuration from google cloud storage buckets in the "Storage" project. 

The configuration of all the permissions, service accounts and cloud storage buckets across all these projects is highly complex, so we have created a terraform script to do it all for you automatically.

_You need to perform this step only once._

At the console:

```console
cd ~/next/terraform/projects
terraform init
terraform apply
```

This step can take around 5-10 minutes to complete.

Once the projects are created, go to https://console.google.com and click on the project selector at the top-left of the screen:

![image](https://github.com/networknext/next/assets/696656/0ecc1ac6-f315-4348-95cc-63ee8669d25b)

You should now be able to see under the tab "ALL" that the projects have been created:

<img width="739" alt="image" src="https://github.com/networknext/next/assets/696656/888f0cd6-6d77-4372-b7ab-a2345d81bbeb">

Inside these projects service accounts have been created with appropriate permissions as needed by terraform to complete the per-project setup, and the google cloud storage buckets have been created under "Storage" project, with files uploaded to some of them.

The "Development", "Development Relays", "Staging", "Production" and "Production Relays" projects are at this point just a skeleton. No relays or backend services are running yet.

(next step)
