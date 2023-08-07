7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup dev environment

In this section you will create a new "Development" project in google cloud, then use terraform to setup a development environment instance in this project. This environment will use the development artifacts built and uploaded by semaphoreci from your "dev" branch in github. When this section is complete you will have a fully functional network next dev backend running in google cloud.

1. Create "Development" project in google cloud

Go to https://console.cloud.google.com and click on the project selector drop down at the top left, then select "NEW PROJECT" in the pop up:

<img width="1518" alt="Screenshot 2023-08-07 at 2 07 05 PM" src="https://github.com/networknext/next/assets/696656/3077567d-c926-42cd-99d8-634de1341ebc">

Give the project the name "Development" then hit "CREATE"

<img width="909" alt="Screenshot 2023-08-07 at 2 07 54 PM" src="https://github.com/networknext/next/assets/696656/0e3ee5ce-5d82-4f45-88ec-ea54cb975071">

Click the project selector then choose "Development" project in the pop up:

<img width="1407" alt="Screenshot 2023-08-07 at 2 09 35 PM" src="https://github.com/networknext/next/assets/696656/3d8a3045-221b-479c-bf85-ae210d642a52">

Take note of the project id for your development project. You'll need it shortly. For example: development-395218

2. Create a new service account called "terraform"

First go to "IAM & Admin" -> "Service Accounts" in the google cloud nav bar:

<img width="1415" alt="Screenshot 2023-08-07 at 2 15 18 PM" src="https://github.com/networknext/next/assets/696656/e4c14caf-ef82-43e7-a5bf-77bd27b5bed2">

Click on "CREATE SERVICE ACCOUNT" and create a new service account with the name "terraform":

<img width="882" alt="Screenshot 2023-08-07 at 2 25 01 PM" src="https://github.com/networknext/next/assets/696656/b6103afe-78c7-4ad5-b559-d45120f2e6eb">

Take note of the terraform service account name, as you will need it shortly. For example: 

Click "CREATE AND CONTINUE" then give the service account "Basic -> Editor" over the "Development" project, then click "DONE":

<img width="1137" alt="Screenshot 2023-08-07 at 2 27 30 PM" src="https://github.com/networknext/next/assets/696656/0cced877-38ef-40d4-9f4e-f171f0948b20">

3. Generate a JSON key for your service account

Click on the "Actions" drop down for your service account and select "Manage Keys":

<img width="1745" alt="Screenshot 2023-08-07 at 2 29 20 PM" src="https://github.com/networknext/next/assets/696656/354bfc13-30b3-4cc6-8525-1b7dba25b95f">

Click on "ADD KEY" then "Create new key":

<img width="844" alt="Screenshot 2023-08-07 at 2 30 19 PM" src="https://github.com/networknext/next/assets/696656/6704d055-9c41-46e9-b971-2b2bd4003bcc">

Leave the selection of key type as JSON and click "CREATE":

<img width="553" alt="Screenshot 2023-08-07 at 2 33 01 PM" src="https://github.com/networknext/next/assets/696656/95b2cf26-45b4-40d2-af31-a2eaf0882515">

The key will download to your computer as a small .json file. Create a new directory under your home directoly called "secrets" and move this json file into this directory, so it has a path of ~/secrets/terraform-development.json". This file and path must match exactly for future steps to work correctly.

4. Grant your terraform service account access to the storage buckets

Go back to the "Storage" project in Google cloud and select "Cloud Storage" -> "Buckets" in the nav menu.

Select all buckets then click "PERMISSIONS":

<img width="1271" alt="Screenshot 2023-08-07 at 6 36 45 PM" src="https://github.com/networknext/next/assets/696656/4c1f7584-bdda-4705-88b8-b7b8368225ca">

Click "ADD PRINCIPAL":

<img width="573" alt="image" src="https://github.com/networknext/next/assets/696656/e903dcbe-590e-4895-a8a4-1f3f6495febd">

Add your terraform service account full name as a principle, and grant it "Cloud storage" -> "Storage object viewer" permissions:

<img width="587" alt="Screenshot 2023-08-07 at 6 44 35 PM" src="https://github.com/networknext/next/assets/696656/2e9a533b-3e69-4968-903e-a0b5c91e10fe">

VMs started with your terraform service account in dev can now access your storage buckets in the "Storage" project.
   
4. Configure terraform variables

Edit the file: `terraform/dev/backend/terraform.tfvars` under your forked "next" repository.

Change "service_account" to the full name of the service account you just created under the "Development" google cloud project (you can get this name by clicking on "IAM & Admin -> Service Accounts" in google cloud to get a list of service accounts under "Development".

Change "project" to the id of the google cloud "Development" project, eg. development-394617 (whatever it is called in your google cloud console)

Change "artifacts_bucket" to the name of the artifacts bucket you created, eg: artifacts_bucket = "gs://[companyname]_network_next_dev_artifacts"

Change "vpn_address" to your VPN IP address. Admin functionality and portal access will only be allowed from this IP address.

5. Build the dev environment with terraform
   
Change to the directory: terraform/dev/backend

Run `terraform init`.

Run `terraform apply`.

Say "yes" to approve the terraform changes.

6. Enable various google cloud APIs in the development project as needed
   
Terraform will initially fail complaining about certain APIs not being enabled in google cloud.

For example:

```
│ Error: Error creating HealthCheck: googleapi: Error 403: Compute Engine API has not been used in project 54620502253 before or it is disabled. Enable it by visiting https://console.developers.google.com/apis/api/compute.googleapis.com/overview?project=54620502253 then retry. If you enabled this API recently, wait a few minutes for the action to propagate to our systems and retry.
```

Follow the instructions in the terrform console output and click the links to enable the google cloud features as required.

Run `terraform apply`, and iterate, enabling google cloud APIs until it succeeds.

Terraform apply will take a long time to succeed the first time it is run. It is not uncommon for it to take 10-15 minutes to finish provisioning the postgres database instance.

7. Verify managed instance groups are healthy in google cloud

Once the terraform apply succeeds, verify the managed instance groups are all healthy in google cloud.

Go to "Compute Engine -> Instance Groups" in the google cloud nav menu:

<img width="672" alt="Screenshot 2023-08-07 at 2 41 57 PM" src="https://github.com/networknext/next/assets/696656/92b17c47-d96e-4dfa-869f-5ce1fd1a7eb3">

You should see instance groups setup for all services. Within 5-10 minutes, all instance groups should turn green which means they are healthy.

<img width="1718" alt="Screenshot 2023-08-07 at 2 43 47 PM" src="https://github.com/networknext/next/assets/696656/28c775d4-6e9e-42d3-b4e3-ee6164e30ee0">

At this point terraform has taken care of setting up all VM instances, IP addresses, subnetworks and managed instance groups and health checks for your dev environment and it's all working correctly.

8. Verify REST API is operating correctly

Go to the command line under the forked 'next' directory from github.

Run `next select dev` to select the dev environment.

Connect to your OpenVPN client setup earlier.

Run `next ping` to ping the REST API for the dev environment.

You should see a response `pong`

Disconnect from the OpenVPN client.

Run `next ping` again. The command should time out and not display a pong. _The network next REST API is only accessible from your VPN._

9. Verify portal is up and operating correctly.

_todo: once the portal is ready_
