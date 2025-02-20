<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Spin dev back up

Tag a new dev build to trigger a deploy:

```
git checkout dev
git tag dev-002
git push origin dev-002
```

Wait for the deploy to complete on https://semaphoreci.com

Activate the google cloud dev configuration on your local machine:

```
gcloud config configurations activate dev
```

Wait for SSL certificates to become active:

```
gcloud compute ssl-certificates list
```

Select the dev environment and ping it:

```
next select dev
next ping
```

You should see a response:

```console
pong [dev-002]
```

Next, create dev relays with terraform:

```
cd ~/next/terraform/dev/relays
terraform init
terraform apply
```

Commit the changes terraform made to the database to make them active:

```
cd ~/next
next database
next commit
```

Connect to the VPN and setup the relays:

```
next setup
```

Disconnect from the VPN.

Wait 5-10 minutes and all the relays should be online:

```console
next relays

	gaffer@batman next % next relays

	┌───────────────────┬──────────────────────┬──────────────────┬────────┬────────┬──────────┬─────────────────────┐
	│ Name              │ PublicAddress        │ Id               │ Status │ Uptime │ Sessions │ Version             │
	├───────────────────┼──────────────────────┼──────────────────┼────────┼────────┼──────────┼─────────────────────┤
	│ akamai.atlanta    │ 66.228.56.126:40000  │ 4c1499bedb76d4c3 │ online │ 30m    │ 0        │ relay-release-1.0.0 │
	│ akamai.dallas     │ 45.56.124.213:40000  │ a93caa50aede83ce │ online │ 26m    │ 0        │ relay-release-1.0.0 │
	│ akamai.fremont    │ 74.207.254.36:40000  │ 93abc98ceb2e90f  │ online │ 27m    │ 0        │ relay-release-1.0.0 │
	│ akamai.newyork    │ 45.79.163.17:40000   │ 6cc0a603455bf226 │ online │ 30m    │ 0        │ relay-release-1.0.0 │
	│ amazon.ohio.1     │ 3.145.161.46:40000   │ 8202db0dab012b82 │ online │ 28m    │ 0        │ relay-release-1.0.0 │
	│ amazon.ohio.2     │ 18.219.60.100:40000  │ ae46ceb0b291cb1  │ online │ 31m    │ 0        │ relay-release-1.0.0 │
	│ amazon.virginia.1 │ 3.231.57.221:40000   │ 5e0e4e9688c34d3  │ online │ 30m    │ 0        │ relay-release-1.0.0 │
	│ amazon.virginia.2 │ 18.204.18.110:40000  │ f958ca961febf2ad │ online │ 31m    │ 0        │ relay-release-1.0.0 │
	│ google.iowa.1.a   │ 34.67.114.105:40000  │ 1e2e20dbe0b72873 │ online │ 26m    │ 0        │ relay-release-1.0.0 │
	│ google.iowa.1.b   │ 34.58.5.125:40000    │ 45dffc7b9af1a152 │ online │ 58s    │ 0        │ relay-release-1.0.0 │
	│ google.iowa.1.c   │ 146.148.94.99:40000  │ 2ff45e2957f7aae4 │ online │ 28m    │ 0        │ relay-release-1.0.0 │
	│ google.iowa.2     │ 130.211.207.80:40000 │ 505bec9a4a376968 │ online │ 27m    │ 0        │ relay-release-1.0.0 │
	│ google.iowa.3     │ 34.41.237.105:40000  │ bc5c83b15fb7ce5d │ online │ 28m    │ 0        │ relay-release-1.0.0 │
	│ google.iowa.6     │ 34.172.89.168:40000  │ c8a8fff602ba9372 │ online │ 28m    │ 0        │ relay-release-1.0.0 │
	│ google.ohio.1     │ 34.162.247.234:40000 │ cf1ee1f55d784043 │ online │ 29m    │ 0        │ relay-release-1.0.0 │
	│ google.ohio.2     │ 34.162.208.105:40000 │ ea918c4b7d07a1d3 │ online │ 28m    │ 0        │ relay-release-1.0.0 │
	│ google.ohio.3     │ 34.162.125.248:40000 │ cf96a8f48138ad41 │ online │ 27m    │ 0        │ relay-release-1.0.0 │
	│ google.virginia.1 │ 34.48.205.128:40000  │ 8a94407262f5dfe2 │ online │ 26m    │ 0        │ relay-release-1.0.0 │
	│ google.virginia.2 │ 35.245.14.224:40000  │ 3a460ae16945cfd9 │ online │ 27m    │ 0        │ relay-release-1.0.0 │
	│ google.virginia.3 │ 34.150.140.229:40000 │ 5928d45a42ab20c4 │ online │ 28m    │ 0        │ relay-release-1.0.0 │
	└───────────────────┴──────────────────────┴──────────────────┴────────┴────────┴──────────┴─────────────────────┘
```

View your dev portal running at **https://portal-dev.yourdomain.com**

You should see sessions running like this:

<img width="1422" alt="raspberry sessions" src="https://github.com/user-attachments/assets/43deea3c-62cd-441f-9d30-a064c16520c2" />

And see your relays are all online:

<img width="1422" alt="relays" src="https://github.com/user-attachments/assets/ed4d7dd0-ef64-462e-8595-78c9e07e9b38" />

Up next: [Run your own client and server](run_your_own_client_and_server.md).
