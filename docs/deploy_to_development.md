<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Deploy to Development

## 1. Deploy the backend

Create a new dev branch and tag it as "dev-001":

```console
git checkout -b dev
git push origin
git tag dev-001
git push origin dev-001
```

## 2. Wait for the deploy to finish

Log in to https://semaphoreci.com

You should see a build job run on your tag, then transition to "Upload Artifacts", and then "Deploy to Development".

<img width="895" alt="image" src="https://github.com/networknext/next/assets/696656/f74cd3dc-8765-4672-bced-3d5f38098f79">

Wait until "Deploy to Development" turns green. This can take 10-15 minutes on the first deploy.

## 3. Initialize the postgres database

Go to [https://console.google.com](https://console.cloud.google.com/) and go to "SQL" under the "Development" project.

<img width="656" alt="image" src="https://github.com/networknext/next/assets/696656/41d5e129-9839-45e0-a95c-a6e80d1a0800">

Click on the "postgres" database then click on "Import".

In the import dialog, enter the filename to import the file: "[company_name]_network_next_sql_files/create.sql" to the database "database".

## 4. Wait for SSL certificates to provision

Setup a new "dev" gcloud configuration on the command line, that points to your new "Development" project:

`gcloud init`

Now you can check the status of your SSL certificates:

`gcloud compute ssl-certificates list`

Wait until all certificates are in the "ACTIVE" state before going to the next step. This usually takes around one hour.

## 5. Select dev environment and ping it

```console
next select dev
next ping
```

You should see a "pong" response from the backend, matching the tag you deployed to trigger the build:

```console
gaffer@batman next % next ping

pong [dev-633]
```

If `next select dev` fails, wait for DNS propagation to complete and try again.

## 5. Initialize the gcloud default service account

Terraform is configured to store state in a google cloud bucket. To give it permission to do this, setup the application default login like so:

`gcloud auth application-default login`

## 6. Setup relays and database

Run the terraform script:

```console
cd ~/next/terraform/dev/relays
terraform init
terraform apply
```

## 7. Commit the database changes to the backend

```console
cd ~/next
next database
next commit
```

## 8. Setup the relays

Connect to the OpenVPN instance you setup. This is necessary for the setup script to be able to SSH into the relays.

Run the relay setup script:

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

## 9. View the portal

Go to https://portal-dev.[yourdomain.com]

(portal screenshot)

Congratulations! Your development environment is now online!

Next step: [Deploy to Staging](deploy_to_staging.md)
