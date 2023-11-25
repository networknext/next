<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Native C++ SDK

## Introduction

The Network Next SDK integrates with your game's client and server and takes over sending and receiving UDP packets.

It supports the following platforms:

1. Windows
2. Mac
3. Linux
4. iOS
5. PS4
6. PS5
7. XBox One
8. XBox Series X
9. Nintendo Switch

The SDK lives under the "sdk" directory of the main repo, and as part of the `next keygen` and `next config` steps the source code has already been customized with your domain names and keypairs.

To get started, view the documentation for instructions on how to build the SDK, then study and run the example programs: [Network Next SDK Documentation](https://network-next-sdk.readthedocs-hosted.com/en/latest/index.html)

## Running SDK Examples

When you run the SDK examples, they will attempt to connect with your development backend instance. Please make sure you have the dev environment up when you run the SDK examples. 

To connect to your production backend, simply make sure that NEXT_DEVELOPMENT is not defined, and the code will automatically point at your production backend instance.

For security reasons, the buyer public and private keypair generated and embedded in your source code is for _internal testing purposes only_ and should be replaced with your own generated buyer account. For instructions on how to do this, see [How to Add a New Game](how_to_add_a_new_game.md).

## Datacenters and Autodetection

In order to accelerate to the correct datacenter, Network Next needs to know the name of the datacenter the server is hosted in. It does this by looking up relays in the same datacenter that your game server is in, and making sure that these relays are the last hop for accelerated traffic on the way to your server. The idea being that latency, jitter and packet loss is effectively zero for traffic sent from this last relay to your server.

For example, "google.iowa.1" is a default datacenter setup in the dev environment, and the dev environment has a relay set in this datacenter already. If you host your test server in that datacenter, you could simply set NEXT_DATACENTER="google.iowa.1" in the environment and Network Next will be able to accelerate traffic.

However, this can get cumbersome for multiple clouds, as well as being _literally impossible_ with a hosting provider like Multiplay who don't provide you with any way to know exactly which datacenter your server is being hosted in.

To get around this, we have implemented autodetection of your datacenter in popular clouds and for multiplay.

For Google Cloud and AWS, simply set your datacenter to "cloud" and the SDK will call internal REST APIs when the server starts up to determine which datacenter it is running in.

For Multiplay, you must ask them to pass in "cloud" when they are running your game server in Google Cloud or AWS, or to pass in the datacenter in the form "multiplay.[cityname]" and our code will use DNS running on your game server to work out which Multiplay provider is actually hosting your game server. This way we can ensure that traffic is accelerated to the correct bare metal provider.

## Configuration of the set of Google Cloud, AWS and Multiplay datacenters

Configuration that drives autodetect is defined in text files under "config".

When you run `next config` the code walks the set of AWS and Google Cloud datacenters available in your accounts, and updates the config/amazon.txt and config/google.txt respectively.

For multiplay, you need to manually determine the set of providers that multiplay is using (a starter set is included in config/multiplay.txt), this file acts as a mapping from the DNS results to the provider name, eg. "multiplay.[cityname]" turns into "inap.[cityname]" when we see "internap", "inap" or "INAP" in the DNS query results.

These files can be updated at any time and uploaded to Google Cloud Storage via semaphore CI, using the "Upload Config" job.

When the SDK runs, it downloads the config files on startup from Google Cloud Storage, so if you need to make changes to these files in production, they will become active to clients and servers running in production almost instantly. This way you you can adjust the set of datacenters available dynamically, without needing to deploy any code changes.
