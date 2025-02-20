<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Move test server to Sao Paulo

Edit the file "terraform/dev/backend/terraform.tfvars" and change the test_server_zone to "southamerica-east1-a":

<img width="1120" alt="change test server region and zone" src="https://github.com/user-attachments/assets/834fe79c-1ecf-4e4d-a4e3-a6dac093ba41" />
	
Commit the changes:

```
git commit -am "change test server zone to sao paulo"
git push origin
```

Tag a new dev release:

```
git tag dev-004
git push origin dev-004
```

When the deploy succeeds, click on the "Servers" item in the portal nav menu, and you should see the test server is now running in Sao Paulo:

<img width="1448" alt="test server in sao paulo portal" src="https://github.com/user-attachments/assets/dcc259ac-4d06-4120-b1c3-b126d7557033" />

Let's SSH into the test server and run "sudo journalctl -fu server" to view its log:

<img width="1499" alt="ssh into test server in sao paulo" src="https://github.com/user-attachments/assets/37e180e1-d657-4231-99bd-658f92055d2c" />

The test server log shows that the server is autodetecting that it's running in google.saopaulo.1, but it cannot find any server relays:

<img width="1076" alt="server found 0 server relays" src="https://github.com/user-attachments/assets/41223c50-ec97-4cc0-9187-fe04caf91759" />

Because of this, right now you can connect a test client to the Sao Paulo datacenter and it will connect fine:

```
run client
```

But the client is not accelerated.

Network Next requires two things before it can accelerate traffic to a server running in a datacenter:

1. The buyer must have acceleration enabled for the datacenter
2. There must be at least one relay in the same datacenter as the server

In the next steps, we're going to enable acceleration for Sao Paulo for the test buyer, and then we'll spin up some google cloud relays in Sao Paulo to act as server relays.

Up next: [Disable the raspberry clients](disable_the_raspberry_clients.md).
