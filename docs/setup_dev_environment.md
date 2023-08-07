7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup dev environment

In this section you will create a new "Development" project in google cloud, then use terraform to setup a development environment instance in this project. This environment will use the development artifacts published by previous steps built in semaphoreci from your "dev" branch in github. When this section is complete you will have a fully functional network next dev backend running in google cloud.

1. Create "Development" project in google cloud

Go to https://console.cloud.google.com and click on the project selector drop down at the top left, then select "NEW PROJECT" in the pop up:

<img width="1518" alt="Screenshot 2023-08-07 at 2 07 05 PM" src="https://github.com/networknext/next/assets/696656/3077567d-c926-42cd-99d8-634de1341ebc">

Give the project the name "Development" then hit "CREATE"

<img width="909" alt="Screenshot 2023-08-07 at 2 07 54 PM" src="https://github.com/networknext/next/assets/696656/0e3ee5ce-5d82-4f45-88ec-ea54cb975071">

Click the project selector then choose "Development" project in the pop up:

<img width="1407" alt="Screenshot 2023-08-07 at 2 09 35 PM" src="https://github.com/networknext/next/assets/696656/3d8a3045-221b-479c-bf85-ae210d642a52">

Take note of the project id for your development project. You'll need it shortly. For example: development-395218

2. Create a new service account called "terraform" and give it "Admin" access to your project

We are going to use terraform to provision the development environment. In order for Terraform to have the access it needs a service account that it operates under. We will export a JSON key file to your computer that gives terraform running locally authentication using actions under the identify of this service account.

First go to "IAM & Admin" -> "Service Accounts" in the google cloud nav bar:

<img width="1415" alt="Screenshot 2023-08-07 at 2 15 18 PM" src="https://github.com/networknext/next/assets/696656/e4c14caf-ef82-43e7-a5bf-77bd27b5bed2">

Make sure that you do this in the context of the new "Development" project that you created in the previous step.

3. 
