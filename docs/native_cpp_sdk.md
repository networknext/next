<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Native C++ SDK

## Introduction

The Network Next SDK integrates with your game's client and server, and takes over sending and receiving UDP packets on behalf of your game. 

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

The SDK lives under the "sdk" directory of the main repo, and as part of the `next keygen` and `next config` steps the source code has been customized with your domain names and keypairs already.

When you run SDK examples, they will attempt to connect with your development backend instance. Please make sure you have the dev environment up when you run the SDK examples. 

To connect to your production backend, simply make sure that NEXT_DEVELOPMENT is not defined in your source code, and the code is already pointing at your production backend instance.

For security reasons, no private keys have been embedded in your source code, so you may need to ensure that NEXT_BUYER_PRIVATE_KEY is set on the command line ahead of time when you run the server examples. Only the server requires the private key, and the clients already have the public key embedded (this is safe). You can find your per-environment private keys under the ~/secrets directory.

To get started, view the documentation for instructions on how to build the SDK, then study and run the example programs: [Network Next SDK Documentation](https://network-next-sdk.readthedocs-hosted.com/en/latest/index.html)

## Datacenters and Autodetection

In order to accelerate to the correct datacenter, Network Next needs to know the name of the datacenter the server is hosted in.

For example, "google.iowa.1" is a default datacenter setup in the dev environment, and if you host your test server in that datacenter, you could simply set NEXT_DATACENTER="google.iowa.1" in the environment and Network Next will be able to accelerate traffic to your server.

However, this can get cumbersome for multiple clouds, as well as being _literally impossible_ with a hosting provider like Multiplay who don't provide you with any way to know exactly which datacenter your server is being hosted in.

To get around this, we have implemented autodetection of your datacenter in popular clouds and for multiplay.

For Google Cloud and AWS, simply set your datacenter to "cloud" and the SDK will call internal REST APIs on each cloud to determine which datacenter it is running in.

For Multiplay, you must ask them to pass in "cloud" when they are running your game server in Google Cloud or AWS, or to pass in the datacenter in the form "multiplay.[cityname]" and our code will use DNS running on your game server to work out which Multiplay provider is actually hosting your game server. This way we can ensure that traffic is accelerated to the correct bare metal provider.

## Configuration of the set of Google Cloud, AWS and Multiplay datacenters

Configuration that drives autodetect is defined in text files under "config".

Most of these files are automatically generated when you run `next config`. They can be uploaded to Google Cloud SDK via semaphore CI, using the "Upload Config" job.

When you run `next config` the code walks the set of AWS and Google Cloud datacenters available, and updates the config/amazon.txt and config/google.txt respectively.

For multiplay, you need to manually determine the set of providers that multiplay is using (a starter set is included in config/multiplay.txt), this file acts as a mapping from the DNS results to the provider name, eg. "multiplay.[cityname]" turns into "inap.[cityname]" when we see "internap", "inap" or "INAP" in the DNS query results.
