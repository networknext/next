<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run a test client

## 1. Disable raspberry clients

Switch to dev branch:

```
cd ~/next
git checkout dev
```

By default, Network Next fills the portal with simulated data from 1024 clients. These are the "raspberry" clients. They're named this because we use to run test clients on Raspberry Pi's around the vorld. 

Nowadays, they're just clients running inside a google cloud datacenter, with simulated random locations and ping data.

Let's turn these clients off so we can see some real data in the portal.

Edit `terraform/dev/backend/main.tf`

Search for "raspberry_backend". In the raspberry backend module, change `target_size` to 0:

```
module "raspberry_backend" {

  source = "../../modules/external_http_service"

  service_name = "raspberry-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a raspberry_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_HOSTNAME="${google_redis_instance.redis_raspberry.host}:6379"
    EOF
    systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "f1-micro"
  project                  = local.google_project_id
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.development.id
  default_subnetwork       = google_compute_subnetwork.development.id
  service_account          = local.google_service_account
  tags                     = ["allow-ssh", "allow-http", "allow-https"]
  domain                   = "raspberry-dev.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.raspberry-dev.id
  target_size              = 0 // 1

  depends_on = [
    module.server_backend
  ]
}
```

Search for "raspberry_client" and "raspberry_server" and make the same change.

This change modifies the managed instance groups that run the raspberry clients down to zero VM instances, effectively disabling them. 

Commit your change and deploy a new dev release:

```
git commit -am "disable raspberry clients"
git push origin
git tag dev-003
git push origin dev-003
```

It will take a few minutes, but when the change is deployed the portal will be empty:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/ee6a485a-2444-48e7-a8fe-5758c5d2df27">

## 2. Run a test client

There is already a test server running in google cloud, under a buyer account called "Test".

You can run a local client, and it will connect to this server automatically:

```
next select dev
run client
```

Wait a minute for the portal to update and you will see your test client session in the portal,  now with real ping data and your ISP name:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/2d4b263b-15f2-4e23-8f65-ea08aecb847b">

You can click on your session id, and see your session updating in real-time, once every 10 seconds:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/61cacf15-4f8a-46a2-aa23-f07c95810c9d">

Here I am in Helsinki, Finland in the Kamppi shopping center, beating google very slightly to their own datacenter with Network Next. This is in dev, which is not a particularly optimized environment, and I'm in an extremely well connected city. Much larger improvements are seen across a typical player base worldwide.

## 3. Edit the test route shader

The test buyer has a "route shader" which specifies the criteria for when players should take Network Next.

What we will do now is set a good default route shader, which will only take Network Next if a significant latency reduction is found, or if the user is experiencing significant packet loss.

Edit the file `terraform/dev/relays/main.tf` and search for "TEST BUYER"

You will find the route shader nearby:

```
resource "networknext_route_shader" test {
  name = "test"
  force_next = true
  acceptable_latency = 50
  latency_reduction_threshold = 10
  route_select_threshold = 0
  route_switch_threshold = 5
  acceptable_packet_loss = 0.1
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
}
```

Comment out the line `force_next`. This line was forcing all players to get accelerated across Network Next, even if there was no improvement.

Now the route shader looks like this:

```
resource "networknext_route_shader" test {
  name = "test"
  // force_next = true
  acceptable_latency = 50
  latency_reduction_threshold = 10
  route_select_threshold = 0
  route_switch_threshold = 5
  acceptable_packet_loss = 0.1
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
}
```

Apply the terraform changes to the postgres database:

```
cd ~/next/terraform/dev/relays
terraform apply
```

Now that you have changed the postgres database, you need to commit it to the backend for it to take effect:

```
cd ~/next
next database
next commit
```

Stop your client, then restart it:

```
next client
```

Now you should after a minute or so see your client session in the portal, but this time it is (probably) not accelerated:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/a226a658-2239-4088-ab8a-277fdc1ad49d">

Click the session id to drill in and you'll see only the blue line for the direct latency (unaccelerated route), and not the green line for accelerated latency:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/d99dd28d-3a50-4172-97bc-d35bb391374a">

This is what you want to see. Network Next is designed to _only_ accelerate players when a very significant improvement can be found. Across an entire game's player base, around 10% of players at any time will be accelerated with the settings above.

Next step: [Create your own buyer](create_your_own_buyer.md)
