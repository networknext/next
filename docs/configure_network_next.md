<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Configure Network Next

## 1. Generate keypairs

Change to the "~/next" directory and run:

  `next keygen`

This generates a fresh set of keypairs for your network next instance, so that it uniquely secured vs. any other network next instances.

Secrets are generated and stored under ~/secrets.

Back up the secrets directory somewhere. If you lose it, you will not be able to administer your network next instance.

## 2. Edit config.json

Edit the config.json file at the root fo the next repository.

By default it should contain something like this:

```
{
  "company_name": "alocasia",
  "vpn_address": "45.79.157.168",
  "cloudflare_zone_id": "eba5d882ea2aa23f92dfb50fbf7e3cf4",
  "cloudflare_domain": "virtualgo.net",
  "google_billing_account": "012279-A33489-722F96",
  "google_org_id": "434699063105",
  "ssh_key": "~/secrets/next_rsa"
}
```

1. Set *company_name* to be some unique identifier. It could be your company name or a random word. It is not publicly visible, but it must be unique. It may contain only letters and underscores.

2. Set *vpn_address* to the IP address of the OpenVPN that you setup in the previous section
   
3. Set *cloudflare_zone_id* to the zone id for your domain managed by cloudflare. You can find the zone id in the cloudflare portal.

4. Set *cloudflare_domain* to the domain name you are using with network next. This domain must correspond to the zone id in cloudflare.

5. Set *google_billing_account* to one of your linked billing accounts in google cloud. Run `gcloud billing accounts list` to list the set of billing accounts linked to your google cloud account. There is usually only one.

6. Set *google_org_id* to your google organization id. Run `gcloud organizations list` to get a list of organization ids linked to your account. There is usually only one.

## 3. Run the configure script

Change to "~/next" and run:

`next config`

This config will modify source files throughout the repository.

Please review the changes with:

`git diff`

And commit these changes to the repository:

`git commit -am "config" && git push origin`

You are now ready to [Create Google Cloud Projects with Terraform](create_google_cloud_projects_with_terraform.md)
