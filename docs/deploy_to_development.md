<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Deploy to Development

## 1. Deploy the backend

Network Next uses tag based deployments from branches.

To deploy to dev, first you merge your code into the "dev" branch. Then in the dev branch, tag something as "dev-[n]", where n increases with each deployment, for example: "dev-001", "dev-002" and so on.

Let's get started with a tag "dev-001":

```console
git checkout -b dev
git push origin
git tag dev-001
git push origin dev-001
```

Once you have pushed a tag, go to https://semaphoreci.com and you should see a build job running on your dev tag:

<img width="638" alt="image" src="https://github.com/networknext/next/assets/696656/868cbef3-1dc1-4c6e-a50a-efeaf54556a9">

What is going on now is:

1. semaphore builds and runs unit tests
2. Once tests pass, semaphore detects there is a tag matching "dev-[n]" and promotes to the "Upload Artifacts" job automatically.
3. This job uploads the results of the build (the artifacts) to the google cloud storage bucket named
   <your_company_name>_network_next_backend_artifacts in the "Storage" project under a subdirectory matching your tag.

<img width="1206" alt="image" src="https://github.com/networknext/next/assets/696656/1438b80b-b2fb-4661-bfca-3095d941949f">

You can view these artifacts by going to https://console.google.com and selecting the "Storage" project, and under "Cloud Storage" in the left window, then selecting the "*_backend_artifacts" storage bucket once the "Upload Artifacts" job succeeds.

Once the artifacts are uploaded:

4. semaphore again detects that there is a tag "dev-[n]" and promotes the job "Deploy to Development" automatically.
5. This job runs terraform internally in the subdirectory "~/next/terraform/dev/backend", with the terraform state stored in another google cloud bucket, so that terraform can be run from multiple locations (locally, on your dev machine, or inside semaphore) without conflicts.

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

In the import dialog, enter the filename to import as: "[company_name]_network_next_sql_files/create.sql", and import it to the "database" db instance (not "postgres").

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

While they are provisioning, you will see their status column as "PROVISIONING". Do not be concerned if you see an error "FAILED_NOT_VISIBLE" this is just a benign race condition between the cloudflare DNS propagation and the google SSL certificate verification process.

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

We will now use terraform to initialize your development environment with a set of datacenters, google, aws and akamai sellers, and some relays to steer traffic. We will also create a test "Raspberry" customer which will drive synthetic traffic data to the portal so you can see everything working.

First, run the terraform script:

```console
cd ~/next/terraform/dev/relays
terraform init
terraform apply
```

You will see output like this on the `terraform apply` step:

```console
          + name               = "google.virginia.1"
          + notes              = ""
          + port_speed         = 0
          + private_key_base64 = (known after apply)
          + public_ip          = (known after apply)
          + public_key_base64  = (known after apply)
          + public_port        = 40000
          + ssh_ip             = (known after apply)
          + ssh_port           = 22
          + ssh_user           = "ubuntu"
          + version            = "relay-debug-1.0.28"
        }
    }
  ~ database_relays = {
      + "akamai.atlanta"         = (known after apply)
      + "akamai.dallas"          = (known after apply)
      + "akamai.fremont"         = (known after apply)
      + "akamai.newyork"         = (known after apply)
      + "google.dallas.1"        = (known after apply)
      + "google.iowa.1"          = (known after apply)
      + "google.iowa.2"          = (known after apply)
      + "google.iowa.3"          = (known after apply)
      + "google.iowa.6"          = (known after apply)
      + "google.losangeles.1"    = (known after apply)
      + "google.ohio.1"          = (known after apply)
      + "google.oregon.1"        = (known after apply)
      + "google.saltlakecity.1"  = (known after apply)
      + "google.southcarolina.2" = (known after apply)
      + "google.virginia.1"      = (known after apply)
    }

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value:
```

Say "yes".

This will take several minutes to run. Terraform communicates with your development network next API service, and google cloud, AWS and linode REST APIs to provision relays and setup the postgres database will all information  it needs to run.

Once it completes successfully, the postgres database is setup, and the relays are provisioned in google cloud, AWS and linode (akamai) for you, but they are not yet linked up to the backend runtime.

To link everything together:

```console
cd ~/next
next select dev
next database
next commit
```

The steps are as follows:

1. Select the dev environment, so the next tool is now commicating with the API at https://api-dev.[yourdomain.com]

2. Download the postgres database into a "database.bin" file, a binary representation of the runtime data needed by the network next backend. This way the network next runtime never communicates directly with the postgres database and is able to operate even if postgres is down.

3. Commit this database.bin to the backend. It becomes active with the development backend runtime within one minute.

Next, you must connect to your VPN to setup the relays, as they have already been configured by terraform to only allow SSH from your VPN address.

While connected to your VPN:

```console
next setup
```

The code pass over each relay defined in the postgres database, SSH's into it, downloads the relay binary and sets it up to run as a service and starts the service.

```console
gaffer@batman next % next setup

setting up relay google.losangeles.1
Warning: Permanently added '34.94.80.8' (ED25519) to the list of known hosts.
making the relay prompt cool
downloading relay binary
--2023-11-15 16:59:36--  https://storage.googleapis.com/alocasia_network_next_relay_artifacts/relay-debug-1.0.28
Resolving storage.googleapis.com (storage.googleapis.com)... 172.217.12.155, 172.217.14.91, 142.250.68.91, ...
Connecting to storage.googleapis.com (storage.googleapis.com)|172.217.12.155|:443... connected.
HTTP request sent, awaiting response... 200 OK
Length: 393120 (384K) [application/octet-stream]
Saving to: ‘relay-debug-1.0.28’

relay-debug-1.0.28                                              100%[====================================================================================================================================================>] 383.91K  --.-KB/s    in 0.006s

2023-11-15 16:59:36 (66.9 MB/s) - ‘relay-debug-1.0.28’ saved [393120/393120]

setting up relay environment
setting up relay service file
moving everything into /app
limiting max journalctl logs to 200MB
installing relay service
Created symlink /etc/systemd/system/multi-user.target.wants/relay.service → /app/relay.service.
Created symlink /etc/systemd/system/relay.service → /app/relay.service.
starting relay service
setup completed
Connection to 34.94.80.8 closed.
etc...
```

Once everything is working correctly, you'll be able to see that all relays are online:

```console
gaffer@batman next % next relays

┌────────────────────────┬──────────────────────┬──────────────────┬──────────────────┬─────────┐
│ Name                   │ PublicAddress        │ InternalAddress  │ Id               │ Status  │
├────────────────────────┼──────────────────────┼──────────────────┼──────────────────┼─────────┤
│ akamai.atlanta         │ 74.207.225.61:40000  │                  │ 57eacb07e26af413 │ offline │
│ akamai.dallas          │ 69.164.203.153:40000 │                  │ acae7ede913e1c61 │ offline │
│ akamai.fremont         │ 45.56.92.195:40000   │                  │ 2c963c503cf8fbd5 │ offline │
│ akamai.newyork         │ 97.107.132.170:40000 │                  │ f779a2db87b24b89 │ offline │
│ google.dallas.1        │ 34.174.171.113:40000 │ 10.206.0.3:40000 │ 4dd2dfb17cbea566 │ offline │
│ google.iowa.1          │ 34.28.83.51:40000    │ 10.128.0.9:40000 │ b21f535edb4bdf65 │ offline │
│ google.iowa.2          │ 34.42.182.189:40000  │ 10.128.0.7:40000 │ dc42ad10d1f6bd49 │ offline │
│ google.iowa.3          │ 34.173.212.153:40000 │ 10.128.0.6:40000 │ dbccf67c4e490a19 │ offline │
│ google.iowa.6          │ 35.222.169.18:40000  │ 10.128.0.8:40000 │ 693255b1c056b806 │ offline │
│ google.losangeles.1    │ 34.94.80.8:40000     │ 10.168.0.3:40000 │ cc3fb71d77575835 │ offline │
│ google.ohio.1          │ 34.162.102.38:40000  │ 10.202.0.3:40000 │ 65cdca8e934c3f83 │ offline │
│ google.oregon.1        │ 35.233.229.153:40000 │ 10.138.0.3:40000 │ 6f5870a39d40e935 │ offline │
│ google.saltlakecity.1  │ 34.106.48.99:40000   │ 10.180.0.3:40000 │ f2007465ecfb8429 │ offline │
│ google.southcarolina.2 │ 35.243.195.72:40000  │ 10.142.0.3:40000 │ 28584bdf56d8f4e0 │ offline │
│ google.virginia.1      │ 35.199.4.54:40000    │ 10.150.0.3:40000 │ f5bc89cadf8dbdb1 │ offline │
└────────────────────────┴──────────────────────┴──────────────────┴──────────────────┴─────────┘
```

## 5. View the portal

Go to https://portal-dev.[yourdomain.com]

You should see a map

(image of map, with session counts).

The system is setup to run with synthetic data from 1024 clients, with roughly 10% accelerated at any time.
