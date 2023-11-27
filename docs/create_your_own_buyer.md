<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Create your own buyer

## 1. Generate new buyer keypair

Go to the console and type:

```
cd ~/next && go run sdk/keygen/keygen.go
```

You will see output like this:

```
gaffer@macbook next % go run sdk/keygen/keygen.go

Welcome to Network Next!

This is your public key:

    SPeLMXdfJRtK3E2rEX7L9JIFxpn+cykxtuWAUCZVLbAEcwFrc0oVoQ==

This is your private key:

    SPeLMXdfJRt83tjKOYXbR0JyLdbuaGH7GpK21oTalLITqCOdBVzZ40rcTasRfsv0kgXGmf5zKTG25YBQJlUtsARzAWtzShWh

IMPORTANT: Save your private key in a secure place and don't share it with anybody, not even us!
```

This is your new buyer keypair. The public key can be safely shared with anybody and embedded in your client. The public key should be known only by the server, and not your players.

Save your keys in your secrets directory as `buyer_public_key.txt` and `buyer_private_key.txt` and then back up your secrets directory somewhere.

## 2. 


