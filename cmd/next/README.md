![Network Next](https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg)

# Network Next Operator Tool

Run `./next` in the root of the repo for available commands.

Arguments surrounded by '[]' are optional. Those surrounded by '<>' are required.

## Select

To set the current environment for the next tool: `next select <local|dev|prod>`

Only valid values are `local`, `dev`, or `prod`.

## Env

To display information about the current environment: `next env`

The environment can also be changed with `next env [local|dev|prod]` in the same way as `next select <local|dev|prod>`
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

## Customers

To list customers: `next customers`

To add a customer: Either add a buyer or seller, or both. The customer root object will be added for you when appropriate automatically.

To remove a customer: Either remove a buyer or seller, or both. The customer root object will be remove for you when appropriate automatically.

To edit a customer's buyer link: `next customer link buyer <customer name> <new buyer ID>`
This will replace the customer's buyer reference with the buyer given by the buyer ID.
This link is normally handled for you, but this may be necessary to edit older existing customer data.

To edit a customer's seller link: `next customer link seller <customer name> <new seller ID>`
This will replace the customer's seller reference with the seller given by the seller ID.
This link is normally handled for you, but this may be necessary to edit older existing customer data.

## Datacenters

To list datacenters: `next datacenters`

To add a datacenter: `next datacenters add [filepath]`

Adds a datacenter to Firestore based on the given JSON file. You can also pipe in the JSON data (ex. `cat datacenter.json | next datacenters add`)
To see an example datacenter JSON schema, use `next datacenters add example`

To remove a datacenter: `next datacenters remove <name>`

Removes a datacenter with the given datacenter name from Firestore.

## Relays

To list relays: `next relays [regex]`

Use the following flags to filter the relays output:
`--enabled` only show enabled relays
`--noenabled` hide enabled relays
`--maintenance` only show relays in maintenane
`--nomaintenance` hide relays in maintenance
`--disabled` only show disabled relays
`--nodisabled` hide disabled relays
`--quarantined` only show quarantined relays
`--noquarantined` hide quarantined relays
`--decommissioned` only show decommissioned relays
`--nodecommissioned` hide decommissioned relays
`--offline` only show offline relays
`--nooffline` hide offline relays
`--down` only show relays that haven't pinged the backend in 30 seconds or more
`--all` show all relays (excluding decommissioned ones)

If no flags are provided, then the `--enabled`, `--quarantined`, and `--nodecommissioned` are set by default.
You can also use a combination of them as well. Ex `--enabled` and `--quaratined` will show all relays that are either enabled or quaratined.

To add a relay: `next relays add [filepath]`

Adds a relay to Firestore based on the given JSON file. You can also pipe in the JSON data (ex. `cat relay.json | next relays add`)
To see an example relay JSON schema, use `next relays add example`

To remove a relay: `next relays remove <name>`

Removes a relay with the given relay name from Firestore.

## Route Shaders

To get a route shader: `next shader <buyer ID>`

Gets a route shader for a given buyer ID. You can find a buyer's ID with `next buyers`.

To set or update a route shader: `next shader set <buyer ID> [filepath]`

Sets a route shader in Firestore for a given buyer based on the given JSON file. You can also pipe in the JSON data (ex. `cat relay.json | next shader set <buyer ID>`)
To see an example route shader JSON schema, use `next shader set example`

## SSH

To SSH: `next ssh <identifier>`

SSH into a remote device. You must set the [SSH key](#Key) before attempting to connect to a device, otherwise you will get denied.

### Key

To set the SSH key: `next ssh key <path to key file>`

- You can't use '~' in the path directly, it must be expanded by the shell first. Or in other words, don't quote the argument.

## Relay

[Enable](#Enable), [Disable](#Disable), [Update](#Update), and [Revert](#Revert) match sellers as well as relay names.

### Enable

To Enable a relay: `next relay enable [regex...]`

The tool will SSH into the relays that match the regex(s), start the relay service, and set the state to offline. If the service is already running the command will only update the state to offline.

Once the relay initializes with the backend the state will be changed to enabled.

### Disable

To Disable a relay: `next relay disable [regex...]`

First the tool will update the matching relays' states in Firestore to the Disabled state. Then it will SSH into the relays and stop their services. If the service is already stopped the tool will do nothing aside from setting the state.

### Update

To Update a relay: `next relay update [-cores n] [regex...]`

The tool will perform several actions to update relays matching the supplied regex(s).

If no arguments are supplied the tool will update relays whose state is enabled and version is not the latest.

If the `-cores` flag is provided followed by a number, it will set `RELAY_MAX_CORES` to that number. If this flag is omitted, then the relay will use all available cores. Generally you want to set this to half of the number of available cores for VMs and omit this flag for bare metal relays.

Before updating make sure you have the desired environment set via the [relay env](#Env) setting.

For each matching relay the tool will:
- download the latest version from the GCP bucket and untar it in the `dist` directory.
- disable the relay using the [disable](#Disable) functionality.
- generate a new set of public and private keys for the relay and set them both in the relay's environment file and publish the changes to Firestore as well.
- update the state of the relay to offline, much the same way the [enable](#Enable) command does it. Once the relay initializes with the backend the state will change to enabled.

#### Flags

`--cores`: Specify the number of cpu cores the relay should use. Typically set to half for VMs.
`--force`: Force the relay update regardless if the version is already the latest.

### Revert

To Revert a relay: `next relay revert [regex...]`

The tool will revert a relay back to the the previous version. It will remove the binary and associated files that are currently active and restore the most recent backup.

However it will not restore the previous public key in firestore as that is lost upon each update.

### Set NIC Speed

To set the NIC speed of a relay: `next relay setnic <relay name> <value (Mbps)>`

This will adjust the NIC speed field in Firestore for the given relay to the provided value in Mbps.

### Keys

To see the public and update keys of a relay: `next relay keys <relay name>`

Shows the public key and update key for the given relay.

### State

To set the state of a relay direct: `next relay state <state> <relay name> [relay names...]`

Sets the state of a relay directly in Firestore. Note that this command should be avoided and only used when the state logic is not working correctly.

Valid states are `enabled`, `offline`, `maintenance`, `disabled`, `quarantine`, and `decommissioned`

### Check

To check relays: `next relay check [filter]`

Displays a table of filtered relays, or all relays if no filter supplied, containing diagnostic information.

If any of the fields contain "SSH Error", a more detailed message will be visible within the messages above the table.

Fields:
- `SSH`: Was the relay able to be SSHed into?
- `Ubuntu`: The version of Ubuntu the relay is running.
- `Cores`: The cores allocated to the relay. If the `RELAY_MAX_CORES` environment variable is set, its value will used here.
- `Ping Backend`: Was the relay able to ping the relay backend?
- `Running`: Is the relay service running?
- `Bound`: Is port 40000 bound to the relay process?

Regarding SSH ability: The tool uses ssh's `-o` flag with `ConnectTimeout=60`. This means the tool will report false if the relay is unreachable for 60 seconds. If the relay is reachable and will take longer than 60 seconds, the tool will continue. So for very distant relays, and especially those with high packet loss, this tool function may take 2 or 3 minutes to complete.

## Cost

To download the cost matrix: `next cost [output file]`

Downloads the current cost matrix from the relay backend to 'cost.bin' or the first argument supplied.

## Optimize

To optimize a cost matrix: `next optimize [rtt threshold] [cost matrix file] [output file]`

Optimizes the downloaded cost matrix into a route matrix.

The first argument is the rtt threshold which defaults to 1. \
The second argument is the path of the cost matrix if you don't have it saved as 'cost.bin'. \
The third argument is the path you want the route matrix written to. Default is 'optimize.bin'.

## Analyze

To analyze a route matrix: `next analyze [route matrix file]`

Analyzes the route matrix and produces high level data. The file read in is 'optimize.bin' or the first argument supplied.

## Remaining Functionality

Remaining `next` commands are described in Slab [here](https://networknext.slab.com/posts/next-tool-updates-cuerpgzy).  
