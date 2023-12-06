<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Operator guide to AWS relays

This section describes how to use the Network Next terraform provider together with the Amazon terraform provider to spin up AWS relays.

## 1. The amazon config tool

The purpose of the amazon config tool is to extract configuration data from your AWS account and get it into a form where it can be used in terraform to create relays in AWS.

The amazon config tool is run automatically when you run `next config` along with config for other sellers, but if you want to run the amazon config by itself, just go:

```
run config-amazon
```

The amazon config tool lives in `~/next/sellers/amazon.go`. 

Because the architecture of AWS is _heavily_ region-based, combined with some limitations in the terraform config language, it's just not possible to programmatically build up all the multi-region resources required to make relays work in terraform script alone. The end result is that you need to describe the set of dev and prod AWS relays inside a data structure inside the amazon config tool itself, and it generates the terraform script required to create them.

The amazon provider is also complicated by the fact that AWS zone ids are _account specific_. This means that us-east-1a in my account is probably not the same availability zone as us-east-1a in your account. You can read more about this here: https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html

It complicated yet again by the fact that AWS has this weird (but cool) local zone thing, where many datacenters you want really want access to have to be manually enabled in your account, and are sort of piggy backed off some parent geographically unrelated parent region like us-east-1 (virginia). You can read more about this here: https://aws.amazon.com/about-aws/global-infrastructure/localzones/locations/

When the amazon config tool runs, it caches data under `~/next/cache` to speed up its operation next time it runs. It generates `~/next/terraform/dev/relays/amazon/generated.tf`, `~/next/terraform/prod/relays/amazon/generated.tf` and `~/next/config/amazon.txt`.

The `terraform/[dev|prod]/relays/amazon/generated.tf` files contains not just the definition of all amazon datacenters in Network Next, but also a huuuuge wad of generated code to do the multi-region dance in AWS for relays, plus code to actually create the relays in the env. Separate files are generated for dev and prod envs.

The `config/amazon.txt` file is used for datacenter autodetection in AWS. It's uploaded to google cloud storage via semaphore "Upload Config" job, and is read by the SDK to perform autodetection of the AWS datacenter your server is running in. This text file is basically just a mapping from the AWS AZID to the network next datacenter name.

## 2. Adding new datacenters in AWS

When you run the amazon config tool via `run config-amazon`, you will see an output describing the set of datacenters in AWS:

```console
gaffer@batman next % run config-amazon

Known datacenters:

  amazon.ohio.2
  amazon.ohio.3
  amazon.sanjose.1
  amazon.sanjose.3
  amazon.oregon.1
  amazon.oregon.2
  amazon.oregon.3
  amazon.oregon.4
  amazon.denver.1
  amazon.lasvegas.1
  amazon.lasvegas.2
  amazon.losangeles.1
  amazon.losangeles.2
  amazon.portland.1
  amazon.phoenix.1
  amazon.seattle.1
  etc...

Unknown datacenters:

  ap-southeast-1-mnl-1a -> apse1-mnl1-az1
  ap-southeast-2-akl-1a -> apse2-akl1-az1
  us-west-2-phx-2a -> usw2-phx2-az1

Excluded regions:

  il-central-1

Generating amazon.txt

Generating dev amazon/generated.tf

Generating prod amazon/generated.tf
```

Here we can see that there are some unknown datacenters, and some excluded regions.

An excluded region means that in your AWS account, that region is not activated yet. It may not be generally available, or you have to do some steps in the AWS console to request that region be enabled. If you need to use this region, enable it in your AWS account and it should no longer be excluded the next time you run the amazon config tool.

The unknown datacenters part means that there are some datacenters available in AWS that are not mapped to Network Next datacenters yet.

To fix this, you would just go into sellers/amazon.go and modify the datacenter map to add these new regions:

```
// Amazon datacenters, eg. "amazon.[country/city].[number]"

var datacenterMap = map[string]*Datacenter{

	// regions (AZID)

	"afs1":  {"johannesburg", -33.9249, 18.4241},
	"ape1":  {"hongkong", 22.3193, 114.1694},
	"apne1": {"tokyo", 35.6762, 139.6503},
	"apne2": {"seoul", 37.5665, 126.9780},
	"apne3": {"osaka", 34.6937, 135.5023},
	"aps1":  {"mumbai", 19.0760, 72.8777},
	"aps2":  {"hyderabad", 17.3850, 78.4867},
	"apse1": {"singapore", 1.3521, 103.8198},
	"apse2": {"sydney", -33.8688, 151.2093},
	"apse3": {"jakarta", -6.2088, 106.8456},
	"apse4": {"melbourne", -37.8136, 144.9631},
	"cac1":  {"montreal", 45.5019, -73.5674},
	"euc1":  {"frankfurt", 50.1109, 8.6821},
	"euc2":  {"zurich", 47.3769, 8.5417},
	"eun1":  {"stockholm", 59.3293, 18.0686},
	"eus1":  {"milan", 45.4642, 9.1900},
	"eus2":  {"spain", 41.5976, -0.9057},
	"euw1":  {"ireland", 53.7798, -7.3055},
	"euw2":  {"london", 51.5072, -0.1276},
	"euw3":  {"paris", 48.8566, 2.3522},
	"mec1":  {"uae", 23.4241, 53.8478},
	"mes1":  {"bahrain", 26.0667, 50.5577},
	"sae1":  {"saopaulo", -23.5558, -46.6396},
	"use1":  {"virginia", 39.0438, -77.4874},
	"use2":  {"ohio", 40.4173, -82.9071},
	"usw1":  {"sanjose", 37.3387, -121.8853},
	"usw2":  {"oregon", 45.8399, -119.7006},

	// local zones (AZID)

	"los1": {"nigeria", 6.5244, 3.3792},
	"tpe1": {"taipei", 25.0330, 121.5654},
	"ccu1": {"kolkata", 22.5726, 88.3639},
	"del1": {"delhi", 28.7041, 77.1025},
	"bkk1": {"bangkok", 13.7563, 100.5018},
	"per1": {"perth", -31.9523, 115.8613},
	"ham1": {"hamburg", 53.5488, 9.9872},
	"waw1": {"warsaw", 52.2297, 21.0122},
	"cph1": {"copenhagen", 55.6761, 12.5683},
	"hel1": {"finland", 60.1699, 24.9384},
	"mct1": {"oman", 23.5880, 58.3829},
	"atl1": {"atlanta", 33.7488, -84.3877},
	"bos1": {"boston", 42.3601, -71.0589},
	"bue1": {"buenosaires", -34.6037, -58.3816},
	"chi1": {"chicago", 41.8781, -87.6298},
	"dfw1": {"dallas", 32.7767, -96.7970},
	"iah1": {"houston", 29.7604, -95.3698},
	"lim1": {"lima", -12.0464, -77.0428},
	"mci1": {"kansas", 39.0997, -94.5786},
	"mia1": {"miami", 25.7617, -80.1918},
	"msp1": {"minneapolis", 44.9778, -93.2650},
	"nyc1": {"newyork", 40.7128, -74.0060},
	"phl1": {"philadelphia", 39.9526, -75.1652},
	"qro1": {"mexico", 23.6345, -102.5528},
	"scl1": {"santiago", -33.4489, -70.6693},
	"den1": {"denver", 39.7392, -104.9903},
	"las1": {"lasvegas", 36.1716, -115.1391},
	"lax1": {"losangeles", 34.0522, -118.2437},
	"pdx1": {"portland", 45.5152, -122.6784},
	"phx1": {"phoenix", 33.4484, -112.0740},
	"sea1": {"seattle", 47.6062, -122.3321},
}
```

Please make sure to follow [naming conventions](datacenter_and_relay_naming_conventions.md) when you add new amazon datacenters.


















------------------------------

## 3. Update datacenter autodetect system

After you add new google datacenters, run the google config tool again and check the changes made with diff.

```
next config-google
git diff
```

You should see the new datacenters added to `sellers/google/generated.tf` file and `config/google.txt`

Check these changes in, then inside semaphore CI, trigger "Upload Config" job on your most recent commit.

This uploads the config/google.txt file to Google Cloud storage, where the SDK will pick it up, and the new datacenters will be available for datacenter autodetection.

_(Datacenter autodetect lets you simply pass in "cloud" when your server runs in AWS or Google Cloud, and the SDK will work out which datacenter it is located in automatically. Saves a lot of time.)_

## 4. Update database with your new datacenters

Depending on your environment you are changing, change into either `~/next/terraform/dev/relays` or `~/next/terraform/prod/relays`.

For example, for dev:

```console
cd ~/next/terraform/dev/relays
terraform apply
```

Once this completes, it will have mutated your Postgres SQL instance in your Network Next env to add the new google cloud datacenters.

## Commit updated database.bin to the backend runtime

The backend runtime does not directly talk with Postgres SQL. Instead they get their configuration from a database.bin file, which is an extracted version of the configuration data stored in Postgres.

To make your Postgres SQL changes active in the backend, you must extract this database.bin and commit it to the backend, after you make changes to Postgres.

For example, for dev:

```console
cd ~/next
next select dev
next database
next commit
```

It takes up to 60 seconds for the runtime backend to pick up your committed database.bin.

After this point, you should be able to load up the portal and see the new datacenters you added for Google Cloud.

## 5. Spin up relays in Google Cloud

It's ridiculously easy! Take a look at `~/terraform/backend/dev/relays/main.tf` or `~/terraform/backend/prod/relays/main.tf`, depending on which environment you want to change.

For example in dev, you can see:

```
# =============
# GOOGLE RELAYS
# =============

locals {

  google_credentials = "~/secrets/terraform-dev-relays.json"
  google_project     = file("../../projects/dev-relays-project-id.txt")
  google_relays = {

    "google.iowa.1" = {
      datacenter_name = "google.iowa.1"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    "google.iowa.2" = {
      datacenter_name = "google.iowa.2"
      type            = "n1-standard-2"
      image           = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    },

    etc...
  }
```

Addding a new relay is as simple as copying and pasting an entry for a new relay and updating its relay name and datacenter name, and then running `terraform apply`.

Once terraform has completed, remember that you must once again commit the database.bin to the backend runtime for your changes to take effect:

For example:

```console
cd ~/next/terraform/dev/relays
terraform apply
cd ~/next
next select dev
next database
next commit
```

Once the database is committed, you then need to connect to your VPN (cannot SSH into relays except from your VPN address), then setup the new relays:

```console
next setup google
```

This loads the relay service on all google relays in your system, skipping over any that are already setup.

Once the setup is complete, you can check your google relays are online with:

```console
gaffer@batman next % next relays google

┌──────────────────────┬──────────────────────┬──────────────────┬────────┬────────┬──────────┬───────────────────┐
│ Name                 │ PublicAddress        │ Id               │ Status │ Uptime │ Sessions │ Version           │
├──────────────────────┼──────────────────────┼──────────────────┼────────┼────────┼──────────┼───────────────────┤
│ google.iowa.3        │ 34.42.110.106:40000  │ 380e3da4fc2ddd77 │ online │ 20h    │ 87       │ relay-debug-1.0.0 │
│ google.iowa.2        │ 34.29.81.36:40000    │ a7f626db601b36ff │ online │ 19h    │ 81       │ relay-debug-1.0.0 │
│ google.iowa.1        │ 34.173.141.155:40000 │ a970f7ebafaa5d0e │ online │ 20h    │ 61       │ relay-debug-1.0.0 │
│ google.iowa.6        │ 34.16.106.87:40000   │ adbc009b12fe54d5 │ online │ 20h    │ 29       │ relay-debug-1.0.0 │
│ google.virginia.3    │ 34.150.187.253:40000 │ e9ace494be91ced8 │ online │ 20h    │ 17       │ relay-debug-1.0.0 │
│ google.ohio.1        │ 34.162.195.174:40000 │ b0cbb9243436b5d8 │ online │ 20h    │ 13       │ relay-debug-1.0.0 │
│ google.virginia.2    │ 34.48.63.170:40000   │ 3b3438bd62d46659 │ online │ 20h    │ 11       │ relay-debug-1.0.0 │
│ google.ohio.2        │ 34.162.91.234:40000  │ d343e8a0f6ab8214 │ online │ 20h    │ 9        │ relay-debug-1.0.0 │
│ google.ohio.3        │ 34.162.149.251:40000 │ fcb14430d000581d │ online │ 20h    │ 8        │ relay-debug-1.0.0 │
│ google.virginia.1    │ 34.48.61.240:40000   │ 7a9169eb2b715499 │ online │ 20h    │ 8        │ relay-debug-1.0.0 │
│ google.finland.1     │ 34.88.111.165:40000  │ b6335a734e81dcc1 │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.finland.2     │ 34.88.178.155:40000  │ c32846ba731949cf │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.finland.3     │ 34.88.153.153:40000  │ 54f6418d0d6d54ce │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.frankfurt.1   │ 34.159.195.194:40000 │ ef43099960e2ee1c │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.frankfurt.2   │ 34.159.181.85:40000  │ 7757125bfbdc13e  │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.frankfurt.3   │ 34.159.230.86:40000  │ 9ad5f7ebf2e1f178 │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.london.1      │ 35.242.181.221:40000 │ 940d78fdf3b5393a │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.london.2      │ 34.89.51.164:40000   │ 2205b2e7cbf53ce0 │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.london.3      │ 34.147.219.214:40000 │ c6fb8a1814e33b23 │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.netherlands.1 │ 34.90.255.68:40000   │ a81985a3307974df │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.netherlands.2 │ 34.90.39.151:40000   │ 220db5ee6a669ef4 │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
│ google.netherlands.3 │ 34.141.240.124:40000 │ 8f7b7e5c773e47fd │ online │ 20h    │ 0        │ relay-debug-1.0.0 │
└──────────────────────┴──────────────────────┴──────────────────┴────────┴────────┴──────────┴───────────────────┘
```

You can also go to the portal and you should see your new relays there as well.

[Back to main documentation](../README.md)
