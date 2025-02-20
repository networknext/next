<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Modify set of amazon relays

Unfortunately, the AWS terraform provider is structured in such a way that we need to generate much more terraform scripts than google cloud.

Because of this, the set of amazon relays in not specified in terraform, but in sellers/amazon.go:

> // DEV RELAYS
>
>var devRelayMap = map[string][]string{
>	"amazon.virginia.1":  {"amazon.virginia.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
>	"amazon.virginia.2":  {"amazon.virginia.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
>	"amazon.ohio.1":      {"amazon.ohio.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
>	"amazon.ohio.2":      {"amazon.ohio.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
> }

To change the set of amazon relays, edit amazon.go and change the devRelayMap to this:

```
// DEV RELAYS

var devRelayMap = map[string][]string{
	"amazon.virginia.1":  {"amazon.virginia.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.virginia.2":  {"amazon.virginia.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.virginia.3":  {"amazon.virginia.3", "m4.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.virginia.4":  {"amazon.virginia.4", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.virginia.5":  {"amazon.virginia.5", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.virginia.6":  {"amazon.virginia.6", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.saopaulo.1":  {"amazon.saopaulo.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.saopaulo.2":  {"amazon.saopaulo.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.saopaulo.3":  {"amazon.saopaulo.3", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
}
```

And then run:

```
next config
```

This updates the generated amazon terraform files.

Next, apply the changes with terraform:

```
cd ~/next/terraform/dev/relays
terraform init
terraform apply
```

Please note that depending on your AWS account, amazon.virginia.n will be mapped to different AWS availability zones in your account. See [this document](https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html) for more information.

Because of this, it's likely that you'll see an error that one of the virginia datacenters does not support m4.large or m5a.large machine types.

To fix this, you can view the mapping between network next datacenter name and availability zones on your account by going:

```
cd ~/next
next datacenters amazon | grep virginia
```

And it will print out the mapping of availability zones to network next datacenters for your AWS account:

<img width="918" alt="next datacenters amazon" src="https://github.com/user-attachments/assets/c7bb6b11-ee68-4bc7-b943-b056380bfc0a" />
	
To fix any terraform errors, you must use the aws command locally to find what instance types are supported in each AWS availability zone:

For example:

```
aws ec2 describe-instance-type-offerings --location-type availability-zone --filters Name=location,Values=us-east-1e
```

Will print out all instance types available in **us-east-1e**, which corresponds to _amazon.virginia.1_ on my AWS account (it will likely be different in your AWS account).0

Then you can edit sellers/amazon.go and setup supported instance types in each availability zone, and then run terraform again:

```
cd ~/next/terraform/dev/relays
terraform init
terraform apply
```

And once you have the correct machine type for each amazon datacenter it should succeed.

Commit the changes made to sellers/amazon.go to the git repo:

```
git commit -am "change amazon relays"
git push origin
```

Once you have the relays created, commit the database to make the new relay configuration active on the backend:

```
cd ~/next
next database
next commit
```

Wait a few minutes for the relays to initialize, then connect to the VPN and setup the new amazon relays:

```
next setup amazon
```

After a short while, the amazon relays should be online:

```
next relays amazon
```

<img width="1164" alt="next relays amazon" src="https://github.com/user-attachments/assets/f41be280-b757-44f1-a6d8-a3618355cab5" />

Up next: [Modify set of akamai relays](modify_set_of_akamai_relays.md).
