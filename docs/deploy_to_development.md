<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Deploy to Development

## 1. Deploy the development backend

Network Next uses tag based deployments to trigger backend system deploys from branches.

To deploy to dev, first you merge your code into the "dev" branch. Then in the dev branch, tag something as "dev-[n]", where n increases with each deployment, for example: "dev-001", "dev-002" and so on.

Let's get started with a tag "dev-001":

```console
git checkout -b dev
git push origin
git tag dev-001
git push origin dev-001
```

Once you have pushed a tag, go to https://semaphoreci.com and you should see a build process running on your dev tag:

<img width="638" alt="image" src="https://github.com/networknext/next/assets/696656/868cbef3-1dc1-4c6e-a50a-efeaf54556a9">

What is going on now is that semaphore builds and runs unit tests first. Once tests pass, semaphore detects there is a tag matching "dev-[n]" and promotes to the "Upload Artifacts" job automatically. This job uploads the results of the build (the artifacts) to the google cloud storage bucket named "[company_name]_network_next_backend_artifacts" in the "Storage" project under a subdirectory with your tag name "dev-001".

<img width="1206" alt="image" src="https://github.com/networknext/next/assets/696656/1438b80b-b2fb-4661-bfca-3095d941949f">

You can view these artifacts by going to https://console.google.com and selecting the "Storage" project, and under "Cloud Storage" in the left window, then selecting the "*_backend_artifacts" storage bucket once the "Upload Artifacts" job succeeds.

Once the artifacts are uploaded, semaphore again detects that there is a tag "dev-[n]" and promotes the job "Deploy to Development" automatically. This job runs terraform internally in the subdirectory "~/next/terraform/dev/backend", with the terraform state stored in another google cloud bucket, so that terraform can be run from multiple locations (locally, on your dev machine, or inside semaphore) without conflicts.

<img width="2080" alt="image" src="https://github.com/networknext/next/assets/696656/2afbebca-5d17-45f6-b95e-35af322413bc">

This whole build process can take 10-15 minutes to run start for the first deploy. Most of the time is spent initializing redis instances and the postgres database.

Once the deploy has completed successfully, you can navigate to https://console.cloud.com, select the "Development" project and you'll see under "Compute Engine -> Managed Instance Groups" the following set of backend services running:

<img width="1210" alt="image" src="https://github.com/networknext/next/assets/696656/ebf0bbc8-88bb-40ff-b31c-e7a670cc0d3f">

## 2. Initialize the postgres database

Unfortunately, it is not possible to automatically initialize the postgres database in google cloud with terraform, so it must be done manually.

Go to https://console.google.com and select the "Development" project.

Click on the hamburger menu in the top left and select "SQL" in the left menu bar.

You should now see the postgres instance created by terraform to managed configuration of your development network next environment.

<img width="641" alt="image" src="https://github.com/networknext/next/assets/696656/0e12bcda-861a-48b7-b8cd-38f661ebd7ae">

Click on "postgres" to view the database, then select "Import" from the list of items at the top of the screen:

<img width="953" alt="image" src="https://github.com/networknext/next/assets/696656/e0e21b98-c968-4d34-8d20-a9db7f824c8f">

In the import dialog, enter the filename to import as: "[company_name]_network_next_sql_files/create.sql", and then select to import it to the "database" database instance (not "postgres" database instance).

For example:

<img width="840" alt="image" src="https://github.com/networknext/next/assets/696656/8571b102-249d-4dd9-a37a-71db6d6fdfec">

It will take just a few seconds to import. Once this is completed your postgres database is initialized with the correct schema.

## 3. Wait for SSL certificates to provision

The terraform scripts that deploy to dev automatically setup google cloud managed SSL certificates for your backend instances.

It takes time for these SSL certificates to become active. Typically, around 1 hour.

You now need to wait for these certificates to become active.

You can check their status by creating a new "dev" gcloud configuration on the command line, that points to your new "Development" project:

`gcloud init`

And then while you are in the "dev" configuration for gcloud, run:

`gcloud compute ssl-certificates list`

While they are provisioning, you will see their status column as "PROVISIONING". Do not be concerned if you see an error "FAILED_NOT_VISIBLE" this is just a benign race condition between your cloudflare DNS propagation for the domains, and the google SSL certificate verification process.

Once the certificates are all in the "ACTIVE" state you are ready to go to the next step.

```console
gaffer@batman docs % gcloud compute ssl-certificates list
NAME           TYPE     CREATION_TIMESTAMP             EXPIRE_TIME                    REGION  MANAGED_STATUS
api-dev        MANAGED  2023-11-13T12:17:11.657-08:00  2024-02-11T12:55:36.000-08:00          ACTIVE
    api-dev.virtualgo.net: ACTIVE
portal-dev     MANAGED  2023-11-13T12:17:23.310-08:00  2024-02-11T12:56:07.000-08:00          ACTIVE
    portal-dev.virtualgo.net: ACTIVE
raspberry-dev  MANAGED  2023-11-13T12:17:11.657-08:00  2024-02-11T12:55:36.000-08:00          ACTIVE
    raspberry-dev.virtualgo.net: ACTIVE
relay-dev      MANAGED  2023-11-13T12:17:23.616-08:00  2024-02-11T12:55:46.000-08:00          ACTIVE
    relay-dev.virtualgo.net: ACTIVE
```

## 4. Setup the development relays and database

We now use terraform to configure your network next development instance postgres database, and create a small amount of demonstration relays in google cloud, aws and linode (akamai).




```
