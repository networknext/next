<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Deploy to Staging

## 1. Deploy the backend

Create a new staging branch and tag it as "staging-001":

```console
git checkout -b staging
git push origin
git tag staging-001
git push origin staging-001
```

## 2. Initialize the postgres database

Go to https://console.google.com and go to "SQL" under the "Staging" project.

Click on the "postgres" database and click on "Import".

In the import dialog, enter the filename to import the file: "[company_name]_network_next_sql_files/create.sql" to the database "database".

## 3. Wait for SSL certificates to provision

Setup a new "staging" gcloud configuration on the command line, that points to your new "Staging" project:

`gcloud init`

Now you can check the status of your SSL certificates:

`gcloud compute ssl-certificates list`

Wait until all certificates are in the "ACTIVE" state before going to the next step.

## 3. View the portal

Go to https://portal-staging.[yourdomain.com]

Your staging environment is now online with 1000 simulated relays, 100k simulated servers, 100k simulated sessions (CCU).
