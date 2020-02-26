# Deployment

## Deploying the "Happy Path"

1. `make publish-backend-artifacts`
2. Restart the `relay-backend-dev-1` and `server-backend-dev-1` VMs in the **Compute Engine** -> **VM Instances** console



## Building and Publishing Artifacts

The relay and server backend artifacts just have 2 files to make deployment easy. These 2 files are compressed into a single `.tar.gz` for publishing.

- `app` binary is the compiled Go source
- `app.env` text file of the environment variables needed to run

The artifacts creates are named `(relay|server)_backend.dev.tar.gz` and end up in the `./dist` folder.

Each relay and server backends have their own `dev.env` for now which gets renamed to `app.dev` during building. In the future we will add `prod.dev` to be able to create production artifacts.

The artifacts get published to the following locations in GCP Storage which is the required location for the VMs to load them in.

- relay backend: `gs://artifacts.network-next-v3-dev.appspot.com/relay_backend.dev.tar.gz`
- server backend: `gs://artifacts.network-next-v3-dev.appspot.com/server_backend.dev.tar.gz`

## Development Instances

- relay_backend.dev.spacecats.net:40000
- server_backend.dev.spacecats.net:40000

NOTE: These domains are configured through cloudflare.com and point to the VMs External IP Addresses. These addresses can be seen in **VPC network** -> **External IP addresses** in the GCP console.

Currently there is a single VM for each backend (relay and server) running in GCP Compute Engine. Each of them are based off of their respective Instance Templates which define the CPU size, network rules, and startup script.

Instances can be seen in **Compute Engine** -> **VM instances**:

- `relay-backend-dev-1`
- `server-backend-dev-1`

When each of the instances starts up or reboots it will perform the follow steps:

1. Creates a `/app` directory in the root of the VM
2. Copies the MaxmindDB file from GCP Storage
3. Copies the artifact file for the instance `(relay|server)_backend.dev.tar.gz` from GCP Storage which was built and published with `make publish-backend-artifacts`.
4. Extracts the `app` binary and the `app.env` environment variables file into `/app`
5. Loads the environment variables with `source app.env`
6. Starts the app as a background process with `./app&`