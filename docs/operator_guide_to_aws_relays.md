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

## 3. Update datacenter autodetect system

After you add new amazon datacenters, run the amazon config tool again and check the changes made with diff.

```
next config-amazon
git diff
```

You should see the new datacenters added to `sellers/amazon/generated.tf` file and `config/amazon.txt`

Check these changes in, then inside semaphore CI, trigger "Upload Config" job on your most recent commit.

This uploads the config/amazon.txt file to Google Cloud storage, where the SDK will pick it up, and the new datacenters will be available for datacenter autodetection.

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

After this point, you should be able to load up the portal and see the new datacenters you added for AWS.

## 5. Spin up relays in AWS.

It's ridiculously easy! Take a look at `~/next/sellers/amazon.go`.

There are two data structures in here. One for dev relays, and one for prod relays:

```
// DEV RELAYS

var devRelayMap = map[string][]string{
	"amazon.virginia.1": {"amazon.virginia.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.virginia.2": {"amazon.virginia.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.ohio.1":     {"amazon.ohio.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.ohio.2":     {"amazon.ohio.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.stockholm.1":  {"amazon.stockholm.1", "m5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.stockholm.2":  {"amazon.stockholm.2", "m5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.stockholm.3":  {"amazon.stockholm.3", "m5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
}

// PROD RELAYS

var prodRelayMap = map[string][]string{
	"amazon.virginia.1": {"amazon.virginia.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.virginia.2": {"amazon.virginia.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.ohio.1":     {"amazon.ohio.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.ohio.2":     {"amazon.ohio.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.oregon.1":   {"amazon.oregon.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.sanjose.1":  {"amazon.sanjose.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
}
```

To add a new relay just create a new entry in the map for your relay. You'll need to select the correct instance type you want, which can vary depending on the datacenter you are picking.

You can see a list of instance types here: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html#AvailableInstanceTypes

But you'll really want to use the AWS console or "aws" command to list the instance types available in the region you want to run a relay in.

Run the amazon config tool:

```console
run amazon-config
```

And it will generate the code to create the relays.

Then change into the terraform directory depending on env, for example in dev go:

```console
cd ~/next/terraform/dev/relays
terraform apply
```

It's common to have to iterate back and forth a bit, for example if the instance type is not available in the datacenter then AWS will error out here until you pick an instance type that is available.

Once terraform apply has completed successfully, remember that you must once again commit the database.bin to the backend runtime for your changes to take effect:

For example:

```console
cd ~/next
next select dev
next database
next commit
```

Once the database is committed, you then need to connect to your VPN (cannot SSH into relays except from your VPN address), then setup the new relays:

```console
next setup amazon
```

This loads the relay service on all amazon relays in your system, skipping over any that are already setup.

Once the setup is complete, you can check your amazon relays are online with:

```console
gaffer@batman next % next relays amazon

┌────────────────────┬──────────────────────┬──────────────────┬────────┬────────┬──────────┬───────────────────┐
│ Name               │ PublicAddress        │ Id               │ Status │ Uptime │ Sessions │ Version           │
├────────────────────┼──────────────────────┼──────────────────┼────────┼────────┼──────────┼───────────────────┤
│ amazon.virginia.1  │ 44.197.190.75:40000  │ 5f379f503bb1c95c │ online │ 1d     │ 17       │ relay-debug-1.0.0 │
│ amazon.ohio.2      │ 3.22.194.113:40000   │ ad1a43212a466bbb │ online │ 1d     │ 14       │ relay-debug-1.0.0 │
│ amazon.virginia.2  │ 44.204.77.144:40000  │ 2475ffad44ea2328 │ online │ 1d     │ 14       │ relay-debug-1.0.0 │
│ amazon.ohio.1      │ 18.220.135.169:40000 │ ace3bb374b852115 │ online │ 1d     │ 2        │ relay-debug-1.0.0 │
│ amazon.stockholm.1 │ 51.20.120.26:40000   │ 5d1df8b2f33beb0a │ online │ 1d     │ 0        │ relay-debug-1.0.0 │
│ amazon.stockholm.2 │ 16.171.149.90:40000  │ e76e968a7a1b3266 │ online │ 1d     │ 0        │ relay-debug-1.0.0 │
│ amazon.stockholm.3 │ 16.171.53.242:40000  │ 147b9a6c163a298f │ online │ 1d     │ 0        │ relay-debug-1.0.0 │
└────────────────────┴──────────────────────┴──────────────────┴────────┴────────┴──────────┴───────────────────┘```
```

You can also go to the portal and you should see your new relays there as well.

[Back to main documentation](../README.md)
