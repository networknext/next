![Network Next](https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg)

# Network Next Operator Tool

Run `./next` in the root of the repo for available commands.

## Env

To set the current environment for the next tool: `next env <local|dev|prod>`

Only valid values are `local`, `dev`, or `prod`.

## Auth

To authorize with the currently selected environment: `next auth`

## Buyers

To list buyers: `next buyers`

To add a buyer: `next buyers add [filepath]`

Adds a buyer to Firestore based on the given JSON file. You can also pipe in the JSON data (ex. `cat buyer.json | next buyers add`)
To see an example buyer JSON schema, use `next buyers add example`

To remove a buyer: `next buyers remove <id>`

Removes a buyer with the given buyer ID from Firestore. The buyer ID can be found with `next buyers`

## Sellers

To list sellers: `next sellers`

To add a seller: `next sellers add [filepath]`

Adds a seller to Firestore based on the given JSON file. You can also pipe in the JSON data (ex. `cat seller.json | next sellers add`)
To see an example seller JSON schema, use `next sellers add example`

To remove a seller: `next sellers remove <id>`

Removes a seller with the given seller ID from Firestore. The seller ID can be found with `next sellers`

## Datacenters

To list datacenters: `next datacenters`

To add a datacenter: `next datacenters add [filepath]`

Adds a datacenter to Firestore based on the given JSON file. You can also pipe in the JSON data (ex. `cat datacenter.json | next datacenters add`)
To see an example datacenter JSON schema, use `next datacenters add example`

To remove a datacenter: `next datacenters remove <name>`

Removes a datacenter with the given datacenter name from Firestore.

## Relays

To list relays: `next relays [optional regex]`

To add a relay: `next relays add [filepath]`

Adds a relay to Firestore based on the given JSON file. You can also pipe in the JSON data (ex. `cat relay.json | next relays add`)
To see an example relay JSON schema, use `next relays add example`

To remove a relay: `next relays remove <name>`

Removes a relay with the given relay name from Firestore.

### Check

To check relays: `next relays check [filter]`

Displays a table of filtered relays, or all relays if no filter supplied, containing helpful information.

The table will display if the relay can be ssh'ed into (yes/no), the version of Ubuntu it is running, and how many logical CPU cores are available. In the case of hyperthreaded CPUs this will be twice the number of physical cores.

To test if the relay can be ssh'ed into, it uses ssh's `-o` flag with `ConnectTimeout=60`. This means the tool will report false if the relay is unreachable for 60 seconds. If the relay is reachable and will take longer than 60 seconds, the tool will continue. So for very distant relays, and especially those with high packet loss, this tool may take 2 or 3 minutes to complete.


## Route Shaders

To get a route shader: `next routeshader <buyer ID>`

Gets a route shader for a given buyer ID. You can find a buyer's ID with `next buyers`.

To set or update a route shader: `next routeshader set <buyer ID> [filepath]`

Sets a route shader in Firestore for a given buyer based on the given JSON file. You can also pipe in the JSON data (ex. `cat relay.json | next routeshader set <buyer ID>`)
To see an example route shader JSON schema, use `next routeshader set example`

## SSH

To SSH: `next ssh [identifier]`

SSH into a remote device. You must set the SSH key before attempting to connect to a device, otherwise you will get denied.

To set the SSH key: `next ssh key [path to key file]`

- You can't use '~' in the path directly, it must be expanded by the shell first. Or in other words don't quote the argument

## Relay

[Enable](#Enable), [Disable](#Disable), [Update](#Update), and [Revert](#Revert) match sellers as well as relay names.

To do so reliably the regex must be of the form of `^seller_name$` otherwise if any relay names also happen to match the seller's name, they will appear in the list too.

### Enable

To Enable a relay: `next relay enable [regex]...`

The tool will SSH into the relays that match the regex(s), start the relay service, and set the state to offline. If the service is already running the command will only update the state to offline.

Once the relay initializes with the backend the state will be changed to enabled.

### Disable

To Disable a relay: `next relay disable [regex]...`

First the tool will update the matching relays' states in Firestore to the Disabled state. Then it will SSH into the relays and stop their services. If the service is already stopped the tool will do nothing aside from setting the state.

### Update

To Update a relay: `next relay update [regex]...`

The tool will perform several actions to update relays matching the supplied regex(s).

Before updating make sure you have the desired environment set via the [relay env](#Env) setting.

Also if any updates fail throughout the process the program will quit and not continue to the next.

For each matching relay the tool will:
- download the latest version from the GCP bucket and untar it in the `dist` directory.
- disable the relay using the [disable](#Disable) functionality.
- generate a new set of public and private keys for the relay and set them both in the relay's environment file and publish the changes to Firestore as well.
- update the state of the relay to offline, much the same way the [enable](#Enable) command does it. Once the relay initializes with the backend the state will change to enabled.

### Revert

To Revert a relay: `next relay revert [regex]...`

The tool will revert a relay back to the the previous version. It will remove the binary and associated files that are currently active and restore the most recent backup.

However it will not restore the previous public key in firestore as that is lost upon each update.

### Set NIC Speed

To set the NIC speed of a relay: `next relay setnic <relay name> <value (Mbps)>`

This will adjust the NIC speed field in Firestore for the given relay to the provided value in Mbps.

### Keys

To see the public and update keys of a relay: `next relay keys <relay name>`

Shows the public key and update key for the given relay.

### Cost

Downloads the current cost matrix from the relay backend to 'cost.bin' or the first argument supplied.

### Optimize

Optimizes the downloaded cost matrix into a route matrix.

The first argument is the rtt threshold which defaults to 1. \
The second argument is the path of the cost matrix if you don't have it saved as 'cost.bin'. \
The third argument is the path you want the route matrix written to. Default is 'optimize.bin'.

### Analyze

Analyzes the route matrix and produces high level data. The file read in is 'optimize.bin' or the first argument supplied.
