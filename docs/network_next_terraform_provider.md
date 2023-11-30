<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Network Next Terraform provider

Network Next is configured via a Terraform provider. This provider talks with your REST API running in your environment to query and mutate the contents of the Postgres SQL database that configures Network Next.

You can see examples of the Network Next Terraform provider in action in the following files:

* `~/next/terraform/dev/relays/main.tf`
* `~/next/terraform/prod/relays/main.tf`

And you can read the Network Next Terrafrom provider docs here: https://registry.terraform.io/providers/networknext/networknext/latest/docs

After you make changes with terraform, always remember to _commit_ your changes to the backend runtime:

```console
(terraform changes)
cd ~/next
next database
next commit
```

This extracts the configuration data defined in Postgres and _commits_ it to the backend runtime. This way the essential runtime backend that plans and executes routing is able to operate independently from Postgres, and continues to function even if it is down.

[Back to main documentation](../README.md)
