7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup dev environment

In this section you will create a new "Development" project in google cloud, then use terraform to setup a development environment instance in this project. This environment will use the development artifacts built and uploaded by semaphoreci from your "dev" branch in github. When this section is complete you will have a fully functional network next dev backend running in google cloud as well as a process for developing code changes to this instance.

1. Create "Development" project in google cloud

Go to https://console.cloud.google.com and click on the project selector drop down at the top left, then select "NEW PROJECT" in the pop up:

<img width="1518" alt="Screenshot 2023-08-07 at 2 07 05 PM" src="https://github.com/networknext/next/assets/696656/3077567d-c926-42cd-99d8-634de1341ebc">

Give the project the name "Development" then hit "CREATE"

<img width="909" alt="Screenshot 2023-08-07 at 2 07 54 PM" src="https://github.com/networknext/next/assets/696656/0e3ee5ce-5d82-4f45-88ec-ea54cb975071">

Click the project selector then choose "Development" project in the pop up:

<img width="1407" alt="Screenshot 2023-08-07 at 2 09 35 PM" src="https://github.com/networknext/next/assets/696656/3d8a3045-221b-479c-bf85-ae210d642a52">

Take note of the project id for your development project. You'll need it shortly. For example: development-395218

2. Create a new service account called "terraform" and give it access to your project

We are going to use terraform to provision the development environment. In order for Terraform to have the access it needs a service account that it operates under. We will export a JSON key file to your computer that gives terraform running locally authentication using actions under the identify of this service account.

First go to "IAM & Admin" -> "Service Accounts" in the google cloud nav bar:

<img width="1415" alt="Screenshot 2023-08-07 at 2 15 18 PM" src="https://github.com/networknext/next/assets/696656/e4c14caf-ef82-43e7-a5bf-77bd27b5bed2">

Click on "CREATE SERVICE ACCOUNT" and create a new service account with the name "terraform":

<img width="882" alt="Screenshot 2023-08-07 at 2 25 01 PM" src="https://github.com/networknext/next/assets/696656/b6103afe-78c7-4ad5-b559-d45120f2e6eb">

Click "CREATE AND CONTINUE" then give the service account "Basic -> Editor" over the "Development" project, then click "DONE":

<img width="1137" alt="Screenshot 2023-08-07 at 2 27 30 PM" src="https://github.com/networknext/next/assets/696656/0cced877-38ef-40d4-9f4e-f171f0948b20">

3. Generate a JSON key for your service account

Click on the "Actions" drop down for your service account and select "Manage Keys":

<img width="1745" alt="Screenshot 2023-08-07 at 2 29 20 PM" src="https://github.com/networknext/next/assets/696656/354bfc13-30b3-4cc6-8525-1b7dba25b95f">

Click on "ADD KEY" then "Create new key":

<img width="844" alt="Screenshot 2023-08-07 at 2 30 19 PM" src="https://github.com/networknext/next/assets/696656/6704d055-9c41-46e9-b971-2b2bd4003bcc">










	Now switch to the "Development" project in google cloud.

	In this project, add a new service account called "terraform".


	Save the full name of the service account somewhere, eg: terraform@development-394617.iam.gserviceaccount.com

	Create a JSON key for the service account and download it.

	-----------

	Create a new directory under your home directory: ~/secrets

	Move the downloaded JSON key to: ~/secrets/terraform-development.json

	-----------

	Edit the file: terraform/dev/backend/terraform.tfvars under the "next" repository.

	Change "service_account" to the name of the service account you just created under the "Development" google cloud project.

	Change "project" to the name of the google cloud project, eg. development-394617 (whatever it is called in your google cloud console)

	Change "artifacts_bucket" to the name of the artifacts bucket you created, eg: artifacts_bucket = "gs://[companyname]_network_next_dev_artifacts"

	Change "vpn_address" to your VPN IP address. Admin functionality will only be allowed from this IP address.

	-----------

	Change to the directory: terraform/dev/backend

	Run "terraform init".

	Run "terraform apply".

	Say "yes" to approve the terraform changes.

	-----------

	Terraform will initially fail complaining about certain APIs not being enabled in google cloud.

	Follow the instructions in the output and enable the google cloud features as required.

	Run "terraform apply", and iterate, fixing disabled APIs until it succeeds.

	Terraform apply will take a long time to succeed on the first pass. It is not uncommon for it to take 10-15 minutes to finish provisioning the postgres database instance.

	-----------

	Once the terraform apply succeeds, verify the managed instance groups are all healthy in google cloud.

	-----------

	Cloudflare step. Point dev.* 3 domains at load balancer IPs

	-----------

	next ping

	-----------

