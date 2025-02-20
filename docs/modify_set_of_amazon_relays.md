<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Modify set of amazon relays

Unfortunately, the AWS terraform provider is structured in such a way that we need to generate much more terraform files to get it to work with Network Next relays.

Because of this, the set of amazon relays in not specified in terraform, but in sellers/amazon.go:

```terraform
// DEV RELAYS

var devRelayMap = map[string][]string{
	"amazon.virginia.1":  {"amazon.virginia.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.virginia.2":  {"amazon.virginia.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.ohio.1":      {"amazon.ohio.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.ohio.2":      {"amazon.ohio.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
}

```

To change the set of amazon relays, just edit this this file and change the devRelayMap:

```terraform
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

This updates the generated amazon terraform files from the definitions in "sellers/amazon.go".

Next, apply the changes with terraform:

```
cd ~/next/terraform/dev/relays
terraform init
terraform apply
```

Please note that depending on your AWS account, amazon.virginia.n will be mapped to different AWS availability zones in your account. This is not something that Network Next does, but it is inherent to how AWS  works. Please see [this document](https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html) for more information.

Because of this, it's likely that you'll see an error that one of the virginia datacenters does not support m4.large or m5a.large machine types.

To fix this, you can view the mapping between Network Next datacenter name and availability zones on your account by going:

```
cd ~/next
next datacenters amazon | grep virginia
```

And it will print out the mapping of availability zones to network next datacenters for your AWS account in virginia:

<img width="910" alt="next datacenters amazon virginia" src="https://github.com/user-attachments/assets/3ae241fc-24b9-4f9f-a86f-235fc8b6549b" />

To fix any terraform errors, you must use the `aws` command locally to find what instance types are supported in each AWS availability zone:

For example:

```
aws ec2 describe-instance-type-offerings --location-type availability-zone --filters Name=location,Values=us-east-1e
```

Will print out all instance types available in **us-east-1e**, which corresponds to _amazon.virginia.1_ on my AWS account (it will likely be different in your AWS account).0

Then you can edit sellers/amazon.go and setup supported instance types in each availability zone, and then run next config and terraform again:

```
cd ~/next
next config
cd terraform/dev/relays
terraform init
terraform apply
```

Once you have the correct machine type for each amazon datacenter it should succeed.

Commit the changes made to sellers/amazon.go to the git repo:

```
git commit -am "change amazon relays"
git push origin
```

And commit the database to make the new relay configuration active on the backend:

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

<img width="1158" alt="next relays amazon" src="https://github.com/user-attachments/assets/401eb119-a6f9-4f4b-acc1-112c7df332e3" />

Up next: [Modify set of akamai relays](modify_set_of_akamai_relays.md).
