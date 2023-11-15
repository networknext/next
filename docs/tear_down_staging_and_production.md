<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Tear down Staging and Production

## 1. Tear down Staging

Staging is meant to be used for load testing, and testing configuration changes before pushing them to production. In its default configuration, it is as expensive as a production backend that can handle 1M CCU.

Do not leave staging running. Bring it up only when you need it for testing, then tear it down.

To tear down staging:

```console
cd ~/next/terraform/staging/backend
terraform init
terraform destroy
```

You will be asked to enter a tag. Tag is not used when destroying the environment. Enter _any_ string and press ENTER. 

## 2. Tear down Production

Production is a relatively expensive environment to run. Unless you plan to take your game to production immediately, you should shut it down to save on cost.

First, tear down the production relays:

```console
cd ~/next/terraform/production/relays
terraform destroy
```

Then tear down the production backend:

```console
cd ~/next/terraform/production/backend
terraform init
terraform destroy
```

Do _NOT_ tear the backend down before you destroy the relays, otherwise you will be stuck manually deleting relay resources from google cloud, AWS and akamai accounts. I've been there, and it's very painful.

Congratulations. You have configured and setup your own instance of network next. You have successfully deployed dev, staging and production environments, and made sure that you have sufficient google cloud quota to support at least 1M CCU going through your production backend.

[Return to main documentation](../README.md)



