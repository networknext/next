# Deployment

## Deploying the "Happy Path"

1. `make publish-backend-artifacts`
2. Restart the `relay-backend-dev-1` and `server-backend-dev-1` VMs in the **Compute Engine** -> **VM Instances** console

## Building and Publishing Artifacts

The relay and server backend artifacts just have 2 files to make deployment easy. These 2 files are compressed into a single `.tar.gz` for publishing.

- `app` binary is the compiled Go source
- `app.env` text file of the environment variables needed to run

The relay artifact contains 3 files compressed into its `.tar.gz`.

- `relay` binary is the compiled C++ source
- `relay.service` the service file that starts the relay
- `install.sh` script that is ran over ssh to take care of installing the relay

The artifacts creates are named `(portal|relay_backend|server_backend|relay).(dev|prod).tar.gz` and end up in the `./dist` folder.

Each relay and server backends have their own `(dev|prod).env` for now which gets renamed to `app.dev` during building.
Relays themselves have a `relay.env` file that is generated for each relay on update so it is not included in the artifact.

## Development

GCP Project: `network-next-v3-dev`

### GCP Artifacts

- portal: `gs://artifacts.network-next-v3-dev.appspot.com/portal.dev.tar.gz`
- relay backend: `gs://artifacts.network-next-v3-dev.appspot.com/relay_backend.dev.tar.gz`
- server backend: `gs://artifacts.network-next-v3-dev.appspot.com/server_backend.dev.tar.gz`
- relay: `gs://artifacts.network-next-v3-dev.appspot.com/relay.dev.tar.gz`

### Domains

- portal.dev.networknext.com
- relay_backend.dev.networknext.com
- server_backend.dev.networknext.com:40000

Domains are configured through cloudflare.com and point to the VMs External IP Addresses. These addresses can be seen in **VPC network** -> **External IP addresses** in the GCP console.

### TLS

The TLS certificates for portal.dev.networknext.com are configured directly inside the Go binary since the dev instances is a VM with a public IP. The certificate needed to be configure right inside the VM. So the dev instance listens on port `443` and that tells the binary to load in the same certificates in prod, but they are defined in [../transport/certificates.go](../transport/certificates.go).

The original certificates can be seen in [Cloudflare](https://dash.cloudflare.com/77635bc76eaa9a2e9226686d693209ea/networknext.com/ssl-tls/origin).

### Instances

Currently there is a single VM for each backend (portal, relay, and server) running in GCP Compute Engine. Each of them are based off of their respective Instance Templates which define the CPU size, network rules, and startup script.

Instances can be seen in **Compute Engine** -> **VM instances**:

- `portal-dev-1`
- `relay-backend-dev-1`
- `server-backend-dev-1`

## Production

GCP Project: `network-next-v3-prod`

### GCP Artifacts

- portal: `gs://us.artifacts.network-next-v3-prod.appspot.com/portal.dev.tar.gz`
- relay backend: `gs://us.artifacts.network-next-v3-prod.appspot.com/relay_backend.dev.tar.gz`
- server backend: `gs://us.artifacts.network-next-v3-prod.appspot.com/server_backend.dev.tar.gz`
- relay: `gs://us.artifacts.network-next-v3-prod.appspot.com/relay.prod.tar.gz`

### Domains

- portal.networknext.com
- portal.prod.networknext.com
- relay_backend.prod.networknext.com
- server_backend.prod.networknext.com:40000

Domains are configured through cloudflare.com and point to the Load Balancers External IP Addresses. These addresses can be seen in **VPC network** -> **External IP addresses** in the GCP console.

### TLS

The TLS certificates for portal.networknext.com are configured on the `portal-lb` balancer in GCP which is were the certificate terminates and then the traffic from the load balancer talks over port 80 to the instance group.

The original certificates can be seen in [Cloudflare](https://dash.cloudflare.com/77635bc76eaa9a2e9226686d693209ea/networknext.com/ssl-tls/origin).

### Instances

Currently there managed instance group for each backend (portal, relay, and server) running in GCP Compute Engine. Each of them are based off of their respective Instance Templates which define the CPU size, network rules, and startup script.

Instances Groups can be seen in **Compute Engine** -> **Instance groups**:

- `portal-mig`
- `relay-backend-mig`
- `server-backend-mig`

## Deploying and Reloading Instances

When each of the instances starts up or reboots it will perform the follow steps:

1. Creates a `/app` directory in the root of the VM
2. Copies the MaxmindDB file from GCP Storage
3. Stops the systemd service `app.service`
3. Copies the artifact file for the instance `(portal|relay_backend|server_backend).dev.tar.gz` from GCP Storage which was built and published with `make publish-backend-artifacts`.
4. Extracts the `app` binary, `app.env` environment variables file, and `app.service` systemd service file into `/app`
5. Moves the `app.service` file into the right place and reloads the systemd daemon
6. Starts the systemd service `app.service`