service_account      = "terraform@development-395218.iam.gserviceaccount.com"
credentials          = "~/secrets/terraform-development.json"
project              = "development-395218"
location             = "US"
region               = "us-central1"
zone                 = "us-central1-a"
artifacts_bucket     = "gs://masbandwidth_network_next_dev_artifacts"
machine_type         = "f1-micro"
git_hash             = "747"
vpn_address          = "45.79.157.168"

cloudflare_api_token              = "~/secrets/terraform-cloudflare.txt"
cloudflare_zone_id_api            = "eba5d882ea2aa23f92dfb50fbf7e3cf4" # -> virtualgo.net
cloudflare_zone_id_relay_backend  = "4690b11494ddc59f6fdf514780d24e0a" # -> losangelesfreewaysatnight.com
cloudflare_zone_id_server_backend = "d74db52c8b51a2b80a76f736713e3edd" # -> spacecats.net
