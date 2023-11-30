<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Next tool user guide

The next tool is primarily used to manage your relay fleet, but it is also used to configure Network Next and generate keypairs, and to select between local, dev, staging and production environments.

You can get a quick overview of commands available to you by just typing `next` at the root directory of the code repository:

```console
gaffer@macbook next % next

Network Next Operator Tool

USAGE
  next <subcommand>

SUBCOMMANDS
  keygen    Generate new keypairs for network next
  config    Configure network next
  secrets   Zip up secrets directory
  select    Select environment to use (local|dev|staging|prod)
  env       Display environment
  ping      Ping the REST API in the current environment
  database  Update local database.bin from the current environment Postgres DB and print it
  commit    Commit the local database.bin to the current environment runtime (server and relay backends)
  relays    List relays in the current environment
  ssh       SSH into the specified relay(s)
  logs      View the journalctl log for a relay
  setup     Setup the specified relay(s)
  start     Start the specified relay(s)
  stop      Stop the specified relay(s)
  load      Load the specific relay binary version onto one or more relays
  upgrade   Upgrade the specified relay(s)
  reboot    Reboot the specified relay(s)
  cost      Get cost matrix from current environment
  optimize  Optimize cost matrix into a route matrix
  analyze   Analyze route matrix from optimize
  routes    Print list of routes from one relay to another
```

## next keygen

This command generates new keypairs for Network Next. Please run this only once during setup. If you run it again and say "yes" to the prompt, it will overwrite your existing keypairs, and you will completely lose access to any Network Next instances you have created.

## next config

This command reads in config.json and updates files in the source code repository with your configuration. It also updates the configuration for various sellers like google, amazon and akamai, and updates the terraform files with the latest complete set of datacenters for each seller. 

_It's safe to call this whenever you want. It's not destructive like `next keygen`_.

## next secrets

Zips up your ~/secrets directory and writes it to ~/next/secrets.tar.gz

Call this during initial setup and anytime when your secrets change and you need to upload them to semaphore ci.

## next select

Selects the current Network Next environment.

Valid environments include:

* local
* dev
* staging
* prod

Example:

`next select prod`

## next env

Prints out information about the currently selected environment.

Example:

```console
gaffer@macbook next % next env

[dev]

 + API URL = https://api-dev.virtualgo.net
 + Portal API Key = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6ZmFsc2UsInBvcnRhbCI6dHJ1ZSwiaXNzIjoibmV4dCBrZXlnZW4iLCJpYXQiOjE3MDExNTc5NjN9.hnXL5hXnytNUBvGI_xkxiuPYmAzwiPYR_usBQ-xrm90
 + VPN Address = 45.79.157.168
 + SSH Key File = ~/secrets/next_ssh
 + Relay Backend URL = https://relay-dev.virtualgo.net
 + Relay Backend Public Key = bKjCNngZ1H+XJppN6MymZ9UoTgewgOsLeAMAOiiWuws=
 + Relay Artifacts Bucket Name = helsinki_network_next_relay_artifacts
 + Raspberry Backend URL = https://raspberry-dev.virtualgo.net
```

## next ping

Pings the currently selected environment and prints a pong response. Use this to verify that the REST API is up and running for the current environment, and at which tag.

Example:

```console
gaffer@macbook next % next ping

pong [dev-674]

```

## next database

Downloads the contents of the postgres database for the currently selected environment and saves it to database.bin

## next commit

Commits the local database.bin to the currently selected environment runtime.

Typical usage is:

```console
next database
next commit
```

... after any changes have been made to the postgres database via terraform.

## next relays

Displays the set of active relays in the current environment:

```console
gaffer@macbook next % next relays

┌──────────────────────┬───────────────────────┬──────────────────┬────────┬────────┬──────────┬─────────┐
│ Name                 │ PublicAddress         │ Id               │ Status │ Uptime │ Sessions │ Version │
├──────────────────────┼───────────────────────┼──────────────────┼────────┼────────┼──────────┼─────────┤
│ google.iowa.2        │ 34.27.134.240:40000   │ 4d1e4c14ead5399f │ online │ 2d     │ 87       │ 1.0.0   │
│ google.iowa.3        │ 34.67.226.145:40000   │ eeb892f0e5987cab │ online │ 2d     │ 77       │ 1.0.0   │
│ google.iowa.1        │ 34.70.248.202:40000   │ 415a4b04c11869d3 │ online │ 2d     │ 73       │ 1.0.0   │
│ amazon.virginia.2    │ 52.91.235.169:40000   │ 58fc285691a56bdc │ online │ 2d     │ 59       │ 1.0.0   │
│ amazon.virginia.1    │ 3.219.234.25:40000    │ cb63819eeb3c55b4 │ online │ 2d     │ 55       │ 1.0.0   │
│ google.virginia.1    │ 34.48.60.13:40000     │ 190eb911aab75ecc │ online │ 2d     │ 39       │ 1.0.0   │
│ google.virginia.2    │ 34.86.70.97:40000     │ 829e7cc3362f7fed │ online │ 2d     │ 38       │ 1.0.0   │
│ google.virginia.3    │ 34.86.226.87:40000    │ e6fdc3f2691ae06b │ online │ 2d     │ 26       │ 1.0.0   │
│ google.ohio.1        │ 34.162.208.240:40000  │ d4704315a3524a49 │ online │ 2d     │ 22       │ 1.0.0   │
│ google.iowa.6        │ 104.197.125.59:40000  │ 16dc95a70f52f5d  │ online │ 2d     │ 13       │ 1.0.0   │
│ amazon.ohio.1        │ 18.227.209.61:40000   │ 80edbdd686d730e9 │ online │ 2d     │ 11       │ 1.0.0   │
│ amazon.ohio.2        │ 3.12.111.230:40000    │ f5114d3661b34651 │ online │ 2d     │ 7        │ 1.0.0   │
│ akamai.newyork       │ 45.79.181.218:40000   │ be6dbe8727a1edf9 │ online │ 2d     │ 6        │ 1.0.0   │
│ google.ohio.2        │ 34.162.226.25:40000   │ 419a50f5f7e1c9a0 │ online │ 2d     │ 5        │ 1.0.0   │
│ google.ohio.3        │ 34.162.227.168:40000  │ 257d19a147ec3d1  │ online │ 2d     │ 4        │ 1.0.0   │
│ akamai.frankfurt     │ 172.105.66.56:40000   │ 1d9ec868ccbfa402 │ online │ 2d     │ 0        │ 1.0.0   │
│ akamai.london        │ 213.168.248.111:40000 │ 27e4be28dc29e16f │ online │ 2d     │ 0        │ 1.0.0   │
│ amazon.stockholm.1   │ 16.171.55.52:40000    │ 3c8715c652e415a4 │ online │ 2d     │ 0        │ 1.0.0   │
│ amazon.stockholm.2   │ 16.171.230.198:40000  │ 53f5cf9f09b92256 │ online │ 2d     │ 0        │ 1.0.0   │
│ amazon.stockholm.3   │ 16.16.107.28:40000    │ 928894bc0d909d9f │ online │ 2d     │ 0        │ 1.0.0   │
│ google.finland.1     │ 34.88.17.12:40000     │ 9c646938f1aab47d │ online │ 2d     │ 0        │ 1.0.0   │
│ google.finland.2     │ 35.228.232.136:40000  │ 44ca7d086dd5e8a8 │ online │ 2d     │ 0        │ 1.0.0   │
│ google.finland.3     │ 34.88.20.86:40000     │ f3447a6ba05317ba │ online │ 2d     │ 0        │ 1.0.0   │
│ google.frankfurt.1   │ 34.159.182.25:40000   │ da0d00aca0b7bc17 │ online │ 2d     │ 0        │ 1.0.0   │
│ google.frankfurt.2   │ 35.246.166.104:40000  │ 4da037d5c86df659 │ online │ 2d     │ 0        │ 1.0.0   │
│ google.frankfurt.3   │ 35.198.159.40:40000   │ f9965823aa737ac2 │ online │ 2d     │ 0        │ 1.0.0   │
│ google.london.1      │ 34.105.217.183:40000  │ 8bcbf1a45c56752  │ online │ 2d     │ 0        │ 1.0.0   │
│ google.london.2      │ 34.39.5.91:40000      │ bf2db1aeab1c21e9 │ online │ 2d     │ 0        │ 1.0.0   │
│ google.london.3      │ 35.246.125.202:40000  │ 5b686bad4600e94f │ online │ 2d     │ 0        │ 1.0.0   │
│ google.netherlands.1 │ 34.141.152.221:40000  │ 9a59e55a442d55   │ online │ 2d     │ 0        │ 1.0.0   │
│ google.netherlands.2 │ 35.234.166.43:40000   │ 5d35faefa1beffb0 │ online │ 2d     │ 0        │ 1.0.0   │
│ google.netherlands.3 │ 35.204.240.65:40000   │ a88bfe2e3ccae516 │ online │ 2d     │ 0        │ 1.0.0   │
└──────────────────────┴───────────────────────┴──────────────────┴────────┴────────┴──────────┴─────────┘

```
