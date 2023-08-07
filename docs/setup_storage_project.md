7<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup storage project

In this section we will setup a google cloud project called "Storage" where we will store build artifacts used by the network next backend and configuration files used by the SDK. We will upload these files to cloud storage using semaphoreci jobs.

1. Click on the project selector at the top left of the screen at http://console.cloud.google.com:

<img width="919" alt="Screenshot 2023-08-06 at 9 05 32 PM" src="https://github.com/networknext/next/assets/696656/90fa1962-a03b-4192-a823-fa955853496f">

2. Select "NEW PROJECT"
   
<img width="751" alt="Screenshot 2023-08-06 at 9 09 37 PM" src="https://github.com/networknext/next/assets/696656/2443b842-73ad-4dd1-b6d6-cd01dd757619">

3. Create a new project called "Storage"

<img width="553" alt="image" src="https://github.com/networknext/next/assets/696656/1f57d358-919b-4f80-87c5-1d3d99ac548c">

Once the project is created it is assigned a project id:

<img width="493" alt="Screenshot 2023-08-06 at 9 12 31 PM" src="https://github.com/networknext/next/assets/696656/5872da7d-df9e-443f-afac-af61e4bc1c84">

For example: storage-395201

Save this project id somewhere for later.

4. Click on "Cloud Storage" -> "Buckets" in the google cloud nav menu
   
<img width="530" alt="Screenshot 2023-08-06 at 9 17 30 PM" src="https://github.com/networknext/next/assets/696656/11eb1b6c-7f59-4e32-8e55-d06d7e21f736">

5. Create a cloud storage bucket for development artifacts
   
<img width="526" alt="Screenshot 2023-08-06 at 9 19 34 PM" src="https://github.com/networknext/next/assets/696656/18b7a2fb-ed49-48b9-b781-cd9dc8ea6d18">

<img width="811" alt="Screenshot 2023-08-06 at 9 23 14 PM" src="https://github.com/networknext/next/assets/696656/11082275-bc03-4f32-a083-065d2a881ba4">

Accept all defaults settings for the bucket and create it, when it asks you if you want to enable public access prevention, just click "CONFIRM".

6. Create a service account to be used by semaphoreci to upload files to cloud storage

Navigate to "IAM & Admin" -> "Service Accounts" in the google cloud nav menu:

<img width="522" alt="Screenshot 2023-08-06 at 9 26 49 PM" src="https://github.com/networknext/next/assets/696656/63a1b35d-23c4-4604-b0aa-a480b0854641">

Create a new service account and called "semaphore" and give it "Cloud Storage" -> "Storage admin" role, so it can upload files.

<img width="1008" alt="Screenshot 2023-08-06 at 9 29 28 PM" src="https://github.com/networknext/next/assets/696656/a8e32e06-5ae6-433f-b95d-c4a6d9ba3132">

7. Create and download a JSON key for the service account

Select "Manage Keys" for your new service account:

<img width="1696" alt="Screenshot 2023-08-06 at 9 33 51 PM" src="https://github.com/networknext/next/assets/696656/ed527ec3-b959-41b8-9cc0-650850c9da2e">

Create a new key:

<img width="842" alt="Screenshot 2023-08-06 at 9 35 47 PM" src="https://github.com/networknext/next/assets/696656/5c7258ba-8bbf-4d22-b594-86740d121166">

Select key type "JSON":

<img width="556" alt="Screenshot 2023-08-06 at 9 36 36 PM" src="https://github.com/networknext/next/assets/696656/fe964bec-9e07-460d-a2f0-f0d8b2d57e87">

The file will download to your computer automatically.

8. Setup the key and your company name in your semaphoreci account

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

9. Create a "dev" branch in your "next" project in github

This is necessary because we are building artifacts from the "dev" branch, which will upload to your dev artifacts bucket. Later on, we'll create staging and production branches and buckets too.

Commit any change in this dev branch, and make sure it triggers a semaphoreci build under "dev" branch.

<img width="810" alt="Screenshot 2023-08-06 at 9 58 24 PM" src="https://github.com/networknext/next/assets/696656/e0a1eec6-0d0f-4634-ba81-2318a7bc4485">

10. Verify that semaphoreci can upload artifacts to google cloud storage

Once the build job completes, the "Upload Artifacts" should automatically trigger in dev branch:

<img width="850" alt="Screenshot 2023-08-06 at 9 59 55 PM" src="https://github.com/networknext/next/assets/696656/0fd601b5-7e69-47af-b396-a1685f9ef879">

It should succeed and turn green in just a few seconds.

If you click on the job, it expands to show you all the artifact upload jobs that succeeded:

<img width="1159" alt="image" src="https://github.com/networknext/next/assets/696656/618b1eab-23a1-4d14-a0e1-73f91ddf5903">

11. Verify the files are uploaded to the google cloud bucket

Go back to the google cloud console and navigate to "Cloud Storage" -> "Buckets", then select your bucket called "[companyname]_network_next_dev_artifacts".

<img width="1444" alt="Screenshot 2023-08-06 at 10 03 23 PM" src="https://github.com/networknext/next/assets/696656/f80b0ad1-4431-47c7-8fdc-04bce960f5b1">

Inside this artifact you should now see some files. These files are the binaries built from the "dev" branch by semaphoreci and uploaded in the "Upload Artifacts" job. The development environment always runs binaries built from the development branch.

12. Create SDK config bucket

The SDK reads config files from a public URL to configure certain aspects like the automatic datacenter detection in public clouds and the support for multiplay.com. Next we will setup another bug for these configuration files, but this time the files will be publicly readable, so the SDK can access them.

Go back to the "Storage" google cloud project and create a new cloud storage bucket called: "[companyname]_network_next_sdk_config"

Accept the default settings for this bucket, but when it asks to enable public access protection, UNCHECK that and hit CONFIRM:

<img width="506" alt="Screenshot 2023-08-07 at 10 46 03 AM" src="https://github.com/networknext/next/assets/696656/9ac8fded-7a15-422d-a691-c9fe7b467958">

Click on "Permissions" for the bucket:

<img width="993" alt="Screenshot 2023-08-07 at 10 47 17 AM" src="https://github.com/networknext/next/assets/696656/34491e5e-8c77-4992-b7fa-30a59915edae">

Click on "Grant Access":

<img width="1015" alt="Screenshot 2023-08-07 at 10 48 49 AM" src="https://github.com/networknext/next/assets/696656/5f4021ac-bad6-43eb-b509-dda7eb77613f">

Give "allUsers" access to the bucket with permission "Cloud storage -> Storage object viewer":

13. Verify that semaphore can upload SDK config

Go back to the last successful semaphore job for your "next" project, and start the job "Upload Config":

<img width="844" alt="Screenshot 2023-08-07 at 10 54 09 AM" src="https://github.com/networknext/next/assets/696656/d575384c-dd7f-4bb4-ae4e-3b9f23507067">

It should complete and turn green in less than a minute:

<img width="818" alt="Screenshot 2023-08-07 at 10 54 53 AM" src="https://github.com/networknext/next/assets/696656/dbac8fad-d336-4ad9-a995-7bad62648e51">

14. Verify that the SDK config files are in the google cloud bucket

Go back to the google cloud bucket and verify that you see text files in it:

<img width="1525" alt="Screenshot 2023-08-07 at 10 56 33 AM" src="https://github.com/networknext/next/assets/696656/2b32609b-c318-4c43-b99b-d0c71860517b">

Make sure the files have permissions "Public to Internet" in the column as highlighted, otherwise the SDK won't be able to download the files when it runs.

_You are now ready to [setup prerequites for the dev environment](setup_prerequisites_for_dev.md)_
