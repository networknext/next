<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Deploy to Staging

## 1. Deploy the backend

Create a new staging branch and tag it as "staging-001":

```console
cd ~/next
git checkout -b staging
git push --set-upstream origin staging
git push origin
git tag staging-001
git push origin staging-001
```

## 2. Wait for the semaphore ci deploy to complete

Wait until "Deploy to Staging" is green in semaphore ci:

<img width="622" alt="image" src="https://github.com/networknext/next/assets/696656/9d98c3f8-180d-4248-b2a6-ee799ff3668b">

## 2. Initialize the postgres database

Go to https://console.cloud.google.com and navigate to "SQL" under the "Staging" project.

Click on the "postgres" database and import two files into "database", in order:

1. "[company_name]_network_next_sql_files/create.sql"
2. "[company_name]_network_next_sql_files/staging.sql"

## 3. Wait for SSL certificates to provision

Setup a new "staging" gcloud configuration on the command line, that points to your new "Staging" project:

`gcloud init`

Now you can check the status of your SSL certificates:

`gcloud compute ssl-certificates list`

Wait until all certificates are in the "ACTIVE" state before going to the next step.

## 4. Verify that all services are green

Go to https://console.google.com and navigate to "Compute Engine -> Instance Groups" under the "Staging" project.

You should see all services up and running and green:

<img width="960" alt="image" src="https://github.com/networknext/next/assets/696656/e190ae8f-2bf1-4ec3-bc1d-ef7c7f630797">

## 5. View the portal

Go to https://portal-staging.[yourdomain.com]

<img width="1547" alt="image" src="https://github.com/networknext/next/assets/696656/3b7ea95b-6c8d-4422-b2a6-519450738b5f">

Congratulations! Your staging environment is now online with 1000 simulated relays, 100k simulated servers and 100k simulated sessions!

Next step: [Deploy to Production](deploy_to_production.md)
