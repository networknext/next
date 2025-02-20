<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Modify route shader for test buyer

The test server belongs to a test buyer, and this test buyer has a "route shader" that describes when to accelerate clients.

Right now the route shader is set to always take network next - even if there is no acceleration found - for testing purposes.

Let's modify the route shader for the test buyer, so it only accelerates players if we can find at least 1 millisecond of latency reduction.

Open the file "terraform/dev/relays/main.tf" and make the following changes:

<img width="556" alt="force next false" src="https://github.com/user-attachments/assets/d76530fd-9505-4c92-9adc-d5d777634ed0" />

Commit the changes:

```
git commit -am "disable force next"
git push origin
```

Apply the changes with terraform:

```
cd ~/next/terraform/dev/relays
terraform init
terraform apply
```

The terraform actions have updated the route shader for the test buyer in the postgres database.

To make these changes live, we need to commit them to the dev environment:

```
cd ~/next
next database
next commit
```

This is the process for making any changes to the dev database configuration in terraform:

1. Modify terraform files
2. terraform init and apply
3. Commit the database changes

Please wait a few minutes for the updated settings to take effect, they should be live in less than a minute.

Connect a test client again:

```
run client
```

You will see that now it probably won't be accelerated:

<img width="1414" alt="session not accelerated" src="https://github.com/user-attachments/assets/500106f0-c5d4-4d4f-bcb6-060805e91ff3" />

Drilling in to the session you can now see only the non-accelerated latency, so the session is not accelerated.

<img width="1413" alt="session detail not accelerated" src="https://github.com/user-attachments/assets/34a08352-6c05-4d3d-8f38-73f7eb6412bf" />

Obviously, it's not particularly impressive to see a session _not_ be accelerated. We want to see some acceleration!

To make it easier to see _real acceleration_, let's move the test server to Sao Paulo, Brasil.

Up next: [Move test server to Sao Paulo](move_test_server_to_sao_paulo.md).
