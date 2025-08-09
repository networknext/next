<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Getting production ready for your game

Here are some things you must do before you go into production with your Network Next system.

## 1. Replace google redis cluster with redis enterprise

The default redis cluster implementation in Google cloud is in beta release, and unfortunately does not currently allow you to set required redis options on the cluster that are needed in production.

The specific options that are needed on the cluster are:

```
maxmemory <x>gb
maxmemory-policy allkeys-lru
```

Where x is some amount of memory that gives you some headroom vs. the actual amount of working memory in your cluster. 

This combination of settings are required to make sure that the portal expires keys. Without these settings, the redis cluster will fill and stop working in production pretty quickly at load. The goal here is to make sure that we store the maximum amount of portal history in your redis cluster, according to the amount of memory it has.

The good news is that this setting _is_ available on Redis Enterprise, which is available in Google Cloud marketplace. However, this is not easily setup via terraform, hence this manual step.

First, delete the portal redis cluster in `~/next/terraform/prod/backend/main.tf`:

Turn this:

```
resource "google_redis_cluster" "portal" {
  provider       = google-beta
  name           = "portal"
  shard_count    = 10
  psc_configs {
    network = google_compute_network.production.id
  }
  region = "us-central1"
  replica_count = 1
  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_DISABLED"
  authorization_mode = "AUTH_MODE_DISABLED"
  depends_on = [
    google_network_connectivity_service_connection_policy.default
  ]
}

resource "google_network_connectivity_service_connection_policy" "default" {
  provider = google-beta
  name = "redis"
  location = "us-central1"
  service_class = "gcp-memorystore-redis"
  description   = "redis cluster service connection policy"
  network = google_compute_network.production.id
  psc_config {
    subnetworks = [google_compute_subnetwork.production.id]
  }
}

locals {
  redis_portal_address = "${google_redis_cluster.portal.discovery_endpoints[0].address}:6379"
}
```

Into this:

```
locals {
  redis_portal_address = "127.0.0.1:6379"
}
```

Then go to the marketplace in Google Cloud and sign up for "Redis Enterprise" (search for it).

Follow the instructions to setup an instance, making sure you set appropriate options for `maxmemory <x>gb` and `maxmemory-policy volatile-lru`.

Follow instructions to link the redis cluster to your production network.

Once you have the instance setup, modify `redis_portal_address` and set it to the IP address and port of the discovery endpoint for your redis cluster.

```
locals {
  redis_portal_address = "<redis_cluster_discovery_address>:6379"
}
```

Trigger a deploy from prod branch with a tag. Once the deploy completes, the production environment should be switched over to your new redis enterprise cluster, and the google cloud redis cluster is shut down.

You can verify everything is working correctly just by looking at the portal. If the portal is working, then everything is fine. The portal cannot work without redis.

## 2. Right size your redis cluster

There are several factors you must consider:

1. The amount of throughput and CPU on the cluster
2. The amount of history you desire for the portal
3. How much availability you want on the cluster

The backend will fill your redis cluster and only expire the oldest entries it needs to keep under max memory by design. So in short, you need to decide how much history you want to have in the portal, work out roughly how much session data you can store (how many hours or days) at expected load for an amount of memory, and plan around that. Next, once you have the amount of memory you want, make sure that you are not using too much CPU or IO on your cluster at the amount of memory you choose. Finally, depending on how much redundancy you want, you can try running multiple replicas in different zones and so on, in case one shard goes down there is an active mirrored fallback.

I recommend for a production environment starting with at least 100GB for the portal if you are around 1M sessions. You can get away with less for a game around 100K CCU, and this is best discovered through synthetic load testing in the staging environment.

## 3. Right size your relay backend

Depending on the number of relays you have, you can adjust the scale of the relay backend. By default it is set to `c3-highcpu-44` instance type, which is powerful enough to perform route optimization easily with 1000 relays. Typical relay fleets usually only have a few hundred relays, so you may be able to reduce the power of the relay backend to save money.

IMPORTANT: The relay backend does all its work in one VM instance. Additional VMs in the instance group just add availability. So the instance type of the VM in this group is how you scale it up and down.

Search for "relay_backend" in `~/next/terraform/prod/backend/main.tf`:

You will find the definition for the relay backend service:

```
module "relay_backend" {

  source = "../../modules/internal_http_service"

  service_name = "relay-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a relay_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    ENABLE_RELAY_HISTORY=true
    GOOGLE_PROJECT_ID=${local.google_project_id}
    REDIS_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    DATABASE_URL="${var.google_database_bucket}/prod.bin"
    DATABASE_PATH="/app/database.bin"
    INITIAL_DELAY=420
    MAX_JITTER=2
    MAX_PACKET_LOSS=0.1
    ENABLE_GOOGLE_PUBSUB=true
    ENABLE_REDIS_TIME_SERIES=true
    REDIS_TIME_SERIES_HOSTNAME="${module.redis_time_series.address}:6379"
    REDIS_PORTAL_CLUSTER="${local.redis_portal_address}"
    EOF
    sudo gsutil cp ${var.google_database_bucket}/prod.bin /app/database.bin
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "c3-highcpu-44"
  project                    = local.google_project_id
  region                     = var.google_region
  zones                      = var.google_zones
  default_network            = google_compute_network.production.id
  default_subnetwork         = google_compute_subnetwork.production.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = local.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-http"]
  target_size                = 3
  tier_1                     = true

  depends_on = [
    google_pubsub_topic.pubsub_topic, 
    google_pubsub_subscription.pubsub_subscription,
    google_redis_instance.redis_relay_backend,
    module.redis_time_series
  ]
}
```

The key things you can change here are:

* **machine_type** - set a less powerful machine type, if you have fewer than 1000 relays.
* **tier_1** - tier 1 bandwidth helps improve IO performance in google cloud, but if you use a machine lighter than the c3-highcpu-44 you'll need to turn it off, because tier_1 bandwidth is only available on heavyweight, high CPU instance types.
* **target_size** - this is the number of VMs in the instance group. It's currently set to 3 for maximum availability (there are three availability zones in Google's us-central-1 region that suuport c3 instance types). You could reduce this to 1 and save on hosting costs significantly, at the cost of reduced availability.

I strongly recommend load testing any changes in the staging environment before pushing to production. Critically, you must make sure that the relay backend has enough CPU to perform the route optimization (see the "Optimize" graph under "Admin", this should stay solidly under 1 second), while ensuring that your relay backend does not get IO bound processing the updates from each relay. If the relay backend gets IO bound, you'll notice randomly that certain relays as you add them get excluded from routing for no reason except that your relay backend is overloaded.

## 4. Right size your server backend

The server backend is the component that monitors session network performance, decides if players should be accelerated or not, and selects the best route.

The current service backend as configured is sufficient to scale to 1M peak CCU without any changes. It's a stateless managed instance group and it is able to scale horizontally by adding more VM instances to the group.

If you have fewer than 1M CCU for your game, you can save money by right sizing the server backend.

Search for "server backend" in `~/next/terraform/prod/backend/main.tf`:

You will find the definition for the server backend instance group:

```
module "server_backend" {

  source = "../../modules/external_udp_service_autoscale"

  service_name = "server-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    sudo ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a server_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=prod
    UDP_PORT=40000
    UDP_BIND_ADDRESS="##########:40000"
    UDP_NUM_THREADS=8
    UDP_SOCKET_READ_BUFFER=104857600
    UDP_SOCKET_WRITE_BUFFER=104857600
    GOOGLE_PROJECT_ID=${local.google_project_id}
    MAGIC_URL="http://${module.magic_backend.address}/magic"
    REDIS_CLUSTER="${local.redis_portal_address}"
    RELAY_BACKEND_PUBLIC_KEY=${var.relay_backend_public_key}
    RELAY_BACKEND_PRIVATE_KEY=${local.relay_backend_private_key}
    SERVER_BACKEND_ADDRESS="##########:40000"
    SERVER_BACKEND_PUBLIC_KEY=${var.server_backend_public_key}
    SERVER_BACKEND_PRIVATE_KEY=${local.server_backend_private_key}
    ROUTE_MATRIX_URL="http://${module.relay_backend.address}/route_matrix"
    PING_KEY=${local.ping_key}
    IP2LOCATION_BUCKET_NAME=${var.ip2location_bucket_name}
    ENABLE_GOOGLE_PUBSUB=true
    ENABLE_REDIS_TIME_SERIES=true
    REDIS_TIME_SERIES_HOSTNAME="${module.redis_time_series.address}:6379"
    REDIS_PORTAL_CLUSTER="${local.redis_portal_address}"
    REDIS_RELAY_BACKEND_HOSTNAME="${google_redis_instance.redis_relay_backend.host}:6379"
    SESSION_CRUNCHER_URL="http://${module.session_cruncher.address}"
    SERVER_CRUNCHER_URL="http://${module.server_cruncher.address}"
    PORTAL_NEXT_SESSIONS_ONLY=false
    ENABLE_IP2LOCATION=true
    EOF
    sudo systemctl start app.service
  EOF1

  tag                        = var.tag
  extra                      = var.extra
  machine_type               = "c3-highcpu-44"
  project                    = local.google_project_id
  region                     = var.google_region
  zones                      = var.google_zones
  port                       = 40000
  default_network            = google_compute_network.production.id
  default_subnetwork         = google_compute_subnetwork.production.id
  load_balancer_subnetwork   = google_compute_subnetwork.internal_http_load_balancer.id
  load_balancer_network_mask = google_compute_subnetwork.internal_http_load_balancer.ip_cidr_range
  service_account            = local.google_service_account
  tags                       = ["allow-ssh", "allow-health-checks", "allow-udp-40000"]
  min_size                   = 3
  max_size                   = 64
  target_cpu                 = 30

  depends_on = [
    google_pubsub_topic.pubsub_topic, 
    google_pubsub_subscription.pubsub_subscription,
    google_redis_cluster.portal
  ]
}
```

The safest option is to just leave this alone, or to simply reduce `min_size` from 3 to 1 instances at the cost of reduced availability. This will change the server backend so it runs in only one availability zone in google cloud, insteaod of three, and divide the cost it takes to run the server backend by 3.

If you are more on the 100k CCU range than million+, then you could try adjustments to bring the instance type down to `n1-standard-8` for reduced cost per-instance, but at the same time you must also reduce the number of UDP threads per-instance with the environment var `UDP_NUM_THREADS=8` from 8 to 2, otherwise the instance type won't have enough CPU power to process all the packets it receives. I strongly recommend for 1M+ CCU or close to it, that you stay with the tried and tested c3-highcpu-44 configuration. In load testing I've found it to be superior to n1-standard-8 based instances at high player counts.

When making any changes to the server backend it is _vital_ to load test your changes in the staging environment prior to deploying to production. It is very easy to get the server backend into a state where it is IO bound, or CPU bound and cannot process requests from the SDK quickly enough, leading to "Retries" and "Fallback to Direct" in the "Admin" page, indicating that player sessions have given up trying to get acceleration because they did not get a response from the server backend quickly enough.

Finally, make sure to conservatively set `min_size` such that you know you can sustain your base load level. Even with correct tuning for the server backend service, if a stampede of players comes at your backend, they will overload the backend before it has time to scale up (you'll see retries and fallbacks to direct under "Admin" in the portal), unless you pre-scale with `min_size`. It generally takes 1-2 minutes for each server backend instance to start and come online, and this is just too slow for a rapid scale up of 100k+ players or 1M+ players. The system will recover, and the fallback to direct is only temporary during scale up, but I always like to see it running clean without retries and fallback to direct where possible for the best player experience.

## 5. If in doubt get help!

I'm always happy to help. With a support contract I can help you right size your production backend and save money. Just email me at glenn@networknext.com. I'm the inventor and author of Network Next.

[Back to main documentation](../README.md)
