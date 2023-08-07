7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup storage project

In this section we will setup a google cloud project called "Storage" and configure it with terraform. 

This project is where we store build artifacts used by the network next backend and configuration files used by the SDK. We will upload these files to cloud storage using semaphoreci jobs.

1. Click on the project selector at the top left of the screen at http://console.cloud.google.com:

<img width="919" alt="Screenshot 2023-08-06 at 9 05 32 PM" src="https://github.com/networknext/next/assets/696656/90fa1962-a03b-4192-a823-fa955853496f">

2. Select "NEW PROJECT"
   
<img width="751" alt="Screenshot 2023-08-06 at 9 09 37 PM" src="https://github.com/networknext/next/assets/696656/2443b842-73ad-4dd1-b6d6-cd01dd757619">

3. Create a new project called "Storage"

<img width="553" alt="image" src="https://github.com/networknext/next/assets/696656/1f57d358-919b-4f80-87c5-1d3d99ac548c">

Once the project is created it is assigned a project id:

<img width="493" alt="Screenshot 2023-08-06 at 9 12 31 PM" src="https://github.com/networknext/next/assets/696656/5872da7d-df9e-443f-afac-af61e4bc1c84">

For example: storage-395201

Save this project id somewhere, as you'll need it shortly.

4. Create a service account to be used by terraform to configure the "Storage" project

Navigate to "IAM & Admin" -> "Service Accounts" in the google cloud nav menu:

<img width="522" alt="Screenshot 2023-08-06 at 9 26 49 PM" src="https://github.com/networknext/next/assets/696656/63a1b35d-23c4-4604-b0aa-a480b0854641">

Create a new service account called "terraform":

<img width="857" alt="Screenshot 2023-08-07 at 4 10 14 PM" src="https://github.com/networknext/next/assets/696656/36fdafc8-0e94-4a01-a987-24ea7b0f6f93">

Give it "Basic" -> "Editor" permissions so it can modify the "Storage" project in google cloud:

<img width="1150" alt="Screenshot 2023-08-07 at 4 10 57 PM" src="https://github.com/networknext/next/assets/696656/196e6d60-25db-4e14-bdc8-2521db4b16ea">

5. Create and download a JSON key for the terraform service account

Select "Manage Keys" under the "Actions" drop down for the service account:

<img width="2038" alt="Screenshot 2023-08-07 at 4 11 57 PM" src="https://github.com/networknext/next/assets/696656/e2642738-1932-4c6d-87ec-e4a9181d16a1">

Click "ADD KEY" -> "Create new key":

<img width="955" alt="Screenshot 2023-08-07 at 4 13 13 PM" src="https://github.com/networknext/next/assets/696656/4c349b4f-38a9-4614-ae06-94b21776b61c">

Accept the default key type of JSON and just hit "CONFIRM":

<img width="556" alt="Screenshot 2023-08-07 at 4 14 06 PM" src="https://github.com/networknext/next/assets/696656/fd19caf1-c064-4583-a79b-c8f30b6ad599">

A json key file will download to your computer at this point. Create a new directory under your home directory called `~/secrets` and move the json file into this directory so that it has the path `~/secrets/terraform-storage.json`. The filename must be exact or the terraform setup process will not work.

6. Edit the terraform storage project configuration

Open the file next/terraform/storage/terraform.tfvars in an editor. This is the configuration file for the terraform script that will configure your "Storage" project for you in google cloud.

`
credentials       = "~/secrets/terraform-storage.json"
project           = "storage-379422"
location          = "US"
region            = "us-central1"
zone              = "us-central1-c"
dev_artifacts     = "masbandwidth_network_next_dev_artifacts"
staging_artifacts = "masbandwidth_network_next_staging_artifacts"
prod_artifacts    = "masbandwidth_network_next_prod_artifacts"
relay_artifacts   = "masbandwidth_network_next_relay_artifacts"
sdk_config        = "masbandwidth_network_next_sdk_config"
`

Please edit the text file such that the bucket names match your company name, for example: dev_artifacts, staging_artifacts, prod_artifacts, relay_artifacts and sdk_config need to be adjusted so that "masbandwidth" is replaced with your company name.

Modify the "project" variable to the project id of your "Storage" project in google cloud.

7. Download and install terraform

You can download terraform from https://www.terraform.io if you don't have it already.

If you are using MacOS, the easiest way to get it is via https://brew.sh - eg. "brew install terraform"

8. Setup the storage project with terraform

This step will provision the storage project in google cloud with the correctly setup cloud storage buckets to store development, staging and production branch artifacts used by the backend, as well as database initialization files, relay binaries and configuration files which will be read by the SDK when deployed.

Change to the directory `next/terraform/storage`

Run `terraform init`

Then run `terraform apply` and enter "yes".

9. Verify that the cloud storage buckets have been created in your "Storage" project

Go to "Cloud Storage" -> "Buckets" via the google cloud nav menu. You should see buckets created with your company name at the start:

<img width="1531" alt="Screenshot 2023-08-07 at 4 36 34 PM" src="https://github.com/networknext/next/assets/696656/688456e7-cb56-49ee-8ae7-c23be9f9c10f">

10. Create a service account to be used by semaphoreci to upload files to cloud storage

Now we need to create a second service account which will be used by semaphoreci to upload files into the cloud storage buckets you just created with terraform.

Create a new service account and called "semaphore" and give it "Cloud Storage" -> "Storage admin" role, so it can upload files.

<img width="1008" alt="Screenshot 2023-08-06 at 9 29 28 PM" src="https://github.com/networknext/next/assets/696656/a8e32e06-5ae6-433f-b95d-c4a6d9ba3132">

11. Create and download a JSON key for the semaphore service account

Select "Manage Keys" for your new service account:

<img width="1696" alt="Screenshot 2023-08-06 at 9 33 51 PM" src="https://github.com/networknext/next/assets/696656/ed527ec3-b959-41b8-9cc0-650850c9da2e">

Create a new key:

<img width="842" alt="Screenshot 2023-08-06 at 9 35 47 PM" src="https://github.com/networknext/next/assets/696656/5c7258ba-8bbf-4d22-b594-86740d121166">

Select key type "JSON":

<img width="556" alt="Screenshot 2023-08-06 at 9 36 36 PM" src="https://github.com/networknext/next/assets/696656/fe964bec-9e07-460d-a2f0-f0d8b2d57e87">

The file will download to your computer automatically.

12. Setup the key and your company name in your semaphoreci account

Select "Settings" in the top right menu in semaphoreci:

<img width="556" alt="Screenshot 2023-08-06 at 9 38 38 PM" src="https://github.com/networknext/next/assets/696656/05d566ce-ea64-4749-9477-01563fc69d82">

Go to "Secrets" and select "New Secret":

<img width="1249" alt="Screenshot 2023-08-06 at 9 39 31 PM" src="https://github.com/networknext/next/assets/696656/2bfdf982-50d1-46fa-a5e3-d530904495bf">

Create a secret with name "google-cloud-platform", of type "Configuration file", with filename of "/home/semaphore/gcp-service-account.json", then click to upload the JSON key file you just downloaded from google cloud in the previous step:

<img width="629" alt="Screenshot 2023-08-06 at 9 42 28 PM" src="https://github.com/networknext/next/assets/696656/2440670f-947e-45ab-8762-baaf3bc348d8">

The secret is now created, and should look like this:

<img width="578" alt="Screenshot 2023-08-06 at 9 44 40 PM" src="https://github.com/networknext/next/assets/696656/f21f4132-b0a6-4c26-8f4a-af4abcae289b">

Create a second secret, and call it "company-name" of type Env var, and set it to COMPANY_NAME="yourcompanyname":

<img width="632" alt="image" src="https://github.com/networknext/next/assets/696656/7377e722-8e2e-4284-9f83-7efc4e82f532">

The company name must match exactly the company name you used above when creating the google cloud storage bucket.

13. Create a "dev" branch in your "next" project in github

This is necessary because we are building artifacts from the "dev" branch, which will upload to your dev artifacts bucket. Later on, we'll create staging and production branches and buckets too.

Commit any change in this dev branch, and make sure it triggers a semaphoreci build under "dev" branch.

<img width="810" alt="Screenshot 2023-08-06 at 9 58 24 PM" src="https://github.com/networknext/next/assets/696656/e0a1eec6-0d0f-4634-ba81-2318a7bc4485">

14. Verify that semaphoreci can upload artifacts to google cloud storage

Once the build job completes, the "Upload Artifacts" should automatically trigger in dev branch:

<img width="850" alt="Screenshot 2023-08-06 at 9 59 55 PM" src="https://github.com/networknext/next/assets/696656/0fd601b5-7e69-47af-b396-a1685f9ef879">

It should succeed and turn green in just a few seconds.

If you click on the job, it expands to show you all the artifact upload jobs that succeeded:

<img width="1159" alt="image" src="https://github.com/networknext/next/assets/696656/618b1eab-23a1-4d14-a0e1-73f91ddf5903">

15. Verify the files are uploaded to the google cloud bucket

Go back to the google cloud console and navigate to "Cloud Storage" -> "Buckets", then select your bucket called "[companyname]_network_next_dev_artifacts".

<img width="1444" alt="Screenshot 2023-08-06 at 10 03 23 PM" src="https://github.com/networknext/next/assets/696656/f80b0ad1-4431-47c7-8fdc-04bce960f5b1">

Inside this artifact you should now see some files. These files are the binaries built from the "dev" branch by semaphoreci and uploaded in the "Upload Artifacts" job. The development environment always runs binaries built from the development branch.

16. Upload SDK config

The SDK reads config files from a public URL to configure certain aspects like the automatic datacenter detection in public clouds and the support for multiplay.com. Next we will setup another bug for these configuration files, but this time the files will be publicly readable, so the SDK can access them.

Go back to the last successful semaphore job for your "next" project, and start the job "Upload Config":

<img width="844" alt="Screenshot 2023-08-07 at 10 54 09 AM" src="https://github.com/networknext/next/assets/696656/d575384c-dd7f-4bb4-ae4e-3b9f23507067">

It should complete and turn green in less than a minute:

<img width="818" alt="Screenshot 2023-08-07 at 10 54 53 AM" src="https://github.com/networknext/next/assets/696656/dbac8fad-d336-4ad9-a995-7bad62648e51">

17. Verify that the SDK config files are in the google cloud bucket

Go back to the google cloud bucket and verify that you see text files in it:

<img width="1525" alt="Screenshot 2023-08-07 at 10 56 33 AM" src="https://github.com/networknext/next/assets/696656/2b32609b-c318-4c43-b99b-d0c71860517b">

_You are now ready to [setup prerequites for the dev environment](setup_prerequisites_for_dev.md)_

