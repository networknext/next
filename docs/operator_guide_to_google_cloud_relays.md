<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Operator guide to Google Cloud relays

This section describes how to use the Network Next terraform provider together with the Google Cloud terraform provider to spin up Google Cloud relays. It sure beats doing it manually!

## The google config tool

The purpose of the google config tool is to extract configuration data from your Google Cloud account and get it into a form where it can be used in terraform to create relays.

The google config tool is run automatically when you run `next config` along with config for other sellers, but if you want to run the google config by itself, just go:

```
run config-google
```

You can see the source code for the google cloud config tool in `~/next/sellers/google.go`. 

When it runs, it will cache some data under `~/next/cache` to speed up its operation, and it will generate `~/next/sellers/google/generated.tf` and `~/next/config/google.txt`.

The `google/generated.tf` file is how we inject the set of google cloud datacenters into terraform. 

The google config tool interacts with your Google Cloud account via REST API and queries data such as the set of regions, and the zones within each region, and maps them to network next datacenter names like "google.iowa.1".

The `config/google.txt` file is uploaded to google cloud storage via semaphore "Upload Config" job, and is read by the SDK to perform autodetection of the Google Cloud datacenter your server is running in. In short, this text file is just a mapping from Google Cloud native zone name to the network next datacenter name.

## 2. Adding new datacenters to google cloud

When you run the google config tool via `run config-google`, you will see an output describing the set of datacenters in google cloud:

```console
gaffer@batman next % run config-google

Known datacenters:

  google.taiwan.1
  google.taiwan.2
  google.taiwan.3
  google.hongkong.1
  google.hongkong.2
  google.hongkong.3
  google.tokyo.1
  google.tokyo.2
  google.tokyo.3
  google.osaka.1
  google.osaka.2
  google.osaka.3
  google.seoul.1
  google.seoul.2
  google.seoul.3
  google.mumbai.1
  google.mumbai.2
  google.mumbai.3
  google.delhi.1
  google.delhi.2
  google.delhi.3
  google.singapore.1
  google.singapore.2
  google.singapore.3
  google.jakarta.1
  google.jakarta.2
  google.jakarta.3
  google.sydney.1
  google.sydney.2
  google.sydney.3
  google.melbourne.1
  google.melbourne.2
  google.melbourne.3
  google.warsaw.1
  google.warsaw.2
  google.warsaw.3
  google.finland.1
  google.finland.2
  google.finland.3
  google.madrid.1
  google.madrid.2
  google.madrid.3
  google.belgium.2
  google.belgium.3
  google.belgium.4
  google.london.1
  google.london.2
  google.london.3
  google.frankfurt.1
  google.frankfurt.2
  google.frankfurt.3
  google.netherlands.1
  google.netherlands.2
  google.netherlands.3
  google.zurich.1
  google.zurich.2
  google.zurich.3
  google.milan.1
  google.milan.2
  google.milan.3
  google.paris.1
  google.paris.2
  google.paris.3
  google.telaviv.1
  google.telaviv.2
  google.telaviv.3
  google.montreal.1
  google.montreal.2
  google.montreal.3
  google.toronto.1
  google.toronto.2
  google.toronto.3
  google.saopaulo.1
  google.saopaulo.2
  google.saopaulo.3
  google.santiago.1
  google.santiago.2
  google.santiago.3
  google.iowa.1
  google.iowa.2
  google.iowa.3
  google.iowa.6
  google.southcarolina.2
  google.southcarolina.3
  google.southcarolina.4
  google.virginia.1
  google.virginia.2
  google.virginia.3
  google.ohio.1
  google.ohio.2
  google.ohio.3
  google.dallas.1
  google.dallas.2
  google.dallas.3
  google.oregon.1
  google.oregon.2
  google.oregon.3
  google.losangeles.1
  google.losangeles.2
  google.losangeles.3
  google.saltlakecity.1
  google.saltlakecity.2
  google.saltlakecity.3
  google.lasvegas.1
  google.lasvegas.2
  google.lasvegas.3

Unknown datacenters:

  europe-west10-a
  europe-west10-b
  europe-west10-c
  europe-west12-a
  europe-west12-b
  europe-west12-c
  me-central1-a
  me-central1-b
  me-central1-c
  me-central2-a
  me-central2-b
  me-central2-c
```

Here we can see that most google datacenters are already mapped, but there are some new, unknown google datacenters.

These are new regions enabled in google cloud, that are not yet mapped to Network Next datacenter names.

To fix this, you would just go into sellers/google.go and modify the datacenter map to add these new regions:

```
// This definition drives the set of google datacenters, eg. "google.[country/city].[number]"

var datacenterMap = map[string]*Datacenter{

	"us-east1":                {"southcarolina", 33.8361, -81.1637},
	"asia-east1":              {"taiwan", 25.105497, 121.597366},
	"asia-east2":              {"hongkong", 22.3193, 114.1694},
	"asia-northeast1":         {"tokyo", 35.6762, 139.6503},
	"asia-northeast2":         {"osaka", 34.6937, 135.5023},
	"asia-northeast3":         {"seoul", 37.5665, 126.9780},
	"asia-south1":             {"mumbai", 19.0760, 72.8777},
	"asia-south2":             {"delhi", 28.7041, 77.1025},
	"asia-southeast1":         {"singapore", 1.3521, 103.8198},
	"asia-southeast2":         {"jakarta", 6.2088, 106.8456},
	"australia-southeast1":    {"sydney", -33.8688, 151.2093},
	"australia-southeast2":    {"melbourne", -37.8136, 144.9631},
	"europe-central2":         {"warsaw", 52.2297, 21.0122},
	"europe-north1":           {"finland", 60.5693, 27.1878},
	"europe-southwest1":       {"madrid", 40.4168, 3.7038},
	"europe-west1":            {"belgium", 50.4706, 3.8170},
	"europe-west2":            {"london", 51.5072, -0.1276},
	"europe-west3":            {"frankfurt", 50.1109, 8.6821},
	"europe-west4":            {"netherlands", 53.4386, 6.8355},
	"europe-west6":            {"zurich", 47.3769, 8.5417},
	"europe-west8":            {"milan", 45.4642, 9.1900},
	"europe-west9":            {"paris", 48.8566, 2.3522},
	"me-west1":                {"telaviv", 32.0853, 34.7818},
	"northamerica-northeast1": {"montreal", 45.5019, -73.5674},
	"northamerica-northeast2": {"toronto", 43.6532, -79.3832},
	"southamerica-east1":      {"saopaulo", -23.5558, -46.6396},
	"southamerica-west1":      {"santiago", -33.4489, -70.6693},
	"us-central1":             {"iowa", 41.8780, -93.0977},
	"us-east4":                {"virginia", 37.4316, -78.6569},
	"us-east5":                {"ohio", 39.9612, -82.9988},
	"us-south1":               {"dallas", 32.7767, -96.7970},
	"us-west1":                {"oregon", 45.5946, -121.1787},
	"us-west2":                {"losangeles", 34.0522, -118.2437},
	"us-west3":                {"saltlakecity", 40.7608, -111.8910},
	"us-west4":                {"lasvegas", 36.1716, -115.1391},
}
```

For example, you would map "europe-west10" to "berlin" and give it the correct latitude and longitude of (52.5200, 13.4050). 

Same for "europe-west12" and "me-central1" and "me-central2", just look up what cities they are in, and set their lat long to an approx location for each city. It doesn't need to be precise.

Please make sure to follow [naming conventions](relay_and_datacenter_naming_conventions.md) when you add new google datacenters.

## Update datacenter autodetect system post-new datacenters in google cloud

After you add new google datacenters, run the google config tool again and check the changes made with diff.

```
next config-google
git diff
```

You should see the new datacenters added to `sellers/google/generated.tf` file and `config/google.txt`

Check these changes in, then inside semaphore CI, trigger "Upload Config" job on your most recent commit.

This uploads the config/google.txt file to Google Cloud storage, where the SDK will pick it up, and the new datacenters will be available for datacenter autodetection.

_(Datacenter autodetect lets you simply pass in "cloud" when your server runs in AWS or Google Cloud, and the SDK will work out which datacenter it is located in automatically. Saves a lot of time.)_

## Update database with your new datacenters

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

## Spinning up relays in Google Cloud

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
