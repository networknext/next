<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run a test client

## 1. Disable raspberry clients

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

Now deploy a new dev release with a tag:

```
git tag dev-003
git push origin dev-003
```

This change modifies the managed instance groups that run the raspberry clients down to zero VM instances, effectively disabling them. 

It will take a few minutes, but when the change is deployed the portal will be empty:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/ee6a485a-2444-48e7-a8fe-5758c5d2df27">

## 2. Run a test client

There is already a test server running in google cloud, under a buyer account called "Test".

You can run a local client, and it will connect to this server automatically:

```
next select dev
run client
```

Wait a minute for the portal to display the data (it updates once per-minute).

You will see your test clients in the portal, now with real ping data and your ISP name:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/2d4b263b-15f2-4e23-8f65-ea08aecb847b">

Here I am in Helsinki, Finland in the Kamppi shopping center, beating google to their own datacenter with Network Next. This is in dev, which is not a particularly optimized environment, and I'm in an extremely well connected city. Much larger improvements are seen across a typical player base worldwide.

I can click on the session id, and see my session updating in real-time, once every 10 seconds:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/61cacf15-4f8a-46a2-aa23-f07c95810c9d">

## 3. Edit the test route shader

The test client has a "route shader" which specifies the criteria for when we should take Network Next.

What we will do now is set a good default route shader, which will only take Network Next if a significant latency reduction is found, or if the user is experiencing significant packet loss.

Edit the file `terraform/dev/relays/main.tf` and search for "TEST BUYER"

You will find this:

```
resource "networknext_route_shader" test {
  name = "test"
  force_next = true
  acceptable_latency = 50
  latency_reduction_threshold = 10
  route_select_threshold = 0
  route_switch_threshold = 5
  acceptable_packet_loss_instant = 0.25
  acceptable_packet_loss_sustained = 0.1
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
}
```

Comment out the line `force_next`. This line forced all players to get accelerated across Network Next, even if there was no improvement.

Now the route shader looks like this:

```
resource "networknext_route_shader" test {
  name = "test"
  // force_next = true
  acceptable_latency = 50
  latency_reduction_threshold = 10
  route_select_threshold = 0
  route_switch_threshold = 5
  acceptable_packet_loss_instant = 0.25
  acceptable_packet_loss_sustained = 0.1
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
}
```

This route shader is configured as follows:

* `acceptable_latency = 50`. Do not accelerate any player when their latency is already below 50ms. This is good enough and not worth accelerating below. This is a recommended value to start with.
  
* `latency_reduction_threshold = 10`. Do not accelerate a player unless we can find a latency reduction of _at least_ 10 milliseconds.

* `route_select_threshold = 0`. This finds the absolute lowest latency route, out of all routes available. In the future, it is best to relax this to 5ms, so we load balance across different routes, instead of forcing everybody down the same route.

* `route_switch_threshold = 5`. Hold the current Network Next route, unless a better route is available with at least 5ms lower latency than the current route. Don't set this too low, or the route will flap around every 10 seconds. In the future, I recommend that you increase this to 10ms, but right now 5ms is fine.

* `acceptable_packet_loss_instanct = 0.25`. If packet loss > 0.25% occurs in any 10 second period, accelerate the player to reduce packet loss. This catches packet loss spikes.

* `acceptable_packet_loss_sustained = 0.1`. If packet loss > 0.1% occurs for 30 seconds, accelerate the player to reduce packet loss. This captures packet loss that is lower in intensity, but is sustained over a longer period.

* `bandwidth_envelope_up_kbps = 256`. This is the maximum bandwidth in kilobits per-second, kilobits, not kilobytes, sent from client to server. We don't need to change this, but later on when you setup your own buyer you should adjust it to the maximum bandwidth your client will send up to to the server. If the client sends more bandwidth than this, it will not be accelerated (eg. during a load screen, this is typically OK).

* `bandwidth_envelope_down_kbps = 256`. The bandwidth down from server to client in kilobits per-second. Again, we don't need to change this yet.

Now, deploy the terraform changes to the postgres database:

```
cd ~/next/terraform/dev/relays
terraform apply
```

Now that we have changed the postgres database, we need to commit it to the backend for it to take effect:

```
cd ~/next
next database
next commit
```

Stop your client, then restart it:

```
next client
```

Now you should after a minute or so see your client session in the portal, but this time it is (almost certainly) not accelerated:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/a226a658-2239-4088-ab8a-277fdc1ad49d">

Click the session id to drill into the session and you'll see only the blue line for the direct latency (unaccelerated route), and not the green line for accelerated latency:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/d99dd28d-3a50-4172-97bc-d35bb391374a">

This is what we want. Network Next is designed to _only_ accelerate players when a very significant improvement can be found. Across an entire game's player base, around 10% of players at any time will be accelerated with the settings above.

Next step: [Add your own buyer](add_your_own_buyer.md)
