<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Tear down Dev, Staging and Production

## 1. Tear down Dev

Dev is your test environment. You can easily bring it back up using terraform whenever you need it.

First, tear down the dev relays:

```console
cd ~/next/terraform/dev/relays
terraform destroy
```

You will be asked to enter a tag. The tag is not used when destroying the environment, so just input _any_ string and press ENTER.

Next, tear down the dev backend:

```console
cd ~/next/terraform/dev/backend
terraform init
terraform destroy
```

Do not be concerned if the terraform complains about not being able to destroy the postgres database. This database has delete protection enabled so terraform can't destroy it. This means the next time you bring dev up using terraform, you don't have to set the database up again, which saves time.

IMPORTANT: Do _NOT_ destroy the backend before destroying the relays, because destroying relays depends on the backend. If you do this in the wrong order you'll have to manually deleting relay resources from Google Cloud, AWS and Akamai and it's very painful.

## 2. Tear down Staging

Staging is meant to be used for load testing, and testing configuration changes before pushing them to production. In its default configuration, it is as expensive as a production backend that can handle 1M CCU so it's not something you want to leave running all the time.

To tear down staging:

```console
cd ~/next/terraform/staging/backend
terraform init
terraform destroy
```

## 3. Tear down Production

Production is an expensive environment to run, so you should shut it down when it's not being used.

First, tear down the production relays:

```console
cd ~/next/terraform/production/relays
terraform destroy
```

Next, tear down the production backend:

```console
cd ~/next/terraform/production/backend
terraform init
terraform destroy
```

Congratulations. You have shut down all environments so they are no longer costing you any money.

[Return to main documentation](../README.md)
