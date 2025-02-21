<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Spin up relays near you

Network Next takes advantage of relays near the client, as well as near the server.

For example, I live in New York, so to get the best acceleration to Sao Paulo, I need to also spin up as many relays near me as possible, from as many different providers that I can find.

I can run ```next datacenters newyork``` to see a list of datacenters where I can run relays in New York:

```console
gaffer@batman next % next datacenters newyork

┌───────────────────────┬──────────────────────────────────┬──────────┬───────────┐
│ Name                  │ Native                           │ Latitude │ Longitude │
├───────────────────────┼──────────────────────────────────┼──────────┼───────────┤
│ akamai.newyork        │ us-east                          │ 40.71    │ -74.01    │
│ amazon.philadelphia.1 │ use1-phl1-az1 (us-east-1-phl-1a) │ 39.95    │ -75.17    │
│ colocrossing.newyork  │                                  │ 40.7128  │ -74.006   │
│ datapacket.newyork    │                                  │ 40.7128  │ -74.006   │
│ equinix.newyork       │ NY                               │ 40.7128  │ -74.006   │
│ gcore.newyork         │                                  │ 40.7128  │ -74.006   │
│ hivelocity.newyork    │ NYC1                             │ 40.7128  │ -74.006   │
│ i3d.newark            │                                  │ 40.7357  │ -74.1724  │
│ latitude.newyork      │ NYC                              │ 40.7128  │ -74.006   │
└───────────────────────┴──────────────────────────────────┴──────────┴───────────┘
```

I also know that Virginia is a major hub on the way down the East Coast, so I could run relays here:

```console
gaffer@batman next % next datacenters virginia

┌────────────────────────────┬───────────────────────┬──────────┬───────────┐
│ Name                       │ Native                │ Latitude │ Longitude │
├────────────────────────────┼───────────────────────┼──────────┼───────────┤
│ amazon.virginia.1          │ use1-az1 (us-east-1c) │ 39.04    │ -77.49    │
│ amazon.virginia.2          │ use1-az2 (us-east-1d) │ 39.04    │ -77.49    │
│ amazon.virginia.3          │ use1-az3 (us-east-1e) │ 39.04    │ -77.49    │
│ amazon.virginia.4          │ use1-az4 (us-east-1a) │ 39.04    │ -77.49    │
│ amazon.virginia.5          │ use1-az5 (us-east-1f) │ 39.04    │ -77.49    │
│ amazon.virginia.6          │ use1-az6 (us-east-1b) │ 39.04    │ -77.49    │
│ datapacket.ashburn         │                       │ 39.0438  │ -77.4874  │
│ equinix.washingtondc       │ DC                    │ 38.9072  │ -77.0369  │
│ gcore.ashburn              │                       │ 39.0438  │ -77.4874  │
│ gcore.manassas             │                       │ 38.7509  │ -77.4753  │
│ google.virginia.1          │ us-east4-a            │ 37.43    │ -78.66    │
│ google.virginia.2          │ us-east4-b            │ 37.43    │ -78.66    │
│ google.virginia.3          │ us-east4-c            │ 37.43    │ -78.66    │
│ i3d.ashburn                │                       │ 39.0438  │ -77.4874  │
│ latitude.ashburn           │ ASH                   │ 39.0438  │ -77.4874  │
│ phoenixnap.ashburn         │                       │ 39.0438  │ -77.4874  │
│ serversdotcom.washingtondc │                       │ 38.9072  │ -77.0369  │
│ zenlayer.washingtondc      │                       │ 38.9072  │ -77.0369  │
└────────────────────────────┴───────────────────────┴──────────┴───────────┘
```

And I know that Miami is a key interchange between the United States and South America:

```
gaffer@batman next % next datacenters miami

┌──────────────────┬────────┬──────────┬───────────┐
│ Name             │ Native │ Latitude │ Longitude │
├──────────────────┼────────┼──────────┼───────────┤
│ akamai.miami     │ us-mia │ 25.76    │ -80.19    │
│ datapacket.miami │        │ 25.7617  │ -80.1918  │
│ equinix.miami    │ MI     │ 25.7617  │ -80.1918  │
│ gcore.miami      │        │ 25.7617  │ -80.1918  │
│ hivelocity.miami │ MIA1   │ 25.7617  │ -80.1918  │
│ latitude.miami   │ MIA    │ 25.7617  │ -80.1918  │
│ velia.miami      │        │ 25.7617  │ -80.1918  │
└──────────────────┴────────┴──────────┴───────────┘
```

As a second example, what instead of New York if you were in Hangzhou, Mainland China?

I would suggest that relays in Hong Kong and Singapore would be an excellent starting point:

```console
gaffer@batman next % next datacenters singapore hong kong

┌─────────────────────────┬─────────────────────────────┬──────────┬───────────┐
│ Name                    │ Native                      │ Latitude │ Longitude │
├─────────────────────────┼─────────────────────────────┼──────────┼───────────┤
│ akamai.singapore.1      │ ap-south                    │ 1.35     │ 103.82    │
│ akamai.singapore.2      │ sg-sin-2                    │ 1.35     │ 103.82    │
│ amazon.singapore.1      │ apse1-az1 (ap-southeast-1a) │ 1.35     │ 103.82    │
│ amazon.singapore.2      │ apse1-az2 (ap-southeast-1b) │ 1.35     │ 103.82    │
│ amazon.singapore.3      │ apse1-az3 (ap-southeast-1c) │ 1.35     │ 103.82    │
│ datapacket.singapore    │                             │ 1.3521   │ 103.8198  │
│ equinix.singapore       │ SG                          │ 1.3521   │ 103.8198  │
│ gcore.singapore         │                             │ 1.3521   │ 103.8198  │
│ google.singapore.1      │ asia-southeast1-a           │ 1.35     │ 103.82    │
│ google.singapore.2      │ asia-southeast1-b           │ 1.35     │ 103.82    │
│ google.singapore.3      │ asia-southeast1-c           │ 1.35     │ 103.82    │
│ i3d.hongkong            │                             │ 1.3521   │ 103.8198  │
│ i3d.singapore           │                             │ 1.3521   │ 103.8198  │
│ oneqode.hongkong        │                             │ 1.3521   │ 103.8198  │
│ oneqode.singapore       │                             │ 1.3521   │ 103.8198  │
│ phoenixnap.singapore    │                             │ 1.3521   │ 103.8198  │
│ serversdotcom.singapore │                             │ 1.3521   │ 103.8198  │
│ velia.singapore         │                             │ 1.3521   │ 103.8198  │
│ zenlayer.singapore      │                             │ 1.3521   │ 103.8198  │
└─────────────────────────┴─────────────────────────────┴──────────┴───────────┘
```

And perhaps on top of this, if you have access to local bare metal and cloud relays in Hangzhou, China that would also be an excellent place to run relays.

You'd probably also have to do some work adding your own seller under `terraform/sellers` so Network Next knows about new sellers and datacenters in Mainland China, but I'm happy to help you with this and it's pretty easy.

Up next: [Test acceleration to Sao Paulo](test_acceleration_to_sao_paolo.md).
