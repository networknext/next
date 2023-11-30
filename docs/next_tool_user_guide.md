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

## 

