<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run a local instance with Docker Compose

You can run a Network Next test environment locally with docker compose on any OS. This lets you quickly get a test instance of the whole system up and running locally, which you can interact with via a web browser.

1. Install Docker from http://docker.com

2. Change into the directory where you cloned the source

   `cd ~/next`

3. Build the system

   `docker compose build`

4. Bring the system up

   `docker compose up`

5. View the portal

Navigate to the network next portal at http://127.0.0.1:8080

It will take a few minutes for the system to fully initialize. Once everything has started up, you should see something like this:

<img width="1582" alt="image" src="https://github.com/networknext/next/assets/696656/0567f170-0beb-4e3b-bc33-6cffcc15d133">

Congratulations! Network Next up is up and running in docker!

6. Take the system down

   `docker compose down`

_You are now ready to [setup your local machine for development](setup_your_local_machine_for_development.md)._
