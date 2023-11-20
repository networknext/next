<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Deploy to Production

## 1. Deploy the backend

Create a new prod branch and tag it as "prod-001":

```console
git checkout -b prod
git push origin
git tag prod-001
git push origin prod-001
```

## 2. Initialize the postgres database

Go to https://console.cloud.google.com and go to "SQL" under the "Production" project.

Click on the "postgres" database and click on "Import".

In the import dialog, import the file "[company_name]_network_next_sql_files/create.sql" to the database "database".

## 3. Wait for SSL certificates to provision

Setup a new "prod" gcloud configuration on the command line, that points to your new "Production" project:

`gcloud init`

Now you can check the status of your SSL certificates:

`gcloud compute ssl-certificates list`

Wait until all certificates are in the "ACTIVE" state before going to the next step.

## 4. Setup the relays and database

Run the terraform script:

```console
cd ~/next/terraform/prod/relays
terraform init
terraform apply
```

## 5. Commit the database changes to the backend

```console
cd ~/next
next select prod
next database
next commit
```

## 6. Setup the relays

Connect to your OpenVPN instance, then run:

```console
next setup
```

Shortly after `next setup` completes, you should see that all relays are online:

```console
gaffer@batman next % next relays

┌────────────────────────┬──────────────────────┬──────────────────┬──────────────────┬────────┐
│ Name                   │ PublicAddress        │ InternalAddress  │ Id               │ Status │
├────────────────────────┼──────────────────────┼──────────────────┼──────────────────┼────────┤
│ akamai.atlanta         │ 74.207.225.61:40000  │                  │ 57eacb07e26af413 │ online │
│ akamai.dallas          │ 69.164.203.153:40000 │                  │ acae7ede913e1c61 │ online │
│ akamai.fremont         │ 45.56.92.195:40000   │                  │ 2c963c503cf8fbd5 │ online │
│ akamai.newyork         │ 97.107.132.170:40000 │                  │ f779a2db87b24b89 │ online │
│ google.dallas.1        │ 34.174.171.113:40000 │ 10.206.0.3:40000 │ 4dd2dfb17cbea566 │ online │
│ google.iowa.1          │ 34.28.83.51:40000    │ 10.128.0.9:40000 │ b21f535edb4bdf65 │ online │
│ google.iowa.2          │ 34.42.182.189:40000  │ 10.128.0.7:40000 │ dc42ad10d1f6bd49 │ online │
│ google.iowa.3          │ 34.173.212.153:40000 │ 10.128.0.6:40000 │ dbccf67c4e490a19 │ online │
│ google.iowa.6          │ 35.222.169.18:40000  │ 10.128.0.8:40000 │ 693255b1c056b806 │ online │
│ google.losangeles.1    │ 34.94.80.8:40000     │ 10.168.0.3:40000 │ cc3fb71d77575835 │ online │
│ google.ohio.1          │ 34.162.102.38:40000  │ 10.202.0.3:40000 │ 65cdca8e934c3f83 │ online │
│ google.oregon.1        │ 35.233.229.153:40000 │ 10.138.0.3:40000 │ 6f5870a39d40e935 │ online │
│ google.saltlakecity.1  │ 34.106.48.99:40000   │ 10.180.0.3:40000 │ f2007465ecfb8429 │ online │
│ google.southcarolina.2 │ 35.243.195.72:40000  │ 10.142.0.3:40000 │ 28584bdf56d8f4e0 │ online │
│ google.virginia.1      │ 35.199.4.54:40000    │ 10.150.0.3:40000 │ f5bc89cadf8dbdb1 │ online │
└────────────────────────┴──────────────────────┴──────────────────┴──────────────────┴────────┘
```

## 5. View the portal

Go to https://portal.[yourdomain.com]

(screenshot)

Your production environment is now online.

Next step: [Tear down Staging and Production](tear_down_staging_and_production.md)
