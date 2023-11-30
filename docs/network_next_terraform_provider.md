<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Network Next Terraform provider

Network Next is configured via a Terraform provider. This provider talks with your REST API running in your environment and mutates and queries the contents of your Postgres SQL database for that environment.

After you make changes with terraforrm, always remember to commit your changes to the backend runtime:

```console
(terraform changes)
cd ~/next
next database
next commit
```

This downloads the postgres database contents into database.bin and then commits that database binary to your Network Next runtime. This way the backend runtime has no dependency on Postgres directly and can continue to work even if postgres is down.

You can read the Network Next Terrafrom docs here:

https://registry.terraform.io/providers/networknext/networknext/latest/docs

Return to [main documentation](../README.md)
